package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Tanveer-rajpurohit/start/internal/db"
	"github.com/Tanveer-rajpurohit/start/internal/utils"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	Q *db.Queries
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}


func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request){
	var req CreateUserRequest
	// json.NewDecoder(r.Body).Decode(&req)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
    	utils.ResponseWithError(w, 400, "invalid request body")
    	return
	}

	user, err := h.Q.CreateUser(context.Background(), db.CreateUserParams{
		Name: req.Name,
		Email: req.Email,
	})

	if err != nil {
		utils.ResponseWithError(w, 500, "failed to create user")
		return
	}

	utils.ResponseWithJSON(w, 201, user)
}


func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request){
	users, err := h.Q.GetUsers(context.Background())
	if err != nil {
		utils.ResponseWithError(w, 500, "failed to fetch users")
		return
	}
	utils.ResponseWithJSON(w, 200, users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request){
	idstr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idstr)

	user, err := h.Q.GetUser(context.Background(), int32(id))
	if err != nil {
		utils.ResponseWithError(w, 500, "failed to fetch user")
		return
	}
	utils.ResponseWithJSON(w, 200, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request){
	idstr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idstr)
	var req CreateUserRequest
	// json.NewDecoder(r.Body).Decode(&req)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
    	utils.ResponseWithError(w, 400, "invalid request body")
    	return
	}

	user, err := h.Q.UpdateUser(context.Background(),db.UpdateUserParams{
		ID:    int32(id),
		Name: req.Name,
		Email: req.Email,
	})

	if err != nil {
		utils.ResponseWithError(w, 500, "update failed")
		return
	}

	utils.ResponseWithJSON(w, 200, user)
}


func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	h.Q.DeleteUser(context.Background(), int32(id))

	utils.ResponseWithJSON(w, 200, map[string]string{"message": "deleted"})
}