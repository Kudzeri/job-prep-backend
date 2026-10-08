package handler

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/Kudzeri/job-prep-backend/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

type IntegrationTestHandler struct{ repo *repository.TestRepository }

func NewIntegrationTestHandler(repo *repository.TestRepository) *IntegrationTestHandler {
	return &IntegrationTestHandler{repo: repo}
}

func (h *IntegrationTestHandler) CreateTest(c fiber.Ctx) error {
	var req domain.CreateTestRequest
	if err := c.Bind().JSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title обязателен"})
	}
	test, err := h.repo.Create(c, strings.TrimSpace(req.Title), strings.TrimSpace(req.Description))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось создать тест"})
	}
	return c.Status(fiber.StatusCreated).JSON(test)
}

func (h *IntegrationTestHandler) AddQuestion(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("testId"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный testId"})
	}
	exists, err := h.repo.Exists(c, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось проверить тест"})
	}
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "тест не найден"})
	}
	var q domain.Question
	if err := c.Bind().JSON(&q); err != nil || !validQuestion(q) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "text обязателен, options и answer должны быть корректным JSON"})
	}
	result, err := h.repo.AddQuestion(c, id, q)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "тест не найден"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось добавить вопрос"})
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *IntegrationTestHandler) AddQuestions(c fiber.Ctx) error {
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
	exists, err := h.repo.Exists(c, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось проверить тест"})
	}
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "тест не найден"})
	}
	results, err := h.repo.AddQuestions(c, id, req.Questions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "не удалось добавить вопросы"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"questions": results})
}

func validQuestion(q domain.Question) bool {
	if strings.TrimSpace(q.Text) == "" {
		return false
	}
	for _, raw := range [][]byte{q.Options, q.Answer} {
		if len(raw) > 0 && !json.Valid(raw) {
			return false
		}
	}
	return true
}
