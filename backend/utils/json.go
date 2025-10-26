package utils

import (
	"encoding/json"
	"net/http"
)

// helper struct for consistent responses
type JSONResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// writes a JSON response with the given status code and payload
func WriteJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// sending error messages
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, JSONResponse{
		Status:  "error",
		Message: message,
	})
}

// sending success responses
func WriteSuccess(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusOK, JSONResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}
