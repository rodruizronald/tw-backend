package httpservice

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Timestamp string            `json:"timestamp" example:"2024-01-15T10:30:45Z"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	db *pgxpool.Pool
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// RegisterRoutes registers health routes (satisfies HTTPHandler interface)
func (h *HealthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.HandleHealth)
}

// HandleHealth godoc
// @Summary Health check endpoint
// @Description Returns the health status of the service and its dependencies
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} HealthResponse "Service is unhealthy"
// @Router /health [get]
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	status := "healthy"
	checks["database"] = "healthy"
	httpStatus := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		status = "unhealthy"
		checks["database"] = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	c.JSON(httpStatus, response)
}
