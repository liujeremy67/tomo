package routes

import (
	"database/sql"
	"net/http"

	"tomo/backend/handlers"
	"tomo/backend/middleware"
)

// NewRouter sets all routes and returns ServeMux
func NewRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize handlers with shared db connection
	authHandler := &handlers.AuthHandler{DB: db}
	userHandler := &handlers.UserHandler{DB: db}
	sessionHandler := &handlers.SessionHandler{DB: db}
	postHandler := &handlers.PostHandler{DB: db}
	mediaHandler := &handlers.MediaHandler{DB: db}

	// --- PUBLIC ROUTES ---
	mux.HandleFunc("POST /auth/google", authHandler.GoogleAuth)
	mux.HandleFunc("GET /users/{username}", userHandler.GetUserByUsername)

	// --- PROTECTED ROUTES (require auth) ---
	// User routes
	mux.Handle("GET /me", middleware.AuthMiddleware(http.HandlerFunc(userHandler.GetMe)))
	mux.Handle("PATCH /me", middleware.AuthMiddleware(http.HandlerFunc(userHandler.UpdateMe)))
	mux.Handle("DELETE /me", middleware.AuthMiddleware(http.HandlerFunc(userHandler.DeleteMe)))

	// Session routes
	mux.Handle("POST /sessions", middleware.AuthMiddleware(http.HandlerFunc(sessionHandler.CreateSession)))
	mux.Handle("GET /sessions", middleware.AuthMiddleware(http.HandlerFunc(sessionHandler.GetMySessions)))
	mux.Handle("GET /sessions/{id}", middleware.AuthMiddleware(http.HandlerFunc(sessionHandler.GetSession)))
	mux.Handle("DELETE /sessions/{id}", middleware.AuthMiddleware(http.HandlerFunc(sessionHandler.DeleteSession)))

	// Post routes
	mux.Handle("POST /posts", middleware.AuthMiddleware(http.HandlerFunc(postHandler.CreatePost)))
	mux.Handle("GET /posts/user/{id}", middleware.AuthMiddleware(http.HandlerFunc(postHandler.GetUserPosts)))
	mux.Handle("GET /posts/{id}", middleware.AuthMiddleware(http.HandlerFunc(postHandler.GetPost)))
	mux.Handle("PATCH /posts/{id}", middleware.AuthMiddleware(http.HandlerFunc(postHandler.UpdatePost)))
	mux.Handle("DELETE /posts/{id}", middleware.AuthMiddleware(http.HandlerFunc(postHandler.DeletePost)))

	// Media routes
	mux.Handle("POST /posts/{id}/media", middleware.AuthMiddleware(http.HandlerFunc(mediaHandler.AddMediaToPost)))
	mux.Handle("POST /posts/{id}/media/upload", middleware.AuthMiddleware(http.HandlerFunc(mediaHandler.UploadMediaFile)))
	mux.Handle("GET /posts/{id}/media", middleware.AuthMiddleware(http.HandlerFunc(mediaHandler.GetPostMedia)))
	mux.Handle("DELETE /media/{id}", middleware.AuthMiddleware(http.HandlerFunc(mediaHandler.DeleteMedia)))

	return mux
}
