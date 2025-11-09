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

	"github.com/rodruizronald/ticos-in-tech/internal/config"
	"github.com/rodruizronald/ticos-in-tech/internal/logger"
)

const (
	componentName = "router"
)

// HTTPHandler defines the interface for HTTP handlers
type HTTPHandler interface {
	RegisterRoutes(rg *gin.RouterGroup)
}

// Router handles HTTP routing and middleware setup
type Router struct {
	jobHandler     HTTPHandler
	companyHandler HTTPHandler
	healthHandler  HTTPHandler
	logger         logrus.FieldLogger
}

// New creates a new Router instance with handlers
func New(
	jobHandler HTTPHandler,
	companyHandler HTTPHandler,
	healthHandler HTTPHandler,
	log *logrus.Logger,
) *Router {
	return &Router{
		jobHandler:     jobHandler,
		companyHandler: companyHandler,
		healthHandler:  healthHandler,
		logger:         logger.WithComponent(log, componentName),
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
	// Health check routes at root level (industry standard)
	r.setupHealthRoutes(engine)

	// Swagger routes
	r.setupSwaggerRoutes(engine)

	// API routes (versioned business logic)
	r.setupAPIRoutes(engine)
}

// setupHealthRoutes configures health check endpoints at root level
func (r *Router) setupHealthRoutes(engine *gin.Engine) {
	// Create a group at root level for health endpoints
	healthGroup := engine.Group("/health")
	r.healthHandler.RegisterRoutes(healthGroup)
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

	// Register business logic handlers
	r.jobHandler.RegisterRoutes(v1)
	r.companyHandler.RegisterRoutes(v1)
}
