package register

import (
	"context"
	"fmt"
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

var (
	errCannotCreateUser = response.Error("cannot create user", http.StatusInternalServerError)
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
			log.Error("failed to decode request body", logger.Error(err))
			codec.JSONResponse(w, r, response.InternalError)
			return
		}

		err = validation.Validate(c)
		if err != nil {
			log.Error("bad request", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("bad request", http.StatusBadRequest, err))
			return
		}

		hash, err := secrets.GenerateFromPassword(c.Password)
		if err != nil {
			log.Error("failed to create password hash", logger.Error(err))
			codec.JSONResponse(w, r, errCannotCreateUser)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), s.Options.WriteTimeout)
		defer cancel()

		userID, err := s.SaveUser(ctx, c.Email, hash)
		if err != nil {
			log.Error("failed to register the user", logger.Error(err))
			codec.JSONResponse(w, r, errCannotCreateUser)
			return
		}

		response := response.Ok(
			fmt.Sprintf("user successfully registered: %v", userID),
			http.StatusCreated,
		)
		codec.JSONResponse(w, r, response)
	}

	return http.HandlerFunc(handler)
}
