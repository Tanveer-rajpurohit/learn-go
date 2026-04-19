package routes

import (
	"github.com/Tanveer-rajpurohit/start/internal/auth"
	"github.com/Tanveer-rajpurohit/start/internal/db"
	"github.com/Tanveer-rajpurohit/start/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(router *chi.Mux, queries *db.Queries) {
	router.Get("/health", handlers.HandlerReadiness)

	ah := &handlers.AuthHandler{Q: queries}
	router.Post("/auth/register", ah.Register)
	router.Post("/auth/login", ah.Login)
	router.Post("/auth/refresh", ah.RefreshToken)

	h := &handlers.UserHandler{Q: queries}

	router.Route("/users", func(r chi.Router) {
		 r.Use(auth.RequireAuth)
		r.Get("/", h.GetUsers)
		r.Get("/{id}", h.GetUserByID)
		r.Post("/", h.CreateUser)
		r.Put("/{id}", h.UpdateUser)
		r.Delete("/{id}", h.DeleteUser)
	})
}
