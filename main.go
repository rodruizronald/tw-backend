// @title Job Board API
// @version 1.0
// @description A job board API for managing job postings
// @contact.name API Support
// @contact.email support@example.com
// @host localhost:8080
// @BasePath /api/v1
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	_ "github.com/rodruizronald/ticos-in-tech/docs"
	"github.com/rodruizronald/ticos-in-tech/internal/config"
	"github.com/rodruizronald/ticos-in-tech/internal/database"
	"github.com/rodruizronald/ticos-in-tech/internal/devmocks"
	"github.com/rodruizronald/ticos-in-tech/internal/jobs"
	"github.com/rodruizronald/ticos-in-tech/internal/jobtech"
	"github.com/rodruizronald/ticos-in-tech/internal/logger"
	"github.com/rodruizronald/ticos-in-tech/internal/router"
	"github.com/rodruizronald/ticos-in-tech/internal/server"
)

const (
	exitWithError      = 1
	exitedSuccessfully = 0
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Run application
	code := run(ctx)

	stop()
	os.Exit(code)
}

// setupJobRepositories creates the appropriate job repositories based on Gin mode
func setupJobRepositories(ctx context.Context, cfg *config.Config) (jobs.DataRepository, func(), error) {
	if cfg.Gin.Mode == gin.TestMode {
		return devmocks.NewJobRepository(), func() {}, nil
	}

	// Connect to the database using config
	dbpool, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	jobRepo := jobs.NewRepository(dbpool)
	jobtechRepo := jobtech.NewRepository(dbpool)
	jobRepos := jobs.NewRepositories(jobRepo, jobtechRepo)

	return jobRepos, func() { dbpool.Close() }, nil
}

func run(ctx context.Context) int {
	// Load configuration
	cfg, err := config.Load(".env")
	if err != nil {
		fmt.Printf("failed to load configuration: %v", err)
		return exitWithError
	}

	log := logger.New(&cfg.Logger)

	// Setup job repositories
	jobRepos, cleanup, err := setupJobRepositories(ctx, cfg)
	if err != nil {
		log.Errorf("Failed to setup repositories: %v", err)
		return exitWithError
	}
	defer cleanup()

	// Create router
	appRouter := router.New(jobRepos, log)
	r := appRouter.Setup(&cfg.Gin)

	// Create and start server
	srv := server.New(&cfg.Server, r, log)
	if err := srv.Start(ctx); err != nil {
		log.Errorf("Application error: %v", err)
		return exitWithError
	}

	return exitedSuccessfully
}
