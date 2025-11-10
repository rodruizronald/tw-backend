// Package devmocks provides mock implementations of repository interfaces for development and testing.
// It loads sample job and technology data from embedded JSON files to simulate database operations
// without requiring an actual database connection.
package devmocks

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rodruizronald/tw-backend/internal/jobs"
	"github.com/rodruizronald/tw-backend/internal/jobtech"
)

//go:embed data/mock_jobs.json
var mockDataFS embed.FS

// JobRepository provides mock data for development and testing
type JobRepository struct {
	jobs         []*jobs.JobWithCompany
	technologies map[int][]*jobtech.JobTechnologyWithDetails
}

// mockJobData represents the JSON structure for a job
type mockJobData struct {
	ID               int      `json:"id"`
	CompanyID        int      `json:"company_id"`
	Title            string   `json:"title"`
	OriginalPost     string   `json:"original_post"`
	Description      string   `json:"description"`
	Responsibilities []string `json:"responsibilities"`
	SkillMustHave    []string `json:"skill_must_have"`
	SkillNiceHave    []string `json:"skill_nice_have"`
	MainTechnologies []string `json:"main_technologies"`
	Benefits         []string `json:"benefits"`
	ExperienceLevel  string   `json:"experience_level"`
	EmploymentType   string   `json:"employment_type"`
	Location         string   `json:"location"`
	WorkMode         string   `json:"work_mode"`
	ApplicationURL   string   `json:"application_url"`
	IsActive         bool     `json:"is_active"`
	Signature        string   `json:"signature"`
	DaysAgo          int      `json:"days_ago"`
	CompanyName      string   `json:"company_name"`
}

// mockTechnologyData represents the JSON structure for a technology
type mockTechnologyData struct {
	JobID        int    `json:"job_id"`
	TechnologyID int    `json:"technology_id"`
	TechName     string `json:"tech_name"`
	TechCategory string `json:"tech_category"`
	IsRequired   bool   `json:"is_required"`
}

// mockData represents the entire JSON structure
type mockData struct {
	Jobs         []mockJobData                   `json:"jobs"`
	Technologies map[string][]mockTechnologyData `json:"technologies"`
}

// NewJobRepository creates a new mock job repository with sample data
func NewJobRepository() *JobRepository {
	mockJobs, mockTechnologies, err := loadMockData()
	if err != nil {
		// Fallback to empty data if loading fails
		return &JobRepository{
			jobs:         []*jobs.JobWithCompany{},
			technologies: make(map[int][]*jobtech.JobTechnologyWithDetails),
		}
	}
	return &JobRepository{
		jobs:         mockJobs,
		technologies: mockTechnologies,
	}
}

// SearchJobsWithCount implements the jobs.DataRepository interface with mock data
func (r *JobRepository) SearchJobsWithCount(_ context.Context, params *jobs.SearchParams) (
	[]*jobs.JobWithCompany, int, error) {
	filteredJobs := r.filterJobs(params)
	total := len(filteredJobs)

	// Apply pagination
	start := params.Offset
	end := start + params.Limit

	if start >= total {
		return []*jobs.JobWithCompany{}, total, nil
	}

	if end > total {
		end = total
	}

	paginatedJobs := filteredJobs[start:end]
	return paginatedJobs, total, nil
}

// GetJobTechnologiesBatch implements the jobs.DataRepository interface
func (r *JobRepository) GetJobTechnologiesBatch(_ context.Context, jobIDs []int) (
	map[int][]*jobtech.JobTechnologyWithDetails, error) {
	result := make(map[int][]*jobtech.JobTechnologyWithDetails)

	for _, jobID := range jobIDs {
		if techs, exists := r.technologies[jobID]; exists {
			result[jobID] = techs
		}
	}

	return result, nil
}

// GetSearchCount implements the jobs.DataRepository interface with mock data
func (r *JobRepository) GetSearchCount(_ context.Context, params *jobs.SearchParams) (int, error) {
	filteredJobs := r.filterJobs(params)
	return len(filteredJobs), nil
}

// filterJobs applies search filters to the mock data
func (r *JobRepository) filterJobs(params *jobs.SearchParams) []*jobs.JobWithCompany {
	var filtered []*jobs.JobWithCompany

	for _, job := range r.jobs {
		if r.jobMatchesFilters(job, params) {
			filtered = append(filtered, job)
		}
	}

	return filtered
}

// jobMatchesFilters checks if a job matches all the search criteria
func (r *JobRepository) jobMatchesFilters(job *jobs.JobWithCompany, params *jobs.SearchParams) bool {
	return r.matchesQuery(job, params.Query) &&
		r.matchesExperienceLevel(job, params.ExperienceLevel) &&
		r.matchesEmploymentType(job, params.EmploymentType) &&
		r.matchesLocation(job, params.Location) &&
		r.matchesWorkMode(job, params.WorkMode) &&
		r.matchesCompany(job, params.Company) &&
		r.matchesDateRange(job, params.DateFrom, params.DateTo)
}

// matchesQuery checks if the job matches the search query
func (r *JobRepository) matchesQuery(job *jobs.JobWithCompany, query string) bool {
	if query == "" {
		return true
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	return strings.Contains(strings.ToLower(job.Title), normalizedQuery) ||
		strings.Contains(strings.ToLower(job.CompanyName), normalizedQuery) ||
		strings.Contains(strings.ToLower(job.Description), normalizedQuery)
}

// matchesExperienceLevel checks if the job matches the experience level filter
func (r *JobRepository) matchesExperienceLevel(job *jobs.JobWithCompany, experienceLevel *string) bool {
	if experienceLevel == nil {
		return true
	}
	return strings.EqualFold(job.ExperienceLevel, *experienceLevel)
}

// matchesEmploymentType checks if the job matches the employment type filter
func (r *JobRepository) matchesEmploymentType(job *jobs.JobWithCompany, employmentType *string) bool {
	if employmentType == nil {
		return true
	}
	return strings.EqualFold(job.EmploymentType, *employmentType)
}

// matchesLocation checks if the job matches the location filter
func (r *JobRepository) matchesLocation(job *jobs.JobWithCompany, location *string) bool {
	if location == nil {
		return true
	}
	return strings.Contains(strings.ToLower(job.Location), strings.ToLower(*location))
}

// matchesWorkMode checks if the job matches the work mode filter
func (r *JobRepository) matchesWorkMode(job *jobs.JobWithCompany, workMode *string) bool {
	if workMode == nil {
		return true
	}
	return strings.EqualFold(job.WorkMode, *workMode)
}

// matchesCompany checks if the job matches the company filter
func (r *JobRepository) matchesCompany(job *jobs.JobWithCompany, company *string) bool {
	if company == nil {
		return true
	}
	return strings.Contains(strings.ToLower(job.CompanyName), strings.ToLower(*company))
}

// matchesDateRange checks if the job matches the date range filters
func (r *JobRepository) matchesDateRange(job *jobs.JobWithCompany, dateFrom, dateTo *time.Time) bool {
	if dateFrom != nil && job.CreatedAt.Before(*dateFrom) {
		return false
	}
	if dateTo != nil && job.CreatedAt.After(*dateTo) {
		return false
	}
	return true
}

// loadMockData loads mock data from the embedded JSON file
func loadMockData() ([]*jobs.JobWithCompany, map[int][]*jobtech.JobTechnologyWithDetails, error) {
	data, err := mockDataFS.ReadFile("data/mock_jobs.json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read mock data file: %w", err)
	}

	var mockDataContent mockData
	if err := json.Unmarshal(data, &mockDataContent); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal mock data: %w", err)
	}

	now := time.Now()
	var jobsWithCompany []*jobs.JobWithCompany
	technologies := make(map[int][]*jobtech.JobTechnologyWithDetails)

	// Convert mock job data to JobWithCompany structs
	for i := range mockDataContent.Jobs {
		jobData := &mockDataContent.Jobs[i]
		job := &jobs.JobWithCompany{
			Job: jobs.Job{
				ID:               jobData.ID,
				CompanyID:        jobData.CompanyID,
				Title:            jobData.Title,
				Description:      jobData.Description,
				Responsibilities: jobData.Responsibilities,
				SkillMustHave:    jobData.SkillMustHave,
				SkillNiceHave:    jobData.SkillNiceHave,
				Benefits:         jobData.Benefits,
				MainTechnologies: jobData.MainTechnologies,
				ExperienceLevel:  jobData.ExperienceLevel,
				EmploymentType:   jobData.EmploymentType,
				Location:         jobData.Location,
				WorkMode:         jobData.WorkMode,
				ApplicationURL:   jobData.ApplicationURL,
				IsActive:         jobData.IsActive,
				Signature:        jobData.Signature,
				CreatedAt:        now.AddDate(0, 0, -jobData.DaysAgo),
				UpdatedAt:        now.AddDate(0, 0, -1),
			},
			CompanyName: jobData.CompanyName,
		}
		jobsWithCompany = append(jobsWithCompany, job)
	}

	// Convert mock technology data
	for jobIDStr, techList := range mockDataContent.Technologies {
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			continue // Skip invalid job IDs
		}

		var jobTechs []*jobtech.JobTechnologyWithDetails
		for i := range techList {
			techData := &techList[i]
			tech := &jobtech.JobTechnologyWithDetails{
				JobID:        techData.JobID,
				TechnologyID: techData.TechnologyID,
				TechName:     techData.TechName,
				TechCategory: techData.TechCategory,
				IsRequired:   techData.IsRequired,
			}
			jobTechs = append(jobTechs, tech)
		}
		technologies[jobID] = jobTechs
	}

	return jobsWithCompany, technologies, nil
}
