package handler

import (
	"github.com/Kudzeri/job-prep-backend/internal/service"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type SendOTPRequest struct {
	Email string `json:"email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// POST /api/v1/auth/email/send
func (h *AuthHandler) SendOTP(c fiber.Ctx) error {
	var req SendOTPRequest
	if err := c.Bind().Body(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "укажите корректный email",
		})
	}

	if err := h.authService.SendOTP(c.Context(), req.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось отправить OTP код",
		})
	}

	return c.JSON(fiber.Map{
		"message": "OTP код отправлен на ваш email",
	})
}

// POST /api/v1/auth/email/verify
func (h *AuthHandler) VerifyOTP(c fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil || req.Email == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email и code обязательны",
		})
	}

	userAgent := c.Get("User-Agent")
	ip := c.IP()

	tokens, err := h.authService.VerifyOTP(c.Context(), req.Email, req.Code, userAgent, ip)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(tokens)
}