package jwt

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/studopolis/auth-server/internal/lib/jwt"
)

type contextKey string

const (
	userKey contextKey = "user"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			tokenHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if tokenHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
			claims, err := jwt.Validate(tokenString)

			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			requiredRole := r.Header.Get("X-Required-Role")
			if requiredRole == "" {
				http.Error(w, "User role missing", http.StatusForbidden)
				return
			}

			// todo: add isAdmin()
			if claims.UserRole != requiredRole {
				http.Error(w, "Access not granted", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(handler)
	}
}
