package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Tanveer-rajpurohit/start/internal/auth"
	"github.com/Tanveer-rajpurohit/start/internal/db"
	"github.com/Tanveer-rajpurohit/start/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Q *db.Queries
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.ResponseWithError(w, 400, "invalid request body")
		return
	}

	_, err := h.Q.GetUserByEmail(context.Background(), req.Email)
	if err == nil {
		utils.ResponseWithError(w, 400, "email already in use")
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}
	user, err := h.Q.CreateUserWithPassword(context.Background(), db.CreateUserWithPasswordParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
	})
	if err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	utils.ResponseWithJSON(w, 201, user)

}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.ResponseWithError(w, 400, "invalid request body")
		return
	}
	user, err := h.Q.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.ResponseWithError(w, 400, "invalid email or password")
			return
		}
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.ResponseWithError(w, 400, "invalid email or password")
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	if _, err := h.Q.SaveRefreshToken(context.Background(), db.SaveRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
	}); err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	utils.ResponseWithJSON(w, 200, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.ResponseWithError(w, 400, "invalid request body")
		return
	}

	tokenStr := body["refresh_token"]
	if tokenStr == "" {
		utils.ResponseWithError(w, 400, "refresh_token is required")
		return
	}

	claims, err := auth.ValidateToken(tokenStr, true)
	if err != nil {
		utils.ResponseWithError(w, 401, "invalid refresh token")
		return
	}

	storedToken, err := h.Q.GetRefreshToken(context.Background(), tokenStr)
	if err != nil || !storedToken.ExpiresAt.Valid || !storedToken.ExpiresAt.Time.After(time.Now()) {
		utils.ResponseWithError(w, 401, "invalid refresh token")
		return
	}

	accessToken, err := auth.GenerateAccessToken(claims.UserID, claims.Role)
	if err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	utils.ResponseWithJSON(w, 200, map[string]string{
		"access_token": accessToken,
	})
}
