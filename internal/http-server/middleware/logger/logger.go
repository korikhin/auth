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
		slog.String("component", "http/logger"),
	)

	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String(logger.RequestIDAttr, requestMiddleware.GetID(r.Context())),
			)

			entry.Info(
				"starting",
			)

			tic := time.Now()
			defer func() {
				entry.Info(
					"completed",
					slog.String("duration", time.Since(tic).String()),
				)
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handler)
	}
}
