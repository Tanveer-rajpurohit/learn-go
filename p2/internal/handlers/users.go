package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"github.com/Tanveer-rajpurohit/p2/internal/auth"
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/utils"
)

type UserHandler struct {
	Q     *db.Queries
	Redis *redis.Client
}

type UpdateUserAvatarRequest struct {
	ID     string `json:"user_id"`
	Avatar string `json:"avatar"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsKey).(*auth.Claims)
	if !ok {
		utils.ResponseWithError(w, 401, "unauthorized")
		return
	}

	parsedID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.ResponseWithError(w, 400, "invalid user id format")
		return
	}

	cacheKey := "user:profile:" + parsedID.String()

	cachedUser, err := h.Redis.Get(context.Background(), cacheKey).Result()// If we have a cached user, return it and .Result() will return the string value of the cached user. If there is no cached user, it will return an error which we ignore and proceed to fetch from the database.
	if err == nil {
		var user db.User
		if json.Unmarshal([]byte(cachedUser), &user) == nil {
			utils.ResponseWithJSON(w, 200, user)
			return
		}
	}


	user, err := h.Q.GetUserByID(r.Context(), parsedID)
	if err != nil {
		utils.ResponseWithError(w, 404, "user not found")
		return
	}

	userBytes, _ := json.Marshal(user) // Marshal the user struct to JSON bytes
	h.Redis.Set(context.Background(), cacheKey, userBytes, 15*time.Minute)

	utils.ResponseWithJSON(w, 200, user)
}

// func (h *UserHandler) UpdateUserAvatar(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.ResponseWithError(w, 400, "invalid request body")
		return
	}

	userId := chi.URLParam(r, "user_id")

	//userId from the URL param should match the userId from the token claims to prevent users from updating other users' profiles
	claims, ok := r.Context().Value(auth.ClaimsKey).(*auth.Claims)
	if !ok || claims.UserID != userId {
		utils.ResponseWithError(w, 401, "unauthorized")
		return
	}

	parsedID, err := uuid.Parse(userId)

	if err != nil {
		utils.ResponseWithError(w, 400, "invalid user id format")
		return
	}

	nameParam := pgtype.Text{}
	if req.Name != nil {
		nameParam = pgtype.Text{String: *req.Name, Valid: true}
	}

	emailParam := pgtype.Text{}
	if req.Email != nil {
		emailParam = pgtype.Text{String: *req.Email, Valid: true}
	}

	updatedUser, err := h.Q.UpdateUser(r.Context(), db.UpdateUserParams{
		ID:    parsedID,
		Name:  nameParam,
		Email: emailParam,
	})

	if err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	// Invalidate the cache for the updated user
	cacheKey := "user:profile:" + parsedID.String()
	h.Redis.Del(context.Background(), cacheKey)

	utils.ResponseWithJSON(w, 200, updatedUser)
}
