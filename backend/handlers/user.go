package handlers

import (
	"database/sql"
	"net/http"

	"login-auth-template/middleware"
	"login-auth-template/models"
	"login-auth-template/utils"
)

type UserHandler struct {
	DB *sql.DB
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
