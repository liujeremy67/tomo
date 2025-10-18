package middleware

import (
	"encoding/json"
	"net/http"
)

type SessionRequest struct {
	UserID  uint   `json:"user_id"`
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
}

func ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body SessionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if body.UserID == 0 || body.StartAt == "" {
			http.Error(w, "Missing required fields: user_id, start_at", http.StatusBadRequest)
			return
		}
		// Attach to context if needed later
		next.ServeHTTP(w, r)
	})
}
