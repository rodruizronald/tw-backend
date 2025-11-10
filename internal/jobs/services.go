package jobs

import (
	"context"

	"github.com/rodruizronald/tw-backend/internal/httpservice"
)

// SearchService implements the httpservice.SearchService interface
type SearchService struct {
	repos DataRepository
}

// NewSearchService creates a new instance of SearchService
func NewSearchService(repos DataRepository) httpservice.SearchService[*SearchParams, JobResponseList] {
	return &SearchService{repos: repos}
}

// ExecuteSearch implements the SearchService interface to execute a search
func (s *SearchService) ExecuteSearch(ctx context.Context, params *SearchParams) (JobResponseList, int, error) {
	// Your existing business logic
	jobs, total, err := s.repos.SearchJobsWithCount(ctx, params)
	if err != nil {
		return nil, 0, &httpservice.SearchError{Operation: "search jobs", Err: err}
	}

	// Get job IDs for batch fetching technologies
	jobIDs := make([]int, len(jobs))
	for i, job := range jobs {
		jobIDs[i] = job.ID
	}

	// Batch fetch technologies for all jobs
	technologiesMap, err := s.repos.GetJobTechnologiesBatch(ctx, jobIDs)
	if err != nil {
		return nil, 0, &httpservice.SearchError{Operation: "fetch job technologies", Err: err}
	}

	// Convert jobs to response format with technologies
	searchResult := MapJobsToResponse(jobs, technologiesMap)

	return searchResult, total, nil
}

// ExecuteSearchCount implements the SearchService interface to execute a count search
func (s *SearchService) ExecuteSearchCount(ctx context.Context, params *SearchParams) (int, error) {
	count, err := s.repos.GetSearchCount(ctx, params)
	if err != nil {
		return 0, &httpservice.SearchError{Operation: "count jobs", Err: err}
	}
	return count, nil
}
