package routes

import (
	"database/sql"
	"net/http"

	"login-auth-template/handlers"
	"login-auth-template/middleware"
)

// SETS ALL ROUTES, RETURNS ServeMux
func NewRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// init handlers with shared db conn
	authHandler := &handlers.AuthHandler{DB: db}
	userHandler := &handlers.UserHandler{DB: db}

	// --- PUBLIC ROUTES ---
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	// --- PROTECTED ROUTES (req auth) ---
	mux.Handle("/me", middleware.AuthMiddleware(http.HandlerFunc(userHandler.GetMe)))
	mux.Handle("/me/delete", middleware.AuthMiddleware(http.HandlerFunc(userHandler.DeleteMe)))

	return mux
}
