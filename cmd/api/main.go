package main

import (
	"log"
	"os"

	"github.com/Kudzeri/job-prep-backend/internal/handler"
	"github.com/Kudzeri/job-prep-backend/internal/repository"
	"github.com/Kudzeri/job-prep-backend/internal/service"
	"github.com/Kudzeri/job-prep-backend/pkg/database"
	"github.com/Kudzeri/job-prep-backend/pkg/email"
	"github.com/Kudzeri/job-prep-backend/pkg/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, берутся переменные окружения ОС")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-key"
	}

	// Считываем конфиг SMTP
	emailSender := email.NewSender(email.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	})

	// 2. Подключаем БД
	dbPool := database.NewPostgresPool(os.Getenv("DB_URL"))
	defer dbPool.Close()

	// 3. Инициализируем Fiber v3
	app := fiber.New(fiber.Config{
		AppName: "Job Prep Backend v1.0",
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(cors.New())

	// Публичный роут проверки здоровья
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running!",
		})
	})

	// Инициализируем слои
	authRepo := repository.NewAuthRepository(dbPool)
	authService := service.NewAuthService(authRepo, emailSender, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	// Эндпоинты авторизации
	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/email/send", authHandler.SendOTP)
	authGroup.Post("/email/verify", authHandler.VerifyOTP)

	// Защищенная группа роутов через middleware.Protected
	api := app.Group("/api/v1", middleware.Protected(jwtSecret))

	api.Get("/profile", func(c fiber.Ctx) error {
		userID := c.Locals(middleware.LocalUserIDKey).(int64)
		role := c.Locals(middleware.LocalRoleKey).(string)

		return c.JSON(fiber.Map{
			"message": "Доступ разрешен!",
			"user_id": userID,
			"role":    role,
		})
	})

	// Запуск сервера
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(app.Listen(":" + port))
}
