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
	"github.com/studopolis/auth-server/internal/lib/jwt"
	"github.com/studopolis/auth-server/internal/lib/logger"
	storage "github.com/studopolis/auth-server/internal/storage/postgres"

	logMW "github.com/studopolis/auth-server/internal/http-server/middleware/logger"
	reqMW "github.com/studopolis/auth-server/internal/http-server/middleware/request"
	corMW "github.com/studopolis/auth-server/internal/lib/http/cors"
)

func usage() {
	w := flag.CommandLine.Output()
	_, _ = fmt.Fprintln(w, "Authentication Server\nFlags:")

	flag.VisitAll(func(f *flag.Flag) {
		_, _ = fmt.Fprintf(w, "  --%-15s %s (default: %q)\n", f.Name, f.Usage, f.DefValue)
	})
}

func main() {
	flag.Usage = usage

	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.Parse()

	// Config and Logger setup
	config := config.MustLoad(configPath)
	log := logger.New(config.Stage)

	log.Info("starting auth service...", logger.Stage(config.Stage))
	log.Debug("debug messages are enabled")

	// Storage setup
	storage, err := storage.New(context.Background(), config.Storage)
	if err != nil {
		log.Error("failed to initialize storage", logger.Error(err))
		os.Exit(1)
	}

	corMW := corMW.New(config.CORS)
	ridMW := reqMW.ID()
	logMW := logMW.New(log)

	// Router setup
	router := handlers.NewRouter()
	router.Use(corMW, ridMW, logMW)

	jwtService := jwt.NewService(config.JWT)

	handlers.Public(router, log, jwtService, storage)
	handlers.Protected(router, log, jwtService, storage)

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
	log.Info("server started")

	healthCheckTerminate := make(chan struct{}, 1)
	go func() {
		log := log.With(logger.Component("system/health"))

		ticker := time.NewTicker(config.HTTPServer.HealthTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-healthCheckTerminate:
				return
			case <-ticker.C:
				conn, err := net.Dial("tcp", server.Addr)
				if err != nil {
					log.Error("server health check failed", logger.Error(err))
					continue
				}
				conn.Close()
			}

			// Health check is successful
			log.Info("")
		}
	}()

	shutdownSignal := <-shutdown
	log.Info("recieved shutdown signal", logger.Signal(shutdownSignal))

	healthCheckTerminate <- struct{}{}
	log.Info("stopping server...")

	// Storage closing
	storage.Stop()

	// Server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", logger.Error(err))
		return
	}

	log.Info("server stopped successfully")
}
