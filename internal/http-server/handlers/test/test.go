package test

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func New(log *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.test.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", request.GetID(r.Context())),
		)

		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"message": "test"}

		log.Info("test handler")
		json.NewEncoder(w).Encode(response)
	}

	return http.HandlerFunc(handler)
}
