package middleware

import (
	"crypto/subtle"
	"github.com/gofiber/fiber/v3"
)

func APIKey(expected string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if expected == "" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "внешний API не настроен"})
		}
		provided := c.Get("X-API-Key")
		if len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неверный API-ключ"})
		}
		return c.Next()
	}
}
