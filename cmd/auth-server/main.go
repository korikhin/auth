package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/http-server/handlers"
	"github.com/studopolis/auth-server/internal/lib/logger"

	jwtMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/jwt"
	logMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/logger"
	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"

	"github.com/gorilla/mux"
)

func main() {
	// config and logger setup
	config := config.MustLoad()
	log := logger.New(config.Env)

	log.Info("starting auth service", slog.String("env", string(config.Env)))
	log.Debug("debug messages are enabled")

	// todo: storage setup
	// ...

	// router setup
	router := mux.NewRouter()
	router.Use(requestMiddleware.RequestID)

	logMiddleware := logMiddleware.New(log)
	router.Use(logMiddleware)

	// handlers: public
	publicRouter := router.PathPrefix("/").Subrouter()
	handlers.Public(publicRouter, log)

	// handlers: protected
	protectedRouter := router.PathPrefix("/").Subrouter()

	jwtMiddleware := jwtMiddleware.New(log)
	protectedRouter.Use(jwtMiddleware)

	handlers.Protected(protectedRouter, log)

	// server setup
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("starting server")

	server := &http.Server{
		Addr:         config.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  config.HTTPServer.ReadTimeout,
		WriteTimeout: config.HTTPServer.WriteTimeout,
		IdleTimeout:  config.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-shutdown
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server")
		return
	}

	log.Info("server stopped")
}
