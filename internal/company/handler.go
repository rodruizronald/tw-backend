package company

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rodruizronald/tw-backend/internal/httpservice"
)

// Constants for company routes
const (
	CompaniesRoute = "/companies"
	CompanyRoute   = "/companies/:id"
)

// DataRepository interface defines the repository methods needed by the handler
type DataRepository interface {
	Create(ctx context.Context, company *Company) error
	Update(ctx context.Context, company *Company) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*Company, error)
}

// Handler handles HTTP requests for company CRUD operations
type Handler struct {
	repo DataRepository
}

// NewHandler creates a new company handler instance
func NewHandler(repo DataRepository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes registers company routes with the given router group
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST(CompaniesRoute, h.CreateCompany)
	rg.GET(CompaniesRoute, h.ListCompanies)
	rg.PUT(CompanyRoute, h.UpdateCompany)
	rg.DELETE(CompanyRoute, h.DeleteCompany)
}

// CreateCompanyRequest represents the request body for creating a company
type CreateCompanyRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

// UpdateCompanyRequest represents the request body for updating a company
type UpdateCompanyRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive *bool  `json:"is_active" binding:"required"`
}

// CreateCompany godoc
// @Summary Create a new company
// @Description Create a new company with the provided name and active status
// @Tags companies
// @Accept json
// @Produce json
// @Param company body CreateCompanyRequest true "Company creation request"
// @Success 201 {object} Company
// @Failure 400 {object} httpservice.ErrorResponse
// @Failure 409 {object} httpservice.ErrorResponse
// @Failure 500 {object} httpservice.ErrorResponse
// @Router /companies [post]
func (h *Handler) CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    httpservice.ErrCodeInvalidRequest,
				Message: "Invalid request body",
				Details: []string{err.Error()},
			},
		})
		return
	}

	// Set default value for IsActive if not provided
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	company := &Company{
		Name:     req.Name,
		IsActive: isActive,
	}

	if err := h.repo.Create(c.Request.Context(), company); err != nil {
		statusCode, errorResp := h.buildErrorResponse(err)
		c.JSON(statusCode, errorResp)
		return
	}

	c.JSON(http.StatusCreated, company)
}

// ListCompanies godoc
// @Summary List all companies
// @Description Retrieve all companies from the database
// @Tags companies
// @Accept json
// @Produce json
// @Success 200 {array} Company
// @Failure 500 {object} httpservice.ErrorResponse
// @Router /companies [get]
func (h *Handler) ListCompanies(c *gin.Context) {
	companies, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    httpservice.ErrCodeInternalError,
				Message: "Failed to retrieve companies",
				Details: []string{err.Error()},
			},
		})
		return
	}

	// Return empty array instead of null if no companies exist
	if companies == nil {
		companies = []*Company{}
	}

	c.JSON(http.StatusOK, companies)
}

// UpdateCompany godoc
// @Summary Update a company
// @Description Update an existing company's name and active status
// @Tags companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Param company body UpdateCompanyRequest true "Company update request"
// @Success 200 {object} Company
// @Failure 400 {object} httpservice.ErrorResponse
// @Failure 404 {object} httpservice.ErrorResponse
// @Failure 409 {object} httpservice.ErrorResponse
// @Failure 500 {object} httpservice.ErrorResponse
// @Router /companies/{id} [put]
func (h *Handler) UpdateCompany(c *gin.Context) {
	// Parse company ID from URL
	var uriParam struct {
		ID int `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    httpservice.ErrCodeInvalidRequest,
				Message: "Invalid company ID",
				Details: []string{err.Error()},
			},
		})
		return
	}

	// Parse request body
	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    httpservice.ErrCodeInvalidRequest,
				Message: "Invalid request body",
				Details: []string{err.Error()},
			},
		})
		return
	}

	company := &Company{
		ID:       uriParam.ID,
		Name:     req.Name,
		IsActive: *req.IsActive,
	}

	if err := h.repo.Update(c.Request.Context(), company); err != nil {
		statusCode, errorResp := h.buildErrorResponse(err)
		c.JSON(statusCode, errorResp)
		return
	}

	c.JSON(http.StatusOK, company)
}

// DeleteCompany godoc
// @Summary Delete a company
// @Description Delete a company by its ID
// @Tags companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpservice.ErrorResponse
// @Failure 404 {object} httpservice.ErrorResponse
// @Failure 500 {object} httpservice.ErrorResponse
// @Router /companies/{id} [delete]
func (h *Handler) DeleteCompany(c *gin.Context) {
	// Parse company ID from URL
	var uriParam struct {
		ID int `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    httpservice.ErrCodeInvalidRequest,
				Message: "Invalid company ID",
				Details: []string{err.Error()},
			},
		})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), uriParam.ID); err != nil {
		statusCode, errorResp := h.buildErrorResponse(err)
		c.JSON(statusCode, errorResp)
		return
	}

	c.Status(http.StatusNoContent)
}

// buildErrorResponse builds an appropriate error response based on the error type
func (h *Handler) buildErrorResponse(err error) (int, httpservice.ErrorResponse) {
	switch {
	case IsNotFound(err):
		return http.StatusNotFound, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    "COMPANY_NOT_FOUND",
				Message: "Company not found",
				Details: []string{err.Error()},
			},
		}
	case IsDuplicate(err):
		return http.StatusConflict, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    "COMPANY_DUPLICATE",
				Message: "Company already exists",
				Details: []string{err.Error()},
			},
		}
	default:
		return http.StatusInternalServerError, httpservice.ErrorResponse{
			Error: httpservice.ErrorDetails{
				Code:    httpservice.ErrCodeInternalError,
				Message: "Internal server error",
				Details: []string{err.Error()},
			},
		}
	}
}
