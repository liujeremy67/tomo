package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"login-auth-template/middleware"
	"login-auth-template/models"
	"login-auth-template/utils"
)

type UserHandler struct {
	DB *sql.DB
}

type UpdateProfileRequest struct {
	Username    *string `json:"username,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	PictureURL  *string `json:"picture_url,omitempty"`
}

// GET /me — return currently authenticated user
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by AuthMiddleware)
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Use user.UserID from JWT claims
	dbUser, err := models.GetUserByID(h.DB, user.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	utils.WriteJSON(w, http.StatusOK, dbUser)
}

// PATCH /me — update current user's profile
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get current user data
	currentUser, err := models.GetUserByID(h.DB, user.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Prepare values (use existing if not provided)
	username := currentUser.Username
	displayName := currentUser.DisplayName
	pictureURL := currentUser.PictureURL

	if req.Username != nil {
		newUsername := strings.TrimSpace(*req.Username)

		// Validate username
		if newUsername == "" {
			http.Error(w, "username cannot be empty", http.StatusBadRequest)
			return
		}
		if len(newUsername) < 3 || len(newUsername) > 30 {
			http.Error(w, "username must be between 3 and 30 characters", http.StatusBadRequest)
			return
		}

		// Check if username is already taken (by someone else)
		if newUsername != currentUser.Username {
			exists, err := models.UsernameExists(h.DB, newUsername)
			if err != nil {
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
			if exists {
				http.Error(w, "username already taken", http.StatusConflict)
				return
			}
		}

		username = newUsername
	}

	if req.DisplayName != nil {
		displayName = strings.TrimSpace(*req.DisplayName)
		if len(displayName) > 50 {
			http.Error(w, "display_name must be 50 characters or less", http.StatusBadRequest)
			return
		}
	}

	if req.PictureURL != nil {
		pictureURL = strings.TrimSpace(*req.PictureURL)
		if len(pictureURL) > 500 {
			http.Error(w, "picture_url must be 500 characters or less", http.StatusBadRequest)
			return
		}
	}

	// Update profile
	if err := models.UpdateProfile(h.DB, user.UserID, username, displayName, pictureURL); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "username already taken", http.StatusConflict)
			return
		}
		http.Error(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	// Return updated user
	updatedUser, err := models.GetUserByID(h.DB, user.UserID)
	if err != nil {
		http.Error(w, "failed to fetch updated user", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, updatedUser)
}

// DELETE /me — delete the current user
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := models.DeleteUser(h.DB, user.UserID); err != nil {
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

// GET /users/{username} — get user by username (public)
func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(h.DB, username)
	if err == sql.ErrNoRows {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Return public user data (don't expose email or google_id)
	publicUser := map[string]interface{}{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"picture_url":  user.PictureURL,
		"created_at":   user.CreatedAt,
	}

	utils.WriteJSON(w, http.StatusOK, publicUser)
}
