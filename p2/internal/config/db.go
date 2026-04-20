package config

import (
    "context"
    "log"
    "os"
    "github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() *pgxpool.Pool {
    url := os.Getenv("DATABASE_URL")
    if url == "" {
        log.Fatal("DATABASE_URL not set")
    }
    pool, err := pgxpool.New(context.Background(), url)
    if err != nil {
        log.Fatal("Failed to connect to DB: ", err)
    }
    return pool
}