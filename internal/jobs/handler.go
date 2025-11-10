package jobs

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/rodruizronald/tw-backend/internal/httpservice"
	"github.com/rodruizronald/tw-backend/internal/jobtech"
)

// Constants for job routes and endpoints
const (
	JobsRoute      = "/jobs"
	JobsCountRoute = "/jobs/count"
)

// DataRepository interface to make database operations for the Job model.
type DataRepository interface {
	GetSearchCount(ctx context.Context, params *SearchParams) (int, error)
	SearchJobsWithCount(ctx context.Context, params *SearchParams) ([]*JobWithCompany, int, error)
	GetJobTechnologiesBatch(ctx context.Context, jobIDs []int) (map[int][]*jobtech.JobTechnologyWithDetails, error)
}

// Repositories struct to hold repositories for job and jobtech models
type Repositories struct {
	jobRepo     *Repository
	jobtechRepo *jobtech.Repository
}

// GetSearchCount delegates to the job repository's GetSearchCount method
func (r *Repositories) GetSearchCount(ctx context.Context, params *SearchParams) (int, error) {
	return r.jobRepo.GetSearchCount(ctx, params)
}

// SearchJobsWithCount delegates to the job repository's SearchJobsWithCount method
func (r *Repositories) SearchJobsWithCount(ctx context.Context, params *SearchParams) ([]*JobWithCompany, int, error) {
	return r.jobRepo.SearchJobsWithCount(ctx, params)
}

// GetJobTechnologiesBatch delegates to the jobtech repository's GetJobTechnologiesBatch method
func (r *Repositories) GetJobTechnologiesBatch(ctx context.Context, jobIDs []int) (
	map[int][]*jobtech.JobTechnologyWithDetails, error) {
	return r.jobtechRepo.GetJobTechnologiesBatch(ctx, jobIDs)
}

// Handler handles HTTP requests for job operations using the generic httpservice
type Handler struct {
	searchHandler *httpservice.SearchHandler[*SearchRequest, *SearchParams, JobResponseList]
}

// NewRepositories creates a new job and jobtech repositories
func NewRepositories(jobRepo *Repository, jobtechRepo *jobtech.Repository) *Repositories {
	return &Repositories{jobRepo: jobRepo, jobtechRepo: jobtechRepo}
}

// NewHandler creates a new job handler using httpservice.NewSearchHandlerWithDefaults
func NewHandler(repos DataRepository) *Handler {
	// Create the search service
	searchService := NewSearchService(repos)

	// Create the generic search handler with defaults
	searchHandler := httpservice.NewSearchHandlerWithDefaults(
		func() *SearchRequest { return &SearchRequest{} }, // Request factory function
		searchService,
	)

	return &Handler{
		searchHandler: searchHandler,
	}
}

// RegisterRoutes registers job routes with the given router group
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET(JobsRoute, h.SearchJobs)
	rg.GET(JobsCountRoute, h.SearchJobsCount)
}

// SearchJobs godoc
// @Summary Search for jobs
// @Description Search for jobs with optional filters and pagination
// @Tags jobs
// @Accept json
// @Produce json
// @Param q query string true "Search query" example("golang developer")
// @Param limit query int false "Number of results to return (max 100)" default(20) example(20)
// @Param offset query int false "Number of results to skip" default(0) example(0)
// @Param experience query string false "Experience level filter" \
// Enums(entry-level,mid-level,senior,manager,director,executive) example("senior")
// @Param type query string false "Employment type filter" \
// Enums(full-time,part-time,contractor,temporary,internship) example("full-time")
// @Param location query string false "Location filter" Enums(costarica,latam) example("costarica")
// @Param mode query string false "Work mode filter" Enums(remote,hybrid,onsite) example("remote")
// @Param company query string false "Company name filter (partial match)" example("Tech Corp")
// @Param date_from query string false "Start date filter (YYYY-MM-DD)" example("2024-01-01")
// @Param date_to query string false "End date filter (YYYY-MM-DD)" example("2024-12-31")
// @Success 200 {object} SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs [get]
func (h *Handler) SearchJobs(c *gin.Context) { h.searchHandler.HandleSearch(c) }

// SearchJobsCount godoc
// @Summary Get count of jobs matching search criteria
// @Description Get the total count of jobs that match the search criteria without returning the actual job data
// @Tags jobs
// @Accept json
// @Produce json
// @Param q query string true "Search query" example("golang developer")
// @Param experience query string false "Experience level filter" \
// Enums(entry-level,mid-level,senior,manager,director,executive) example("senior")
// @Param type query string false "Employment type filter" \
// Enums(full-time,part-time,contractor,temporary,internship) example("full-time")
// @Param location query string false "Location filter" Enums(costarica,latam) example("costarica")
// @Param mode query string false "Work mode filter" Enums(remote,hybrid,onsite) example("remote")
// @Param company query string false "Company name filter (partial match)" example("Tech Corp")
// @Param date_from query string false "Start date filter (YYYY-MM-DD)" example("2024-01-01")
// @Param date_to query string false "End date filter (YYYY-MM-DD)" example("2024-12-31")
// @Success 200 {object} map[string]int "Returns count in format: {"count": 42}"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs/count [get]
func (h *Handler) SearchJobsCount(c *gin.Context) { h.searchHandler.HandleSearchCount(c) }
