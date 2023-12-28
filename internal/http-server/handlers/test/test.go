package test

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/lib/api/response"
	"github.com/studopolis/auth-server/internal/lib/http/codec"
	"github.com/studopolis/auth-server/internal/lib/logger"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger, s *storage.Storage) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.test.New"

		log := log.With(
			logger.Operation(op),
			logger.RequestID(requestMiddleware.GetID(r.Context())),
		)

		err := s.Ping(context.Background())
		if err != nil {
			log.Error("failed to ping storage", logger.Error(err))
			codec.JSONResponse(w, r, response.Error("Failed to establish database connection"))
			return
		}

		response := response.Ok("Database connection established")
		codec.JSONResponse(w, r, response)
	}

	return http.HandlerFunc(handler)
}
