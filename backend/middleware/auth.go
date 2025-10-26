package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"login-auth-template/utils"
)

// key type to store user info in context
// look into context later
type contextKey string

const userContextKey = contextKey("user")

// AuthMiddleware validates JWTs for protected routes
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Expect format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid Authorization format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		// Validate token using utils.ValidateToken
		claims, err := utils.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Optional: extract useful info from claims
		userData := map[string]interface{}{
			"user_id": (*claims)["sub"],
			"email":   (*claims)["email"],
		}

		// Attach claims to context for use by handlers
		ctx := context.WithValue(r.Context(), userContextKey, userData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to extract user info from context
func GetUserFromContext(r *http.Request) (map[string]interface{}, bool) {
	user, ok := r.Context().Value(userContextKey).(map[string]interface{})
	return user, ok
}

// Optional: JSON helper (for consistent responses)
func JSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
