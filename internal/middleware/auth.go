package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/feline-dis/go-radio-v2/internal/services"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware creates middleware that validates JWT tokens
func AuthMiddleware(jwtService *services.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				http.Error(w, "Token is required", http.StatusUnauthorized)
				return
			}

			// Validate the token
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user info to request context
			ctx := context.WithValue(r.Context(), UserContextKey, claims.Username)
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts the username from the request context
func GetUserFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UserContextKey).(string)
	return username, ok
} 