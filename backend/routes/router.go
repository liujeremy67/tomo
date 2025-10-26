package routes

import (
	"database/sql"
	"net/http"

	"login-auth-template/handlers"
	"login-auth-template/middleware"
)

// NewRouter sets all routes and returns ServeMux
func NewRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize handlers with shared db connection
	authHandler := &handlers.AuthHandler{DB: db}
	userHandler := &handlers.UserHandler{DB: db}

	// --- PUBLIC ROUTES ---
	mux.HandleFunc("POST /auth/google", authHandler.GoogleAuth)

	// --- PROTECTED ROUTES (require auth) ---
	mux.Handle("GET /me", middleware.AuthMiddleware(http.HandlerFunc(userHandler.GetMe)))
	mux.Handle("DELETE /me", middleware.AuthMiddleware(http.HandlerFunc(userHandler.DeleteMe)))

	return mux
}
