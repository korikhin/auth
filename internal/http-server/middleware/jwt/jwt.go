package jwt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	ctxlib "github.com/korikhin/auth/internal/lib/context"
	"github.com/korikhin/auth/internal/lib/jwt"
	"github.com/korikhin/auth/internal/lib/logger"
	storage "github.com/korikhin/auth/internal/storage/postgres"

	reqMW "github.com/korikhin/auth/internal/http-server/middleware/request"
)

// TODO?: Refactor token (re)issuing
func New(log *slog.Logger, a *jwt.JWTService, s *storage.Storage) func(next http.Handler) http.Handler {
	log.Info("jwt middleware enabled")
	log = log.With(logger.Component("middleware/jwt"))

	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			log := log.With(
				logger.RequestID(reqMW.GetID(r.Context())),
			)

			accessToken, err := jwt.GetAccessToken(r)
			if err != nil {
				log.Error("cannot get access token", logger.Error(err))
				http.Error(w, "Token is missing", http.StatusUnauthorized)
				return
			}

			opts := jwt.ValidationOptions{
				Issuer: a.Options.Issuer,
				Leeway: a.Options.Leeway,
			}

			claims, err := a.ValidateAccess(accessToken, opts)
			if err != nil && !errors.Is(err, jwt.ErrTokenExpiredOnly) {
				log.Error("cannot validate token", logger.Error(err))
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if errors.Is(err, jwt.ErrTokenExpiredOnly) {
				ctxStorage, cancel := context.WithTimeout(context.Background(), s.Options.ReadTimeout)
				defer cancel()

				userID := claims.Subject
				user, err := s.User(ctxStorage, userID)
				if err != nil {
					log.Warn(fmt.Sprintf("cannot find user: %v", userID), logger.Error(err))
					http.Error(w, "User not found", http.StatusNotFound)
					return
				}

				refreshToken, err := jwt.GetRefreshToken(r)
				if err != nil {
					log.Error("cannot get refresh token", logger.Error(err))
					http.Error(w, "Token is missing", http.StatusUnauthorized)
					return
				}

				opts := jwt.ValidationOptions{
					Issuer:  claims.Issuer,
					Leeway:  a.Options.Leeway,
					Subject: claims.Subject,
				}

				if _, err := a.ValidateRefresh(refreshToken, opts); err != nil {
					log.Error("cannot validate refresh token", logger.Error(err))
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}

				refreshToken, exp, err := a.IssueRefresh(user)
				if err != nil {
					log.Error("cannot issue refresh token", logger.Error(err))
					http.Error(w, "Cannot issue token", http.StatusInternalServerError)
					return
				}
				jwt.SetRefreshToken(w, refreshToken, exp)

				accessToken, _, err = a.IssueAccess(user)
				if err != nil {
					log.Error("cannot issue token", logger.Error(err))
					http.Error(w, "Cannot issue token", http.StatusInternalServerError)
					return
				}
				jwt.SetAccessToken(w, accessToken)
			}

			ctx := context.WithValue(r.Context(), ctxlib.UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(handler)
	}
}
