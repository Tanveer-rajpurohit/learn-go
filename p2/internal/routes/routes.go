package routes

import (
	"time"

	"github.com/Tanveer-rajpurohit/p2/internal/auth"
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/handlers"
	"github.com/Tanveer-rajpurohit/p2/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/Tanveer-rajpurohit/p2/internal/storage"
)

func SetupRouter(router *chi.Mux, queries *db.Queries, rdb *redis.Client, s3Client *storage.S3Client) {

	registerLimiter := middleware.NewStore(1*time.Hour, 3); // 3 accounts / hour
	loginLimiter := middleware.NewStore(10*time.Second, 3); // 3 login attempts / 10 seconds
	userLimiter := middleware.NewStore(1*time.Minute, 20); // 20 requests / minute 


	router.Get("/health", handlers.HandlerReadiness)

	authHandler := &handlers.AuthHandler{Q: queries}

	router.With(registerLimiter.Limit).Post("/auth/register", authHandler.Register)
	router.With(loginLimiter.Limit).Post("/auth/login", authHandler.Login)
	router.Post("/auth/refresh", authHandler.RefreshToken)

	userHandler := &handlers.UserHandler{
		Q:     queries,
		Redis: rdb,
		S3:    s3Client,
	}

	router.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Use(userLimiter.Limit)

		r.Get("/user", userHandler.GetUser)
		r.Put("/user/avatar",userHandler.UpdateUserAvatar)
		r.Patch("/user/{user_id}", userHandler.UpdateUser)
	})

}