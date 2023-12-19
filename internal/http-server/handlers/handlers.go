package handlers

import (
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/http-server/handlers/authenticate"
	"github.com/studopolis/auth-server/internal/http-server/handlers/login"
	"github.com/studopolis/auth-server/internal/http-server/handlers/register"
	"github.com/studopolis/auth-server/internal/http-server/handlers/test"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	"github.com/gorilla/mux"
)

func Public(r *mux.Router, log *slog.Logger, s *storage.Storage) {
	register := register.New(log)
	r.Handle("/users", register).Methods(http.MethodPost)

	login := login.New(log)
	r.Handle("/users", login).Methods(http.MethodGet)

	auth := authenticate.New()
	r.Handle("/auth", auth).Methods(http.MethodGet)
}

func Protected(r *mux.Router, log *slog.Logger, s *storage.Storage) {
	test := test.New(log, s)
	r.Handle("/test", test).Methods(http.MethodGet)
}
