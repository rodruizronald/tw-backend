package jobs

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rodruizronald/tw-backend/internal/httpservice"
)

// Constants for job attributes and values
const (
	// Experience levels
	experienceLevelEntry     = "entry-level"
	experienceLevelMid       = "mid-level"
	experienceLevelSenior    = "senior"
	experienceLevelManager   = "manager"
	experienceLevelDirector  = "director"
	experienceLevelExecutive = "executive"

	// Employment types
	employmentTypeFullTime   = "full-time"
	employmentTypePartTime   = "part-time"
	employmentTypeContractor = "contractor"
	employmentTypeTemporary  = "temporary"
	employmentTypeInternship = "internship"

	// Locations
	locationCostaRica = "costarica"
	locationLATAM     = "latam"

	// Work modes
	workModeRemote = "remote"
	workModeHybrid = "hybrid"
	workModeOnsite = "onsite"
)

// Validation collections for job attributes and values
var (
	validExperienceLevels = []string{
		experienceLevelEntry,
		experienceLevelMid,
		experienceLevelSenior,
		experienceLevelManager,
		experienceLevelDirector,
		experienceLevelExecutive,
	}
	validEmploymentTypes = []string{
		employmentTypeFullTime,
		employmentTypePartTime,
		employmentTypeContractor,
		employmentTypeTemporary,
		employmentTypeInternship,
	}
	validLocations = []string{
		locationCostaRica,
		locationLATAM,
	}
	validWorkModes = []string{
		workModeRemote,
		workModeHybrid,
		workModeOnsite,
	}
)

// Constants for search query validation limits
const (
	MaxQueryLength = 100 // Maximum characters for search query
	MinQueryLength = 2   // Minimum meaningful search length
)

// Data Transfer Objects (DTOs) for the job API layer.
// This file contains request/response structures used for HTTP API communication.
// These models define the external API contract and handle JSON serialization/deserialization.
// They are decoupled from database models to allow independent evolution of API and database schemas.

// SearchRequest represents the search request parameters (API layer)
type SearchRequest struct {
	Query           string `form:"q" binding:"required" example:"golang developer"`
	Limit           int    `form:"limit" example:"20"`
	Offset          int    `form:"offset" example:"0"`
	ExperienceLevel string `form:"experience" example:"Senior"`
	EmploymentType  string `form:"type" example:"Full-time"`
	Location        string `form:"location" example:"Costa Rica"`
	WorkMode        string `form:"mode" example:"Remote"`
	Company         string `form:"company" example:"Tech Corp"`
	DateFrom        string `form:"date_from" example:"2024-01-01"`
	DateTo          string `form:"date_to" example:"2024-12-31"`
}

// ToSearchParams converts a SearchRequest to SearchParams
func (req *SearchRequest) ToSearchParams() (httpservice.SearchParams, error) {
	// Set defaults for limit and offset
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	limit = min(limit, MaxLimit) // Max limit to prevent abuse

	offset := max(req.Offset, 0) // Min offset to prevent negative pagination

	searchParams := &SearchParams{
		Query:  req.Query,
		Limit:  limit,
		Offset: offset,
	}

	// Set optional filters
	if req.ExperienceLevel != "" {
		searchParams.ExperienceLevel = &req.ExperienceLevel
	}
	if req.EmploymentType != "" {
		searchParams.EmploymentType = &req.EmploymentType
	}
	if req.Location != "" {
		searchParams.Location = &req.Location
	}
	if req.WorkMode != "" {
		searchParams.WorkMode = &req.WorkMode
	}
	if req.Company != "" {
		searchParams.Company = &req.Company
	}

	// Parse dates if provided
	if req.DateFrom != "" && req.DateTo != "" {
		dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
		if err != nil {
			return nil, &httpservice.ConversionError{
				Field: "date_from",
				Value: req.DateFrom,
				Err:   err,
			}
		}
		dateTo, err := time.Parse("2006-01-02", req.DateTo)
		if err != nil {
			return nil, &httpservice.ConversionError{
				Field: "date_to",
				Value: req.DateTo,
				Err:   err,
			}
		}
		searchParams.DateFrom = &dateFrom
		searchParams.DateTo = &dateTo
	}

	return searchParams, nil
}

// Validate validates the search request parameters
func (req *SearchRequest) Validate() error {
	var errors []string

	// Validate query
	req.validateQuery(&errors)

	// Validate enum fields
	req.validateEnumFields(&errors)

	// Validate date range
	req.validateDateRange(&errors)

	if len(errors) > 0 {
		return &httpservice.ValidationError{Errors: errors}
	}

	return nil
}

// validateQuery validates the search query parameter
func (req *SearchRequest) validateQuery(errors *[]string) {
	trimmedQuery := strings.TrimSpace(req.Query)
	if trimmedQuery == "" {
		*errors = append(*errors, "search query cannot be empty")
		return
	}

	// Validate query length
	if len(trimmedQuery) < MinQueryLength {
		*errors = append(*errors, fmt.Sprintf("search query must be at least %d characters", MinQueryLength))
	}
	if len(trimmedQuery) > MaxQueryLength {
		*errors = append(*errors, fmt.Sprintf("search query cannot exceed %d characters", MaxQueryLength))
	}

	// Validate for potentially malicious patterns
	if containsSuspiciousPatterns(trimmedQuery) {
		*errors = append(*errors, "search query contains invalid characters")
	}
}

// validateEnumFields validates enum field values
func (req *SearchRequest) validateEnumFields(errors *[]string) {
	if req.ExperienceLevel != "" && !slices.Contains(validExperienceLevels, req.ExperienceLevel) {
		*errors = append(*errors, "invalid value for field: 'experience_level'")
	}

	if req.EmploymentType != "" && !slices.Contains(validEmploymentTypes, req.EmploymentType) {
		*errors = append(*errors, "invalid value for field: 'employment_type'")
	}

	if req.Location != "" && !slices.Contains(validLocations, req.Location) {
		*errors = append(*errors, "invalid value for field: 'location'")
	}

	if req.WorkMode != "" && !slices.Contains(validWorkModes, req.WorkMode) {
		*errors = append(*errors, "invalid value for field: 'work_mode'")
	}
}

// validateDateRange validates date range parameters
func (req *SearchRequest) validateDateRange(errors *[]string) {
	hasDateFrom := req.DateFrom != ""
	hasDateTo := req.DateTo != ""

	if hasDateFrom != hasDateTo {
		*errors = append(*errors, "both date_from and date_to must be provided together")
		return
	}

	// Validate date format if provided
	if hasDateFrom && hasDateTo {
		dateFrom, dateFromErr := time.Parse("2006-01-02", req.DateFrom)
		if dateFromErr != nil {
			*errors = append(*errors, "date_from must be in YYYY-MM-DD format")
		}

		dateTo, dateToErr := time.Parse("2006-01-02", req.DateTo)
		if dateToErr != nil {
			*errors = append(*errors, "date_to must be in YYYY-MM-DD format")
		}

		// Check date range if both dates are valid
		if dateFromErr == nil && dateToErr == nil && dateFrom.After(dateTo) {
			*errors = append(*errors, "date_from cannot be after date_to")
		}
	}
}

// JobResponse represents the API response for a single job
type JobResponse struct {
	ID               int                     `json:"job_id"`
	CompanyID        int                     `json:"company_id"`
	CompanyName      string                  `json:"company_name"`
	CompanyLogoURL   string                  `json:"company_logo_url"`
	Title            string                  `json:"title"`
	Description      string                  `json:"description"`
	Responsibilities []string                `json:"responsibilities"`
	Requirements     JobRequirementsResponse `json:"requirements"`
	MainTechnologies []string                `json:"main_technologies"`
	Benefits         []string                `json:"benefits"`
	ExperienceLevel  string                  `json:"experience_level"`
	EmploymentType   string                  `json:"employment_type"`
	Location         string                  `json:"location"`
	WorkMode         string                  `json:"work_mode"`
	ApplicationURL   string                  `json:"application_url"`
	Technologies     []TechnologyResponse    `json:"technologies"`
	PostedAt         time.Time               `json:"posted_at"`
}

// JobRequirementsResponse represents the requirements section of a job posting
type JobRequirementsResponse struct {
	MustHave   []string `json:"must_have"`
	NiceToHave []string `json:"nice_to_have"`
}

// TechnologyResponse represents the API response for job technologies
type TechnologyResponse struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Required bool   `json:"required"`
}

// SearchResponse represents the search response with pagination
type SearchResponse struct {
	Data       []*JobResponse    `json:"data"`
	Pagination PaginationDetails `json:"pagination"`
}

// PaginationDetails contains pagination metadata
type PaginationDetails struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

// ErrorDetails contains error information
type ErrorDetails struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// JobResponseList is a slice of JobResponse that implements httpservice.SearchResult interface
type JobResponseList []*JobResponse

// GetItems returns the job responses as []any to satisfy httpservice.SearchResult interface
func (jrl JobResponseList) GetItems() []any {
	items := make([]any, len(jrl))
	for i, item := range jrl {
		items[i] = item
	}
	return items
}

// GetTotal returns the length of the slice to satisfy httpservice.SearchResult interface
// Note: This returns the count of items in this slice, not the total search results count
func (jrl JobResponseList) GetTotal() int {
	return len(jrl)
}
