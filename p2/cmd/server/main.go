package main

import (
	"log"
	"net/http"
	"os"

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
	routes.SetupRouter(v1router)
	router.Mount("/api/v1", v1router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + PORT,
	}

	log.Printf("Server starting on port %v", PORT)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Failed to start server : ", err)
	}
}