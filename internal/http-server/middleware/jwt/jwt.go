package jwt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/config"
	httplib "github.com/studopolis/auth-server/internal/lib/http"
	"github.com/studopolis/auth-server/internal/lib/jwt"
	"github.com/studopolis/auth-server/internal/lib/logger"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger, s *storage.Storage, config config.JWT) func(next http.Handler) http.Handler {
	log.Info("jwt middleware enabled")

	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			log = log.With(
				logger.Component("middleware/jwt"),
				logger.RequestID(requestMiddleware.GetID(r.Context())),
			)

			accessToken, err := jwt.GetAccessToken(r)
			if err != nil {
				log.Error("cannot get access token", logger.Error(err))
				http.Error(w, "Token is missing", http.StatusUnauthorized)
				return
			}

			mask := &jwt.ValidationMask{
				IssuedAt: true,
				Issuer:   config.Issuer,
				Leeway:   config.Leeway,
			}

			claims, err := jwt.Validate(accessToken, jwt.AccessTokenScope, mask)
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
					log.Error(fmt.Sprintf("cannot find user: %s", userID), logger.Error(err))
					http.Error(w, "User not found", http.StatusInternalServerError)
					return
				}

				// update claims
				// claims.UserRole = user.Role

				refreshToken, err := jwt.GetRefreshToken(r)
				if err != nil {
					log.Error("cannot get refresh token", logger.Error(err))
					http.Error(w, "Token is missing", http.StatusUnauthorized)
					return
				}

				mask := &jwt.ValidationMask{
					IssuedAt: true,
					Issuer:   claims.Issuer,
					Subject:  claims.Subject,
					Leeway:   config.Leeway,
				}

				if _, err := jwt.Validate(refreshToken, jwt.RefreshTokenScope, mask); err != nil {
					log.Error("cannot validate refresh token", logger.Error(err))
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}

				refreshToken, err = jwt.Issue(user, jwt.RefreshTokenScope, config)
				if err != nil {
					log.Error("cannot issue refresh token", logger.Error(err))
					http.Error(w, "Cannot issue token", http.StatusInternalServerError)
					return
				}

				if err = jwt.SetRefreshToken(w, refreshToken); err != nil {
					log.Error("cannot set refresh token", logger.Error(err))
					http.Error(w, "Cannot issue token", http.StatusInternalServerError)
					return
				}

				accessToken, err = jwt.Issue(user, jwt.AccessTokenScope, config)
				if err != nil {
					log.Error("cannot issue token", logger.Error(err))
					http.Error(w, "Cannot issue token", http.StatusInternalServerError)
					return
				}

				jwt.SetAccessToken(w, accessToken)
			}

			ctxHTTP := context.WithValue(r.Context(), httplib.UserCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctxHTTP))
		}

		return http.HandlerFunc(handler)
	}
}
