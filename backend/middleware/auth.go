package middleware

import (
	"context"
	"net/http"
	"strings"

	"tomo/backend/utils"
)

// key type to store user info in context
// look into context later
type contextKey string

const UserContextKey = contextKey("user")

type UserClaims struct {
	UserID int
	Email  string
}

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

		user := UserClaims{
			UserID: int((*claims)["sub"].(float64)), // JWT numbers decode as float64
			Email:  (*claims)["email"].(string),
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to extract user info from context
func GetUserFromContext(r *http.Request) (map[string]interface{}, bool) {
	user, ok := r.Context().Value(UserContextKey).(map[string]interface{})
	return user, ok
}
