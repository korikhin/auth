package request

import (
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/lib/logger"
)

func NotEmpty(log *slog.Logger) func(next http.Handler) http.Handler {
	log = log.With(logger.Component("middleware/request"))

	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			log := log.With(
				logger.RequestID(GetID(r.Context())),
			)

			if r.Method == http.MethodPost && (r.Body == nil || r.ContentLength == 0) {
				log.Error("request body is empty")
				http.Error(w, "Request body is empty", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handler)
	}
}
