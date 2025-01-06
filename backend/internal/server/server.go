package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ndn/backend/internal/config"
	"github.com/ndn/backend/internal/container"
	"github.com/ndn/backend/internal/handlers"
	"github.com/ndn/backend/internal/routes"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
)

type Server struct {
	router *chi.Mux
	logger *zap.Logger
	nrApp  *newrelic.Application
	config *config.Config
	server *http.Server
}

// New creates a new server instance with all dependencies
func New() (*Server, error) {
	// Initialize container with all dependencies
	c := container.BuildContainer()

	// Get dependencies from container
	var (
		cfg    *config.Config
		logger *zap.Logger
		nrApp  *newrelic.Application
	)

	if err := c.Invoke(func(
		c *config.Config,
		l *zap.Logger,
		nr *newrelic.Application,
	) {
		cfg = c
		logger = l
		nrApp = nr
	}); err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %v", err)
	}

	// Get handlers
	var (
		authHandler     *handlers.AuthHandler
		movieHandler    *handlers.MovieHandler
		categoryHandler *handlers.CategoryHandler
		userHandler     *handlers.UserHandler
	)

	if err := c.Invoke(func(
		ah *handlers.AuthHandler, mh *handlers.MovieHandler, ch *handlers.CategoryHandler, uh *handlers.UserHandler) {
		authHandler = ah
		movieHandler = mh
		categoryHandler = ch
		userHandler = uh
	}); err != nil {
		return nil, fmt.Errorf("failed to get handlers: %v", err)
	}

	// Setup routes
	router := routes.SetupRoutes(
		authHandler,
		movieHandler,
		categoryHandler,
		userHandler,
	)

	// Create server instance
	srv := &Server{
		router: router,
		logger: logger,
		nrApp:  nrApp,
		config: cfg,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	return srv, nil
}

// Start begins serving the HTTP server and handles graceful shutdown
func (s *Server) Start() error {
	// Start server
	go func() {
		s.logger.Info("server starting", zap.String("port", s.config.Server.Port))
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("server is shutting down...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %v", err)
	}

	s.logger.Info("server exited properly")
	return nil
}
