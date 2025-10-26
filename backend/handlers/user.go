package handlers

import (
	"database/sql"
	"net/http"

	"login-auth-template/models"
	"login-auth-template/utils"
)

// UserHandler holds DB reference
type UserHandler struct {
	DB *sql.DB
}

// GET /me — return currently authenticated user
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by AuthMiddleware)
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := models.GetUserByID(h.DB, userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// DELETE /me — delete the current user
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := models.DeleteUser(h.DB, userID); err != nil {
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
