package handler

import (
	"github.com/Kudzeri/job-prep-backend/internal/service"
	"github.com/Kudzeri/job-prep-backend/pkg/middleware"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GET /api/v1/users/me — получение профиля
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.LocalUserIDKey).(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}

	user, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "пользователь не найден"})
	}

	return c.JSON(user)
}

// POST /api/v1/users/me/onboarding — завершение онбординга
func (h *UserHandler) CompleteOnboarding(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.LocalUserIDKey).(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}

	var req struct {
		FirstName string `json:"first_name"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный формат данных"})
	}

	user, err := h.userService.CompleteProfile(c.Context(), userID, req.FirstName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(user)
}

// PATCH /api/v1/users/me — редактирование профиля
func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.LocalUserIDKey).(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неавторизован"})
	}

	var req service.UpdateProfileInput
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "некорректный формат данных"})
	}

	user, err := h.userService.UpdateProfile(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(user)
}