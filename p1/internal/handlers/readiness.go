package handlers

import (
	"net/http"

	"github.com/Tanveer-rajpurohit/start/internal/utils"
)

func HandlerReadiness(w http.ResponseWriter, r *http.Request){
	data := map[string]string{
		"message": "ok",
	}
	utils.ResponseWithJSON(w, http.StatusOK, data)
}