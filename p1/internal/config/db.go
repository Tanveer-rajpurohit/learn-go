package config

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() *pgxpool.Pool {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not found in env")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	return pool
}