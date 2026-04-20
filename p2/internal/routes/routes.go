package routes

import (
	"github.com/Tanveer-rajpurohit/p2/internal/auth"
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(router *chi.Mux, queries *db.Queries, rdb *redis.Client) {
	router.Get("/health", handlers.HandlerReadiness)

	authHandler := &handlers.AuthHandler{Q: queries}

	router.Post("/auth/register", authHandler.Register)
	router.Post("/auth/login", authHandler.Login)
	router.Post("/auth/refresh", authHandler.RefreshToken)

	userHandler := &handlers.UserHandler{
		Q:     queries,
		Redis: rdb,
	}

	router.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)

		r.Get("/user", userHandler.GetUser)
		// r.Put("/user/avatar",userHandler.UpdateUserAvatar)
		r.Patch("/user/{user_id}", userHandler.UpdateUser)
	})

}