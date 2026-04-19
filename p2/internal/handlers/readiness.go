package handlers

import (
	"net/http"

	"github.com/Tanveer-rajpurohit/p2/internal/utils"
)

func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"message": "ok",
	}
	utils.ResponseWithJSON(w, http.StatusOK, data)
}