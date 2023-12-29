package register

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/lib/api/response"
	"github.com/studopolis/auth-server/internal/lib/api/validation"
	"github.com/studopolis/auth-server/internal/lib/http/codec"
	"github.com/studopolis/auth-server/internal/lib/logger"
	"github.com/studopolis/auth-server/internal/lib/secrets"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger, s *storage.Storage) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.register.New"

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
			codec.JSONResponse(w, r, response.InternalError())
			return
		}

		err = validation.Validate(c)
		if err != nil {
			log.Error("bad request", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("Bad request", err))
			return
		}

		// use cost <= bcrypt.DefaultCost
		hash, err := secrets.GenerateFromPassword(c.Password)
		if err != nil {
			log.Error("failed to create password hash", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("Cannot create user"))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), s.Options.WriteTimeout)
		defer cancel()

		userID, err := s.SaveUser(ctx, c.Email, hash)
		if err != nil {
			log.Error("failed to register the user", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("Cannot create user"))
			return
		}

		response := response.Ok(fmt.Sprintf("User successfully created with ID: %d", userID))
		codec.JSONResponse(w, r, response)
	}

	return http.HandlerFunc(handler)
}
