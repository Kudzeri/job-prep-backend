package handler

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/Kudzeri/job-prep-backend/internal/repository"
	"github.com/Kudzeri/job-prep-backend/pkg/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	maxAssistantFileSize  = 10 << 20
	maxAssistantTotalSize = 24 << 20
)

type AssistantHandler struct {
	requests       *repository.AssistantRepository
	serviceURL     string
	serviceAPIKey  string
	callbackSecret string
	publicAPIURL   string
	client         *http.Client
}

func NewAssistantHandler(requests *repository.AssistantRepository, serviceURL, serviceAPIKey, callbackSecret, publicAPIURL string) *AssistantHandler {
	return &AssistantHandler{requests: requests, serviceURL: strings.TrimRight(strings.TrimSpace(serviceURL), "/"), serviceAPIKey: serviceAPIKey, callbackSecret: callbackSecret, publicAPIURL: strings.TrimRight(strings.TrimSpace(publicAPIURL), "/"), client: &http.Client{Timeout: 30 * time.Second}}
}

func (h *AssistantHandler) Chat(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.LocalUserIDKey).(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}
	if h.serviceURL == "" || h.callbackSecret == "" || h.publicAPIURL == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "генерация тестов пока не подключена на сервере"})
	}
	if !isHTTPURL(h.serviceURL) || !isHTTPURL(h.publicAPIURL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "неверно настроен адрес сервиса генерации"})
	}
	message := strings.TrimSpace(c.FormValue("message"))
	if message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "сообщение message обязательно"})
	}
	resume, err := c.FormFile("resume")
	if err != nil || resume == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "прикрепи резюме в формате PDF"})
	}
	attachments := []*multipart.FileHeader{}
	if form, formErr := c.MultipartForm(); formErr == nil && form != nil {
		attachments = form.File["attachments"]
	}
	if len(attachments) > 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "можно добавить не более двух дополнительных файлов"})
	}
	if strings.ToLower(filepath.Ext(resume.Filename)) != ".pdf" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "резюме должно быть PDF-файлом"})
	}
	files := append([]*multipart.FileHeader{resume}, attachments...)
	var total int64
	for _, file := range files {
		if file.Size <= 0 || file.Size > maxAssistantFileSize {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "каждый файл должен быть не пустым и не больше 10 МБ"})
		}
		if !allowedAssistantFile(file.Filename) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "поддерживаются PDF, DOCX, TXT и MD файлы"})
		}
		total += file.Size
	}
	if total > maxAssistantTotalSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "суммарный размер файлов не должен превышать 24 МБ"})
	}
	requestID := uuid.NewString()
	if err := h.requests.Create(c, requestID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось создать запрос на генерацию"})
	}
	body, contentType, err := buildAssessmentMultipart(requestID, message, h.publicAPIURL+"/api/v1/integrations/assistant/callback", resume, attachments)
	if err != nil {
		_ = h.requests.Fail(c, requestID, "не удалось прочитать файлы")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "не удалось прочитать загруженные файлы"})
	}
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, h.serviceURL+"/requests", body)
	if err != nil {
		_ = h.requests.Fail(c, requestID, "не удалось подготовить запрос")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось подготовить запрос к сервису генерации"})
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Callback-URL", h.publicAPIURL+"/api/v1/integrations/assistant/callback")
	req.Header.Set("X-Assessment-Callback-Secret", h.callbackSecret)
	if h.serviceAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.serviceAPIKey)
	}
	response, err := h.client.Do(req)
	if err != nil {
		_ = h.requests.Fail(c, requestID, "сервис генерации недоступен")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "сервис генерации недоступен"})
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		_ = h.requests.Fail(c, requestID, "сервис генерации отклонил запрос")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "сервис генерации не принял запрос", "upstream_status": response.StatusCode})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"request_id": requestID, "status": "pending", "message": "Запрос принят. Результат появится здесь после обработки документов."})
}

func isHTTPURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http")
}

func allowedAssistantFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf", ".docx", ".txt", ".md":
		return true
	default:
		return false
	}
}

func buildAssessmentMultipart(requestID, message, callbackURL string, resume *multipart.FileHeader, attachments []*multipart.FileHeader) (*bytes.Buffer, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{"request_id": requestID, "message": message, "callback_url": callbackURL}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}
	files := append([]*multipart.FileHeader{resume}, attachments...)
	for i, fileHeader := range files {
		field := "attachments"
		if i == 0 {
			field = "resume"
		}
		part, err := writer.CreateFormFile(field, filepath.Base(fileHeader.Filename))
		if err != nil {
			return nil, "", err
		}
		file, err := fileHeader.Open()
		if err != nil {
			return nil, "", err
		}
		_, copyErr := io.Copy(part, io.LimitReader(file, maxAssistantFileSize+1))
		closeErr := file.Close()
		if copyErr != nil {
			return nil, "", copyErr
		}
		if closeErr != nil {
			return nil, "", closeErr
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return &body, writer.FormDataContentType(), nil
}

func (h *AssistantHandler) GetRequest(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.LocalUserIDKey).(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}
	requestID := c.Params("requestId")
	if _, err := uuid.Parse(requestID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный requestId"})
	}
	item, err := h.requests.Get(c, requestID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "запрос не найден"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось получить статус запроса"})
	}
	return c.JSON(item)
}

func (h *AssistantHandler) Callback(c fiber.Ctx) error {
	if h.callbackSecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "callback не настроен"})
	}
	provided := c.Get("X-Assessment-Callback-Secret")
	if len(provided) != len(h.callbackSecret) || subtle.ConstantTimeCompare([]byte(provided), []byte(h.callbackSecret)) != 1 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неверный callback secret"})
	}
	var payload domain.AssessmentCallback
	if err := c.Bind().JSON(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректное JSON-тело callback"})
	}
	if _, err := uuid.Parse(payload.RequestID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный request_id"})
	}
	if payload.Status == "failed" {
		msg := strings.TrimSpace(payload.Result.Message)
		if msg == "" {
			msg = "не удалось обработать документы"
		}
		if err := h.requests.Fail(c, payload.RequestID, msg); err != nil {
			return callbackStateError(c, err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
	switch payload.Result.Type {
	case "message":
		if strings.TrimSpace(payload.Result.Message) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "result.message обязателен"})
		}
		if err := h.requests.CompleteMessage(c, payload.RequestID, strings.TrimSpace(payload.Result.Message)); err != nil {
			return callbackStateError(c, err)
		}
	case "test":
		test := domain.Test{Title: strings.TrimSpace(payload.Result.Title), Description: strings.TrimSpace(payload.Result.Description)}
		questions := payload.Result.Questions
		if test.Title == "" || len(questions) == 0 || len(questions) > 500 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "для результата test нужны title и от 1 до 500 questions"})
		}
		for _, q := range questions {
			if !validQuestion(q) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "каждый вопрос должен содержать text и корректные options/answer"})
			}
		}
		testID, err := h.requests.CompleteTest(c, payload.RequestID, test, questions)
		if err != nil {
			return callbackStateError(c, err)
		}
		return c.JSON(fiber.Map{"status": "completed", "test_id": testID})
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "result.type должен быть message или test"})
	}
	return c.JSON(fiber.Map{"status": "completed"})
}

func callbackStateError(c fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrAssistantRequestFinal) || errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "запрос уже обработан или не существует"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось сохранить результат"})
}
