package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"github.com/Tanveer-rajpurohit/p2/internal/auth"
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/storage"
	"github.com/Tanveer-rajpurohit/p2/internal/utils"
)

type UserHandler struct {
	Q     *db.Queries
	Redis *redis.Client
	S3    *storage.S3Client
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

	cachedUser, err := h.Redis.Get(context.Background(), cacheKey).Result() // If we have a cached user, return it and .Result() will return the string value of the cached user. If there is no cached user, it will return an error which we ignore and proceed to fetch from the database.
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

func (h *UserHandler) UpdateUserAvatar(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsKey).(*auth.Claims)
	if !ok {
		utils.ResponseWithError(w, 401, "unauthorized")
		return
	}

	// ParseMultipartForm loads up to 10MB into memory, rest goes to temp disk file
	// This prevents a 100MB upload from crashing your server RAM
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.ResponseWithError(w, 400, "file too large or invalid form")
		return
	}

	// "avatar" is the form field name the client must use
	file, header, err := r.FormFile("avatar")
	if err != nil {
		utils.ResponseWithError(w, 400, "avatar file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" && contentType != "image/webp" {
		utils.ResponseWithError(w, 400, "unsupported file type (only JPEG, PNG, JPG, and WEBP allowed)")
		return
	}

	// Upload raw file to S3 immediately — stream directly, no RAM copy
	rawKey := fmt.Sprintf("raw/avatars/%s/original.jpg", claims.UserID)
	rawUrl, err := h.S3.Upload(r.Context(), rawKey, file, contentType)
	if err != nil {
		utils.ResponseWithError(w, 500, "failed to upload avatar")
		log.Printf("S3 upload error: %v", err)
		return
	}

	userIdUUID, _ := uuid.Parse(claims.UserID)
	_, err = h.Q.UpdateUserAvatar(r.Context(), db.UpdateUserAvatarParams{
		ID:        userIdUUID,
		AvatarRaw: pgtype.Text{String: rawUrl, Valid: true},
	})
	if err != nil {
		utils.ResponseWithError(w, 500, "failed to update user avatar")
		log.Printf("DB UpdateUserAvatar error: %v", err)
		return
	}

	// Push a job to the avatar queue — worker picks it up and processes it

	utils.ResponseWithJSON(w, 202, map[string]string{
		"message":      "avatar uploaded, processing in background",
		"original_url": rawUrl,
	})
}

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
