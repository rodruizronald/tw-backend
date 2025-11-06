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
	"github.com/rodruizronald/ticos-in-tech/internal/company"
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

// setupCompanyRepositories creates the appropriate company repositories based on Gin mode
func setupCompanyRepositories(ctx context.Context, cfg *config.Config) (company.DataRepository, func(), error) {
	if cfg.Gin.Mode == gin.TestMode {
		return devmocks.NewCompanyRepository(), func() {}, nil
	}

	// Connect to the database using config
	dbpool, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	companyRepo := company.NewRepository(dbpool)

	return companyRepo, func() { dbpool.Close() }, nil
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
	jobRepos, jobCleanup, err := setupJobRepositories(ctx, cfg)
	if err != nil {
		log.Errorf("Failed to setup job repositories: %v", err)
		return exitWithError
	}
	defer jobCleanup()

	// Setup company repositories
	companyRepos, companyCleanup, err := setupCompanyRepositories(ctx, cfg)
	if err != nil {
		log.Errorf("Failed to setup company repositories: %v", err)
		return exitWithError
	}
	defer companyCleanup()

	// Create router
	appRouter := router.New(jobRepos, companyRepos, log)
	r := appRouter.Setup(&cfg.Gin)

	// Create and start server
	srv := server.New(&cfg.Server, r, log)
	if err := srv.Start(ctx); err != nil {
		log.Errorf("Application error: %v", err)
		return exitWithError
	}

	return exitedSuccessfully
}
