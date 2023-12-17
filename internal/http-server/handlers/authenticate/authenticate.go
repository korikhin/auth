package authenticate

import (
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/lib/http/responder"
	"github.com/studopolis/auth-server/internal/lib/logger"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.authenticate.New"

		log := log.With(
			logger.Operand(op),
			logger.RequestID(requestMiddleware.GetID(r.Context())),
		)

		log.Info("auth handler")
		response := map[string]string{"message": "auth"}
		responder.JSON(w, r, response)
	}

	return http.HandlerFunc(handler)
}
