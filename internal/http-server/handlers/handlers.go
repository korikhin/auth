package handlers

import (
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/http-server/handlers/authenticate"
	"github.com/studopolis/auth-server/internal/http-server/handlers/health"
	"github.com/studopolis/auth-server/internal/http-server/handlers/login"
	"github.com/studopolis/auth-server/internal/http-server/handlers/register"
	"github.com/studopolis/auth-server/internal/lib/jwt"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	return mux.NewRouter().PathPrefix("/api").Subrouter()
}

func Public(r *mux.Router, log *slog.Logger, a *jwt.JWTService, s *storage.Storage) {
	health := health.New()
	r.Handle("/v1/health", health).Methods(http.MethodGet)

	register := register.New(log, s)
	r.Handle("/v1/users", register).Methods(http.MethodPost)

	login := login.New(log, a, s)
	r.Handle("/v1/auth", login).Methods(http.MethodPost)
}

func Protected(r *mux.Router, log *slog.Logger, s *storage.Storage) {
	auth := authenticate.New()
	r.Handle("/v1/auth", auth).Methods(http.MethodGet)
}
