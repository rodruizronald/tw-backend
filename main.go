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
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/rodruizronald/tw-backend/docs"
	"github.com/rodruizronald/tw-backend/internal/company"
	"github.com/rodruizronald/tw-backend/internal/config"
	"github.com/rodruizronald/tw-backend/internal/database"
	"github.com/rodruizronald/tw-backend/internal/devmocks"
	"github.com/rodruizronald/tw-backend/internal/httpservice"
	"github.com/rodruizronald/tw-backend/internal/jobs"
	"github.com/rodruizronald/tw-backend/internal/jobtech"
	"github.com/rodruizronald/tw-backend/internal/logger"
	"github.com/rodruizronald/tw-backend/internal/router"
	"github.com/rodruizronald/tw-backend/internal/server"
)

const (
	exitWithError      = 1
	exitedSuccessfully = 0
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	code := run(ctx)

	stop()
	os.Exit(code)
}

func run(ctx context.Context) int {
	// Load configuration
	cfg, err := config.Load(".env")
	if err != nil {
		fmt.Printf("failed to load configuration: %v", err)
		return exitWithError
	}
	log := logger.New(&cfg.Logger)

	dbpool, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		log.Errorf("unable to connect to database: %v", err)
		return exitWithError
	}
	defer dbpool.Close()

	// Setup repositories based on Gin mode
	jobRepo, companyRepo := setupRepositories(cfg.Gin.Mode, dbpool)

	// Create handlers
	jobHandler := jobs.NewHandler(jobRepo)
	companyHandler := company.NewHandler(companyRepo)
	healthHandler := httpservice.NewHealthHandler(dbpool)

	// Create router with handlers
	appRouter := router.New(jobHandler, companyHandler, healthHandler, log)
	r := appRouter.Setup(&cfg.Gin)

	// Create and start server
	srv := server.New(&cfg.Server, r, log)
	if err := srv.Start(ctx); err != nil {
		log.Errorf("Application error: %v", err)
		return exitWithError
	}

	return exitedSuccessfully
}

// setupRepositories creates all repositories based on Gin mode
func setupRepositories(mode string, dbpool *pgxpool.Pool) (
	jobDataRepo jobs.DataRepository,
	companyDataRepo company.DataRepository,
) {
	// Use mock repositories in test mode
	if mode == gin.TestMode {
		return devmocks.NewJobRepository(), devmocks.NewCompanyRepository()
	}

	// Create repositories
	jobRepo := jobs.NewRepository(dbpool)
	jobtechRepo := jobtech.NewRepository(dbpool)
	jobDataRepo = jobs.NewRepositories(jobRepo, jobtechRepo)
	companyDataRepo = company.NewRepository(dbpool)

	return jobDataRepo, companyDataRepo
}
