// Package server provides HTTP server management with graceful shutdown capabilities.
// It wraps the standard http.Server with structured logging and handles startup/shutdown lifecycle.
package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/rodruizronald/tw-backend/internal/config"
	"github.com/rodruizronald/tw-backend/internal/logger"
)

const (
	componentName = "server"
)

// Server represents the HTTP server with its dependencies
type Server struct {
	srv    *http.Server
	logger logrus.FieldLogger
}

// New creates a new Server instance
func New(cfg *config.ServerConfig, router *gin.Engine, log *logrus.Logger) *Server {
	return &Server{
		logger: logger.WithComponent(log, componentName),
		srv: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: router,
		},
	}
}

// Start starts the HTTP server and handles graceful shutdown
func (s *Server) Start(ctx context.Context) error {
	// Create error group with context
	g, gCtx := errgroup.WithContext(ctx)

	// Start HTTP server in goroutine
	g.Go(func() error {
		// Extract port from Addr (remove the ":" prefix)
		port := strings.TrimPrefix(s.srv.Addr, ":")

		s.logger.Printf("Server starting on port %s", port)
		s.logger.Printf("Swagger UI available at: http://localhost:%s/swagger/index.html", port)

		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("Server failed to start: %v", err)
			return err
		}
		return nil
	})

	// Handle graceful shutdown in another goroutine
	g.Go(func() error {
		<-gCtx.Done() // Wait for context cancellation (SIGINT/SIGTERM)

		s.logger.Println("Shutting down server...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Printf("Server forced to shutdown: %v", err)
			return err
		}

		s.logger.Println("Server exited gracefully")
		return nil
	})

	// Wait for all goroutines to complete
	return g.Wait()
}
