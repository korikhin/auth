package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/studopolis/auth-server/internal/lib/logger"

	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	log.Info("logger middleware enabled")
	log = log.With(
		logger.Component("middleware/logger"),
	)

	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			log := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				logger.RequestID(requestMiddleware.GetID(r.Context())),
			)

			log.Info(
				"starting",
			)

			tic := time.Now()
			defer func() {
				log.Info(
					"completed",
					logger.Duration(time.Since(tic)),
				)
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handler)
	}
}
