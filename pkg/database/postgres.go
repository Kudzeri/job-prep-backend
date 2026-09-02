package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(dbURL string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), dbURL)

	if err != nil {
		log.Fatal("Не удалось подключиться к БД: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}

	log.Println("Успешное подключение к PostgreSQL!")
	return pool
}
