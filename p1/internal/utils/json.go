package utils

import (
	"encoding/json"
	"net/http"
)

func ResponseWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload) // json.marshal converts the payload to JSON format and returns the byte slice and error and if we want to cover in string we can use string(response)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		// ResponseWithError(w, http.StatusInternalServerError, "Failed to marshal JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
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