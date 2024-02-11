package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/http-server/handlers"
	"github.com/studopolis/auth-server/internal/lib/http/cors"
	"github.com/studopolis/auth-server/internal/lib/jwt"
	"github.com/studopolis/auth-server/internal/lib/logger"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	jwtMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/jwt"
	logMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/logger"
	requestMiddleware "github.com/studopolis/auth-server/internal/http-server/middleware/request"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Description:")
		fmt.Fprintln(flag.CommandLine.Output(), "   - Studopolis Authentication Server")
		fmt.Fprintf(flag.CommandLine.Output(), "   - https://github.com/studopolis/auth-server\n\n")
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(flag.CommandLine.Output(), "   --%-14s %s\n", f.Name, f.Usage)
		})
	}

	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config YAML file (development only)")
	flag.Parse()

	// Config and Logger setup
	config := config.MustLoad(configPath)
	log := logger.New(config.Stage)

	log.Info("starting auth service", logger.Stage(config.Stage))
	log.Debug("debug messages are enabled")

	// Storage setup
	storage, err := storage.New(context.Background(), config.Storage)
	if err != nil {
		log.Error("failed to initialize storage", logger.Error(err))
		os.Exit(1)
	}

	// Router setup
	router := handlers.NewRouter()

	// CORS
	cors := cors.New(config.CORS)
	router.Use(cors)

	// Request ID middleware
	requestMiddleware := requestMiddleware.New()
	router.Use(requestMiddleware)

	// Logger middleware
	logMiddleware := logMiddleware.New(log)
	router.Use(logMiddleware)

	// JWT: service
	jwtService := jwt.NewService(config.JWT)

	// JWT: middleware
	jwtMiddleware := jwtMiddleware.New(log, jwtService, storage)

	// Public handlers
	publicRouter := router.PathPrefix("/").Subrouter()
	handlers.Public(publicRouter, log, jwtService, storage)

	// Protected handlers
	protectedRouter := router.PathPrefix("/").Subrouter()
	protectedRouter.Use(jwtMiddleware)
	handlers.Protected(protectedRouter, log, storage)

	// Server setup
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

	healthCheckPassed := make(chan bool, 1)
	go func() {
		time.Sleep(config.HTTPServer.HealthTimeout)

		_, err := net.Dial("tcp", server.Addr)
		if err != nil {
			log.Error("server health check failed", logger.Error(err))
			healthCheckPassed <- false
		}
		healthCheckPassed <- true
	}()

	select {
	case <-shutdown:
	case passed := <-healthCheckPassed:
		if passed {
			log.Info("server started")
			<-shutdown
		}
	}

	log.Info("stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), config.HTTPServer.ShutdownTimeout)
	defer cancel()

	// Storage closing
	storage.Stop()

	// Server shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", logger.Error(err))
		return
	} else {
		log.Info("server stopped")
	}
}
