package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/Kudzeri/job-prep-backend/internal/repository"
	"github.com/Kudzeri/job-prep-backend/pkg/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

type UserTestHandler struct{ repo *repository.TestRepository }

func NewUserTestHandler(repo *repository.TestRepository) *UserTestHandler {
	return &UserTestHandler{repo: repo}
}

func (h *UserTestHandler) userID(c fiber.Ctx) (int64, bool) {
	id, ok := c.Locals(middleware.LocalUserIDKey).(int64)
	return id, ok
}

func (h *UserTestHandler) List(c fiber.Ctx) error {
	userID, ok := h.userID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}
	items, err := h.repo.ListForUser(c, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось загрузить тесты"})
	}
	return c.JSON(fiber.Map{"tests": items})
}

func (h *UserTestHandler) Create(c fiber.Ctx) error {
	userID, ok := h.userID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}
	var req domain.CreateTestRequest
	if err := c.Bind().JSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title обязателен"})
	}
	test, err := h.repo.CreateForUser(c, userID, strings.TrimSpace(req.Title), strings.TrimSpace(req.Description))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось создать тест"})
	}
	return c.Status(fiber.StatusCreated).JSON(test)
}

func (h *UserTestHandler) Get(c fiber.Ctx) error {
	userID, ok := h.userID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}
	id, err := strconv.ParseInt(c.Params("testId"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный testId"})
	}
	details, err := h.repo.GetForUser(c, id, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "тест не найден"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось загрузить тест"})
	}
	return c.JSON(details)
}

func (h *UserTestHandler) AddQuestion(c fiber.Ctx) error {
	userID, ok := h.userID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}
	id, err := strconv.ParseInt(c.Params("testId"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный testId"})
	}
	var q domain.Question
	if err := c.Bind().JSON(&q); err != nil || !validQuestion(q) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "text обязателен, options и answer должны быть корректным JSON"})
	}
	result, err := h.repo.AddQuestionForUser(c, id, userID, q)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "тест не найден"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось добавить вопрос"})
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *UserTestHandler) AddQuestions(c fiber.Ctx) error {
	userID, ok := h.userID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}
	id, err := strconv.ParseInt(c.Params("testId"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный testId"})
	}
	var req struct {
		Questions []domain.Question `json:"questions"`
	}
	if err := c.Bind().JSON(&req); err != nil || len(req.Questions) == 0 || len(req.Questions) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "передайте от 1 до 500 вопросов в questions"})
	}
	for _, q := range req.Questions {
		if !validQuestion(q) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "у каждого вопроса обязателен text, options и answer должны быть корректным JSON"})
		}
	}
	results, err := h.repo.AddQuestionsForUser(c, id, userID, req.Questions)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "тест не найден"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось добавить вопросы"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"questions": results})
}
