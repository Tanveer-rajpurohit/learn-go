package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Tanveer-rajpurohit/p2/internal/auth"
	"github.com/Tanveer-rajpurohit/p2/internal/db"
	"github.com/Tanveer-rajpurohit/p2/internal/utils"
)

type UserHandler struct {
	Q *db.Queries
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

	user, err := h.Q.GetUserByID(r.Context(), parsedID)
	if err != nil {
		utils.ResponseWithError(w, 404, "user not found")
		return
	}

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

	utils.ResponseWithJSON(w, 200, updatedUser)
}
