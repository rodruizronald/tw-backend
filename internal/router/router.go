// Package router provides HTTP routing configuration and middleware setup for the job board API.
// It handles CORS, Swagger documentation, and API route registration using the Gin web framework.
package router

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/rodruizronald/ticos-in-tech/internal/company"
	"github.com/rodruizronald/ticos-in-tech/internal/config"
	"github.com/rodruizronald/ticos-in-tech/internal/jobs"
	"github.com/rodruizronald/ticos-in-tech/internal/logger"
)

const (
	componentName = "router"
)

// Router handles HTTP routing and middleware setup
type Router struct {
	jobRepo     jobs.DataRepository
	companyRepo company.DataRepository
	logger      logrus.FieldLogger
}

// New creates a new Router instance with dependencies
func New(jobRepo jobs.DataRepository, companyRepo company.DataRepository, log *logrus.Logger) *Router {
	return &Router{
		jobRepo:     jobRepo,
		companyRepo: companyRepo,
		logger:      logger.WithComponent(log, componentName),
	}
}

// Setup configures and returns a Gin engine with all middleware and routes
func (r *Router) Setup(cfg *config.GinConfig) *gin.Engine {
	gin.SetMode(cfg.Mode)
	engine := gin.Default()

	// Log the Gin mode being used
	r.logger.Infof("Gin engine configured in %s mode", strings.ToUpper(gin.Mode()))

	r.setupMiddleware(engine)
	r.setupRoutes(engine)

	return engine
}

// setupMiddleware configures all middleware
func (r *Router) setupMiddleware(engine *gin.Engine) {
	// Add CORS middleware
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // React app URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "X-Request-Timestamp"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes(engine *gin.Engine) {
	// Swagger routes
	r.setupSwaggerRoutes(engine)

	// API routes
	r.setupAPIRoutes(engine)
}

// setupSwaggerRoutes configures Swagger documentation routes
func (r *Router) setupSwaggerRoutes(engine *gin.Engine) {
	if gin.Mode() != gin.ReleaseMode {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

// setupAPIRoutes configures all API routes
func (r *Router) setupAPIRoutes(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")

	// Job routes
	jobHandler := jobs.NewHandler(r.jobRepo)
	jobHandler.RegisterRoutes(v1)

	// Company routes
	companyHandler := company.NewHandler(r.companyRepo)
	companyHandler.RegisterRoutes(v1)
}
