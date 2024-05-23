package login

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/korikhin/auth/internal/lib/api/response"
	"github.com/korikhin/auth/internal/lib/api/validation"
	"github.com/korikhin/auth/internal/lib/http/codec"
	"github.com/korikhin/auth/internal/lib/jwt"
	"github.com/korikhin/auth/internal/lib/logger"
	st "github.com/korikhin/auth/internal/storage"
	storage "github.com/korikhin/auth/internal/storage/postgres"

	reqMW "github.com/korikhin/auth/internal/http-server/middleware/request"

	"golang.org/x/crypto/bcrypt"
)

// TODO?: Refactor token (re)issuing
func New(log *slog.Logger, a *jwt.JWTService, s *storage.Storage) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.login.New"

		log := log.With(
			logger.Operation(op),
			logger.RequestID(reqMW.GetID(r.Context())),
		)

		c := &validation.Credentials{}
		err := codec.DecodeJSON(r.Body, c)
		if err != nil {
			log.Error("failed to decode request body", logger.Error(err))
			codec.JSONResponse(w, response.InternalError, http.StatusInternalServerError)
			return
		}

		ctxStorage, cancel := context.WithTimeout(context.Background(), s.Options.ReadTimeout)
		defer cancel()

		user, err := s.UserByEmail(ctxStorage, c.Email)
		if err != nil {
			if errors.Is(err, st.ErrUserNotFound) {
				log.Warn("user not found", logger.Error(err))
				codec.JSONResponse(w, response.Error("user not found"), http.StatusNotFound)
				return
			}

			log.Error("failed to get user", logger.Error(err))
			codec.JSONResponse(w, response.InternalError, http.StatusInternalServerError)
			return
		}

		if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(c.Password)); err != nil {
			log.Info("invalid credentials", logger.Error(err))
			codec.JSONResponse(w, response.Error("invalid credentials"), http.StatusUnauthorized)
			return
		}

		refreshToken, exp, err := a.IssueRefresh(user)
		if err != nil {
			log.Error("cannot issue refresh token", logger.Error(err))
			codec.JSONResponse(w, response.InternalError, http.StatusInternalServerError)
			return
		}
		jwt.SetRefreshToken(w, refreshToken, exp)

		accessToken, _, err := a.IssueAccess(user)
		if err != nil {
			log.Error("cannot issue token", logger.Error(err))
			codec.JSONResponse(w, response.InternalError, http.StatusInternalServerError)
			return
		}
		jwt.SetAccessToken(w, accessToken)

		codec.JSONResponse(w, response.Ok("user logged successfully"), http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
