package handlers

import (
	"log/slog"
	"net/http"

	"github.com/studopolis/auth-server/internal/http-server/handlers/authn"
	"github.com/studopolis/auth-server/internal/http-server/handlers/health"
	"github.com/studopolis/auth-server/internal/http-server/handlers/login"
	"github.com/studopolis/auth-server/internal/http-server/handlers/register"
	"github.com/studopolis/auth-server/internal/lib/jwt"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	jwtMW "github.com/studopolis/auth-server/internal/http-server/middleware/jwt"
	reqMW "github.com/studopolis/auth-server/internal/http-server/middleware/request"

	"github.com/gorilla/mux"
)

// TODO: replace with net/http someday
func NewRouter() *mux.Router {
	return mux.NewRouter().PathPrefix("/api").Subrouter()
}

func Public(r *mux.Router, log *slog.Logger, a *jwt.JWTService, s *storage.Storage) {
	p := r.PathPrefix("/").Subrouter()

	// MWs
	empMW := reqMW.NotEmpty(log)

	health := health.New()
	p.Handle("/v1/health", health)

	register := register.New(log, s)
	p.Handle("/v1/users", empMW(register)).Methods(http.MethodPost)

	login := login.New(log, a, s)
	p.Handle("/v1/auth", empMW(login)).Methods(http.MethodPost)
}

func Protected(r *mux.Router, log *slog.Logger, a *jwt.JWTService, s *storage.Storage) {
	p := r.PathPrefix("/").Subrouter()

	// MWs
	// empMW := reqMW.NotEmpty(log)
	jwtMW := jwtMW.New(log, a, s)

	p.Use(jwtMW)

	authn := authn.New()
	p.Handle("/v1/auth", authn)

	// deleteUser := delete.New()
	// p.Handle("/v1/users/{id}", deleteUser).Methods(http.MethodDelete)
}
