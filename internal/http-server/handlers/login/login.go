package login

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/lib/api/response"
	"github.com/studopolis/auth-server/internal/lib/api/validation"
	"github.com/studopolis/auth-server/internal/lib/http/codec"
	"github.com/studopolis/auth-server/internal/lib/jwt"
	"github.com/studopolis/auth-server/internal/lib/logger"
	"github.com/studopolis/auth-server/internal/lib/secrets"
	st "github.com/studopolis/auth-server/internal/storage"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger, a *jwt.JWTService, s *storage.Storage) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.login.New"

		log := log.With(
			logger.Operation(op),
			logger.RequestID(requestMiddleware.GetID(r.Context())),
		)

		c := &validation.Credentials{}

		err := codec.DecodeJSON(r.Body, c)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")
				codec.JSONResponse(w, r, response.Error("request body is empty"))
				return
			}

			log.Error("failed to decode request body", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("cannot create user"))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), s.Options.ReadTimeout)
		defer cancel()

		user, err := s.UserByEmail(ctx, c.Email)
		if err != nil {
			if errors.Is(err, st.ErrUserNotFound) {
				log.Warn("user not found", logger.Error(err))
				codec.JSONResponse(w, r, response.Error("invalid credentials"))
				return
			}

			log.Error("failed to get user", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError())
			return
		}

		if err = secrets.CompareHashAndPassword(user.PasswordHash, c.Password); err != nil {
			log.Info("invalid credentials", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("invalid credentials"))
			return
		}

		refreshToken, exp, err := a.IssueRefresh(user)
		if err != nil {
			log.Error("cannot issue refresh token", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError())
			return
		}
		jwt.SetRefreshToken(w, refreshToken, exp)

		accessToken, _, err := a.IssueAccess(user)
		if err != nil {
			log.Error("cannot issue token", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError())
			return
		}
		jwt.SetAccessToken(w, accessToken)

		response := response.Ok("user logged in successfully")
		codec.JSONResponse(w, r, response)
	}

	return http.HandlerFunc(handler)
}
