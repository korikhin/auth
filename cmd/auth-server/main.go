package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/http-server/handlers"
	"github.com/studopolis/auth-server/internal/lib/logger"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

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

	// storage setup
	storage, err := storage.New(context.Background(), config.Storage)
	if err != nil {
		log.Error("failed to initialize storage", logger.Error(err))
		os.Exit(1)
	}

	// router setup
	router := mux.NewRouter()

	// middleware
	router.Use(requestMiddleware.RequestID)

	logMiddleware := logMiddleware.New(log)
	router.Use(logMiddleware)

	// handlers: public
	publicRouter := router.PathPrefix("/").Subrouter()
	handlers.Public(publicRouter, log, storage, *config)

	// handlers: protected
	protectedRouter := router.PathPrefix("/").Subrouter()

	jwtMiddleware := jwtMiddleware.New(log, storage, config.JWT)
	protectedRouter.Use(jwtMiddleware)

	handlers.Protected(protectedRouter, log, storage)

	// server setup
	server := &http.Server{
		Addr:         config.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  config.HTTPServer.ReadTimeout,
		WriteTimeout: config.HTTPServer.WriteTimeout,
		IdleTimeout:  config.HTTPServer.IdleTimeout,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("starting server...")
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("failed to start server", logger.Error(err))
			}
		}
	}()

	healthCheck := make(chan bool, 1)
	go func() {
		time.Sleep(config.HTTPServer.HealthTimeout)

		_, err := net.Dial("tcp", server.Addr)
		if err != nil {
			log.Error("server health check failed", logger.Error(err))
			healthCheck <- false
		}
		healthCheck <- true
	}()

	if <-healthCheck {
		log.Info("server started")
	}

	<-shutdown
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", logger.Error(err))
		return
	}

	log.Info("server stopped")
}
