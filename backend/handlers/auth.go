package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"login-auth-template/models"
	"login-auth-template/utils"
)

// go's encoding/json is type safe
// define struct to match payload
// lets us safely check shape
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthHandler holds DB reference for handlers that need it
// define this in main. lets you reuse connection
type AuthHandler struct {
	DB *sql.DB
}

// POST /register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// hashing
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	// create users in db
	user, err := models.CreateUser(h.DB, req.Email, req.Username, hash)
	if err != nil {
		http.Error(w, "failed to create user (email may already exist)", http.StatusBadRequest)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user)
}

// POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Look up user by email
	user, err := models.GetUserByEmail(h.DB, req.Email)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	// Check password match
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	// Create JWT token
	token, err := utils.CreateToken(user.ID, user.Email, 24*time.Hour)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
