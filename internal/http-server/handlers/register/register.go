package register

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/korikhin/auth/internal/lib/api"
	"github.com/korikhin/auth/internal/lib/http/codec"
	"github.com/korikhin/auth/internal/lib/logger"
	storage "github.com/korikhin/auth/internal/storage/postgres"

	reqMW "github.com/korikhin/auth/internal/http-server/middleware/request"

	"golang.org/x/crypto/bcrypt"
)

const hashCost = 7

var (
	errCannotCreateUser = api.Error("cannot create user")
)

func New(log *slog.Logger, s *storage.Storage) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.register.New"

		log := log.With(
			logger.Operation(op),
			logger.RequestID(reqMW.GetID(r.Context())),
		)

		c := &api.Credentials{}

		err := codec.DecodeJSON(r.Body, c)
		if err != nil {
			log.Error("failed to decode request body", logger.Error(err))
			codec.ResponseJSON(w, api.InternalError, http.StatusInternalServerError)
			return
		}

		err = api.Validate(c)
		if err != nil {
			log.Error("bad request", logger.Error(err))
			codec.ResponseJSON(w, api.Error("bad request", err), http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), hashCost)
		if err != nil {
			log.Error("failed to create password hash", logger.Error(err))
			codec.ResponseJSON(w, errCannotCreateUser, http.StatusInternalServerError)
			return
		}

		ctxStorage, cancel := context.WithTimeout(context.Background(), s.Options.WriteTimeout)
		defer cancel()

		userID, err := s.SaveUser(ctxStorage, c.Email, hash)
		if err != nil {
			log.Error("failed to register the user", logger.Error(err))
			codec.ResponseJSON(w, errCannotCreateUser, http.StatusInternalServerError)
			return
		}

		response := api.Ok(fmt.Sprintf("user successfully registered: %v", userID))
		codec.ResponseJSON(w, response, http.StatusCreated)
	}

	return http.HandlerFunc(handler)
}
