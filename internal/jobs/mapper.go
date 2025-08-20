package jobs

import "github.com/rodruizronald/ticos-in-tech/internal/jobtech"

// Mapping functions to convert between database and API models.
// This file contains transformation logic that bridges the repository layer (database models)
// and the API layer (DTOs). Mappers handle the conversion of internal data structures
// to external API representations, including data aggregation and formatting.

// MapJobToResponse converts a single job with company data to API response format.
// It transforms a database model into a DTO suitable for API responses.
func MapJobToResponse(job *JobWithCompany, technologies []TechnologyResponse) *JobResponse {
	return &JobResponse{
		ID:               job.ID,
		CompanyID:        job.CompanyID,
		CompanyName:      job.CompanyName,
		Title:            job.Title,
		Description:      job.Description,
		Responsibilities: job.Responsibilities,
		Requirements: JobRequirementsResponse{
			MustHave:   job.SkillMustHave,
			NiceToHave: job.SkillNiceHave,
		},
		MainTechnologies: job.MainTechnologies,
		Benefits:         job.Benefits,
		ExperienceLevel:  job.ExperienceLevel,
		EmploymentType:   job.EmploymentType,
		Location:         job.Location,
		WorkMode:         job.WorkMode,
		ApplicationURL:   job.ApplicationURL,
		Technologies:     technologies,
		PostedAt:         job.CreatedAt,
	}
}

// MapJobsToResponse converts jobs with technologies to API response format.
// It takes jobs with company data and technologies map, transforming them into JobResponse DTOs.
func MapJobsToResponse(jobs []*JobWithCompany, techMap map[int][]*jobtech.JobTechnologyWithDetails) []*JobResponse {
	jobResponses := make([]*JobResponse, len(jobs))

	for i, job := range jobs {
		// Convert technologies for this job
		jobTechnologies := techMap[job.ID]
		technologies := make([]TechnologyResponse, len(jobTechnologies))
		for j, tech := range jobTechnologies {
			technologies[j] = TechnologyResponse{
				Name:     tech.TechName,
				Category: tech.TechCategory,
				Required: tech.IsRequired,
			}
		}

		// Use the single job mapper
		jobResponses[i] = MapJobToResponse(job, technologies)
	}

	return jobResponses
}
