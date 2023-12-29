package login

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/lib/api/response"
	"github.com/studopolis/auth-server/internal/lib/api/validation"
	"github.com/studopolis/auth-server/internal/lib/http/codec"
	"github.com/studopolis/auth-server/internal/lib/jwt"
	"github.com/studopolis/auth-server/internal/lib/logger"
	"github.com/studopolis/auth-server/internal/lib/secrets"
	stg "github.com/studopolis/auth-server/internal/storage"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger, s *storage.Storage, config config.JWT) http.Handler {
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
				codec.JSONResponse(w, r, response.Error("Request body is empty"))
				return
			}

			log.Error("failed to decode request body", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("Cannot create user"))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), s.Options.ReadTimeout)
		defer cancel()

		user, err := s.UserByEmail(ctx, c.Email)
		if err != nil {
			if errors.Is(err, stg.ErrUserNotFound) {
				log.Warn("user not found", logger.Error(err))
				codec.JSONResponse(w, r, response.Error("Invalid credentials"))
				return
			}

			log.Error("failed to get user", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError())
			return
		}

		if err = secrets.CompareHashAndPassword(user.PasswordHash, c.Password); err != nil {
			log.Info("invalid credentials", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("Invalid credentials"))
			return
		}

		refreshToken, err := jwt.Issue(user, jwt.RefreshTokenScope, config)
		if err != nil {
			log.Error("cannot issue refresh token", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError())
			return
		}

		if err = jwt.SetRefreshToken(w, refreshToken); err != nil {
			log.Error("cannot set refresh token", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError())
			return
		}

		accessToken, err := jwt.Issue(user, jwt.AccessTokenScope, config)
		if err != nil {
			log.Error("cannot issue token", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError())
			return
		}

		jwt.SetAccessToken(w, accessToken)

		response := response.Ok("User logged in successfully")
		codec.JSONResponse(w, r, response)
	}

	return http.HandlerFunc(handler)
}
