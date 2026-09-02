package main

import (
	"log"
	"os"

	"github.com/Kudzeri/job-prep-backend/pkg/database"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3"
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

	// 2. Подключаем БД
	dbPool := database.NewPostgresPool(os.Getenv("DB_URL"))
	defer dbPool.Close()

	// 3. Инициализируем Fiber
	app := fiber.New(fiber.Config{
		AppName: "Job Prep Backend v1.0",
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(cors.New())

	// Базовый роут проверки здоровья
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running!",
		})
	})

	// Запуск сервера
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(app.Listen(":" + port))
}
