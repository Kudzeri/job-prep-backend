package main

import (
	"log"
	"os"
	"strings"

	"github.com/Kudzeri/job-prep-backend/internal/handler"
	"github.com/Kudzeri/job-prep-backend/internal/repository"
	"github.com/Kudzeri/job-prep-backend/internal/service"
	"github.com/Kudzeri/job-prep-backend/pkg/database"
	"github.com/Kudzeri/job-prep-backend/pkg/email"
	"github.com/Kudzeri/job-prep-backend/pkg/middleware"
	"github.com/Kudzeri/job-prep-backend/pkg/swagger"
	"github.com/Kudzeri/job-prep-backend/pkg/telegram"
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
	authRepo := repository.NewAuthRepository(dbPool)

	// 3. Инициализируем Fiber v3
	app := fiber.New(fiber.Config{
		AppName: "Job Prep Backend v1.0",
	})

	// Запуск Telegram-бота в фоне
	tgBot, err := telegram.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"), authRepo)
	if err != nil {
		log.Printf("Предупреждение: Telegram бот не запущен: %v", err)
	} else {
		go tgBot.Start()
		log.Println("Telegram бот успешно запущен")
	}

	// Middlewares
	app.Use(logger.New())
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000,http://127.0.0.1:3000,http://localhost:8081"
	}
	origins := make([]string, 0)
	for _, origin := range strings.Split(corsOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	isLoopbackOrigin := func(origin string) bool {
		return origin == "null" ||
			origin == "http://localhost" ||
			strings.HasPrefix(origin, "http://localhost:") ||
			origin == "https://localhost" ||
			strings.HasPrefix(origin, "https://localhost:") ||
			origin == "http://127.0.0.1" ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			origin == "https://127.0.0.1" ||
			strings.HasPrefix(origin, "https://127.0.0.1:") ||
			origin == "http://[::1]" ||
			strings.HasPrefix(origin, "http://[::1]:") ||
			origin == "https://[::1]" ||
			strings.HasPrefix(origin, "https://[::1]:")
	}

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			if isLoopbackOrigin(origin) {
				return true
			}

			for _, allowedOrigin := range origins {
				if origin == allowedOrigin {
					return true
				}
			}

			return false
		},
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "User-Agent"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))
	swagger.Register(app)

	// Публичный роут проверки здоровья
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running!",
		})
	})

	// Инициализируем слои
	authService := service.NewAuthService(authRepo, emailSender, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)
	userRepo := repository.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Эндпоинты авторизации
	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/email/send", authHandler.SendOTP)
	authGroup.Post("/email/verify", authHandler.VerifyOTP)
	authGroup.Post("/telegram/init", authHandler.InitTelegramAuth)
	authGroup.Get("/telegram/check", authHandler.CheckTelegramAuth)

	// Защищенные роуты
	protectedUsers := app.Group("/api/v1/users", middleware.Protected(jwtSecret))
	protectedUsers.Get("/me", userHandler.GetMe)
	protectedUsers.Post("/me/onboarding", userHandler.CompleteOnboarding)
	protectedUsers.Patch("/me", userHandler.UpdateProfile)

	app.Get("/api/v1/profile", middleware.Protected(jwtSecret), func(c fiber.Ctx) error {
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
