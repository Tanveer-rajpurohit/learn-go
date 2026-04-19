package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Tanveer-rajpurohit/start/internal/config"
	"github.com/Tanveer-rajpurohit/start/internal/db"
	"github.com/Tanveer-rajpurohit/start/internal/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hello, World!")
	godotenv.Load() // Load .env file
	PORT := os.Getenv("PORT")
	if PORT == "" {
		log.Fatal("PORT is not found in env")
	}

	pool := config.ConnectDB()
	defer pool.Close()

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
	routes.SetupRouter(v1router, db.New(pool))
	router.Mount("/api/v1", v1router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + PORT,
	}

	log.Printf("Server starting on port %v", PORT)
	// err := http.ListenAndServe(":8080", router)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Failed to start server")
	}
}
