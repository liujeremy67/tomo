package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"tomo/backend/models"
	"tomo/backend/utils"
)

type GoogleAuthRequest struct {
	IDToken string `json:"id_token"`
}

// AuthHandler holds DB reference for handlers that need it
type AuthHandler struct {
	DB *sql.DB
}

// POST /auth/google
// Accepts Google ID token, verifies it, creates or logs in user
func (h *AuthHandler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	var req GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.IDToken == "" {
		http.Error(w, "id_token is required", http.StatusBadRequest)
		return
	}

	// Verify the Google token
	googlePayload, err := utils.VerifyGoogleToken(r.Context(), req.IDToken)
	if err != nil {
		http.Error(w, "invalid or expired Google token", http.StatusUnauthorized)
		return
	}

	// Try to find existing user by Google ID
	user, err := models.GetUserByGoogleID(h.DB, googlePayload.GoogleID)

	if err == sql.ErrNoRows {
		// User doesn't exist - create blank profile with just Google ID and email
		user, err = models.CreateUser(h.DB, googlePayload.GoogleID, googlePayload.Email)
		if err != nil {
			// Check if email already exists with different Google ID
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
				http.Error(w, "email already registered with different account", http.StatusConflict)
				return
			}
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Create JWT token for our application
	token, err := utils.CreateToken(user.ID, user.Email, 24*time.Hour)
	if err != nil {
		http.Error(w, "failed to create session token", http.StatusInternalServerError)
		return
	}

	// Return token and user info
	response := map[string]interface{}{
		"token": token,
		"user":  user,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}
