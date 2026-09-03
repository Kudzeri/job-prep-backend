package middleware

import (
	"strings"

	"github.com/Kudzeri/job-prep-backend/pkg/jwt"
	"github.com/gofiber/fiber/v3"
)

const (
	LocalUserIDKey = "userID"
	LocalRoleKey   = "userRole"
)

// Protected проверяет наличие и валидность Access-токена
func Protected(jwtSecret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "отсутствует заголовок Authorization",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "неверный формат токена (ожидается Bearer <token>)",
			})
		}

		tokenString := parts[1]
		claims, err := jwt.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "недействительный или истекший токен",
			})
		}

		// В Fiber v3 c.Locals принимает любой ключ и значение
		c.Locals(LocalUserIDKey, claims.UserID)
		c.Locals(LocalRoleKey, claims.Role)

		return c.Next()
	}
}