package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Tanveer-rajpurohit/p2/internal/config"
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	pool := config.ConnectDB()
	// make sure db is connected before starting the server
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	rdb := config.ConnectRedis()
	// make sure redis is connected before starting the server
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer rdb.Close()

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1router := chi.NewRouter()
	routes.SetupRouter(v1router, db.New(pool), rdb)
	router.Mount("/api/v1", v1router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + PORT,
	}

	log.Printf("Server starting on port %v", PORT)

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server : ", err)
	}

	log.Printf("Server stopped gracefully")
}
