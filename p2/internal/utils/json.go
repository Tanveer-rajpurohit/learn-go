package utils

import (
	"encoding/json"
	"net/http"
)

func ResponseWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		http.Error(w,"Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func ResponseWithError(w http.ResponseWriter, status int, message string) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}

	ResponseWithJSON(w,status,ErrorResponse{
		Error: message,
	})
}