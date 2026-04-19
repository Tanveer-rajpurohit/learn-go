package routes

import (
	"github.com/Tanveer-rajpurohit/p2/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(router *chi.Mux) {
	router.Get("/health", handlers.HandlerReadiness)
}