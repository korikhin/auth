package jwt

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	ctx "github.com/studopolis/auth-server/internal/lib/context"
	"github.com/studopolis/auth-server/internal/lib/jwt"
	// "github.com/studopolis/auth-server/internal/lib/logger"
)

var userCtxKey = &ctx.ContextKey{Name: "User"}

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	log.Info("jwt middleware enabled")
	// log = log.With(
	// 	logger.Component("middleware/jwt"),
	// )

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

			// todo: add isAdmin() (is iam.admin)
			if claims.UserRole != requiredRole {
				http.Error(w, "Access not granted", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(handler)
	}
}
