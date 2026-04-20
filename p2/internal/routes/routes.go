package routes

import (
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(router *chi.Mux, queries *db.Queries) {
	router.Get("/health", handlers.HandlerReadiness)

	authHandler := &handlers.AuthHandler{Q: queries}

	router.Post("/auth/register", authHandler.Register)
	router.Post("/auth/login", authHandler.Login)
	router.Post("/auth/refresh", authHandler.RefreshToken)
}