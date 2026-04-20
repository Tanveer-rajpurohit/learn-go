package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Tanveer-rajpurohit/p2/internal/auth"
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/utils"
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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
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
		log.Printf("GetUserByEmail error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt GenerateFromPassword error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	user, err := h.Q.CreateUserWithPassword(context.Background(), db.CreateUserWithPasswordParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
	})

	if err != nil {
		log.Printf("CreateUserWithPassword error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID.String(), user.Role, user.Email)
	if err != nil {
		log.Printf("GenerateAccessToken error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID.String(), user.Role, user.Email)
	if err != nil {
		log.Printf("GenerateRefreshToken error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	if _, err := h.Q.SaveRefreshToken(context.Background(), db.SaveRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
	}); err != nil {
		log.Printf("SaveRefreshToken error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	utils.ResponseWithJSON(w, 201, map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
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
		log.Printf("Login GetUserByEmail error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.ResponseWithError(w, 400, "invalid email or password")
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID.String(), user.Role, user.Email)
	if err != nil {
		log.Printf("Login GenerateAccessToken error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID.String(), user.Role, user.Email)
	if err != nil {
		log.Printf("Login GenerateRefreshToken error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	if _, err := h.Q.SaveRefreshToken(context.Background(), db.SaveRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
	}); err != nil {
		log.Printf("Login SaveRefreshToken error: %v", err)
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	utils.ResponseWithJSON(w, 200, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})

}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.ResponseWithError(w, 400, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		utils.ResponseWithError(w, 400, "refresh_token is required")
		return
	}
	claims, err := auth.ValidateToken(req.RefreshToken, true)
	if err != nil {
		utils.ResponseWithError(w, 400, "invalid refresh token")
		return
	}

	storedToken, err := h.Q.GetRefreshToken(context.Background(), req.RefreshToken)
	if err != nil || !storedToken.ExpiresAt.Valid || !storedToken.ExpiresAt.Time.After(time.Now()) || storedToken.UserID.String() != claims.UserID {
		utils.ResponseWithError(w, 401, "invalid refresh token")
		return
	}

	accessToken, err := auth.GenerateAccessToken(claims.UserID, claims.Role, claims.Email)
	if err != nil {
		utils.ResponseWithError(w, 500, "internal server error")
		return
	}

	utils.ResponseWithJSON(w, 200, map[string]string{
		"access_token": accessToken,
	})
}
