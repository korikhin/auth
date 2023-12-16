package handlers

import (
	"log/slog"
	"net/http"

	authenticate "github.com/studopolis/auth-server/internal/http-server/handlers/authenticate"
	login "github.com/studopolis/auth-server/internal/http-server/handlers/login"
	register "github.com/studopolis/auth-server/internal/http-server/handlers/register"
	test "github.com/studopolis/auth-server/internal/http-server/handlers/test"

	"github.com/gorilla/mux"
)

// todo: add storage
func Public(r *mux.Router, log *slog.Logger) {
	register := register.New(log)
	r.Handle("/users", register).Methods(http.MethodPost)

	login := login.New(log)
	r.Handle("/users", login).Methods(http.MethodGet)

	auth := authenticate.New(log)
	r.Handle("/auth", auth).Methods(http.MethodGet)
}

func Protected(r *mux.Router, log *slog.Logger) {
	test := test.New(log)
	r.Handle("/test", test).Methods(http.MethodGet)
}
