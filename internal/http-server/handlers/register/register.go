package register

import (
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/lib/http/codec"
	"github.com/studopolis/auth-server/internal/lib/logger"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.register.New"

		log := log.With(
			logger.Operation(op),
			logger.RequestID(requestMiddleware.GetID(r.Context())),
		)

		log.Info("register handler")
		response := map[string]string{"message": "register"}
		codec.JSONResponse(w, r, response)
	}

	return http.HandlerFunc(handler)
}
