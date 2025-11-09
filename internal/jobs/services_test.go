package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodruizronald/tw-backend/internal/httpservice"
	"github.com/rodruizronald/tw-backend/internal/jobtech"
)

func TestJobSearchService_ExecuteSearch(t *testing.T) {
	t.Parallel()
	now := time.Now()
	searchError := errors.New("search error")
	technologiesError := errors.New("technologies error")

	tests := []struct {
		name         string
		params       *SearchParams
		mockSetup    func(mockRepo *MockDataRepository, params *SearchParams)
		checkResults func(t *testing.T, result JobResponseList, total int, err error)
	}{
		{
			name: "successful search with technologies",
			params: &SearchParams{
				Query:  "golang developer",
				Limit:  10,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				jobs := []*JobWithCompany{
					{
						Job: Job{
							ID:              1,
							CompanyID:       1,
							Title:           "Golang Developer",
							Description:     "Backend developer position",
							ExperienceLevel: "Mid-Level",
							EmploymentType:  "Full-Time",
							Location:        "Remote",
							WorkMode:        "Remote",
							ApplicationURL:  "https://example.com/apply1",
							IsActive:        true,
							Signature:       "job-signature-1",
							CreatedAt:       now,
							UpdatedAt:       now,
						},
						CompanyName: "Tech Corp",
					},
					{
						Job: Job{
							ID:              2,
							CompanyID:       2,
							Title:           "Senior Golang Engineer",
							Description:     "Senior backend position",
							ExperienceLevel: "Senior",
							EmploymentType:  "Full-Time",
							Location:        "San Francisco",
							WorkMode:        "Hybrid",
							ApplicationURL:  "https://example.com/apply2",
							IsActive:        true,
							Signature:       "job-signature-2",
							CreatedAt:       now,
							UpdatedAt:       now,
						},
						CompanyName: "Innovation Inc",
					},
				}
				technologiesMap := map[int][]*jobtech.JobTechnologyWithDetails{
					1: {
						{
							JobID:        1,
							TechnologyID: 1,
							TechName:     "Go",
							TechCategory: "Programming Language",
							IsRequired:   true,
						},
						{
							JobID:        1,
							TechnologyID: 2,
							TechName:     "PostgreSQL",
							TechCategory: "Database",
							IsRequired:   false,
						},
					},
					2: {
						{
							JobID:        2,
							TechnologyID: 1,
							TechName:     "Go",
							TechCategory: "Programming Language",
							IsRequired:   true,
						},
					},
				}
				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return(jobs, 25, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{1, 2}).
					Return(technologiesMap, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, result, 2)
				assert.Equal(t, 25, total)

				// Check first job
				assert.Equal(t, 1, result[0].ID)
				assert.Equal(t, "Golang Developer", result[0].Title)
				assert.Equal(t, "Tech Corp", result[0].CompanyName)
				assert.Len(t, result[0].Technologies, 2)
				assert.Equal(t, "Go", result[0].Technologies[0].Name)
				assert.True(t, result[0].Technologies[0].Required)
				assert.Equal(t, "PostgreSQL", result[0].Technologies[1].Name)
				assert.False(t, result[0].Technologies[1].Required)

				// Check second job
				assert.Equal(t, 2, result[1].ID)
				assert.Equal(t, "Senior Golang Engineer", result[1].Title)
				assert.Equal(t, "Innovation Inc", result[1].CompanyName)
				assert.Len(t, result[1].Technologies, 1)
				assert.Equal(t, "Go", result[1].Technologies[0].Name)
				assert.True(t, result[1].Technologies[0].Required)
			},
		},
		{
			name: "successful search with no results",
			params: &SearchParams{
				Query:  "nonexistent technology",
				Limit:  20,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return([]*JobWithCompany{}, 0, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{}).
					Return(map[int][]*jobtech.JobTechnologyWithDetails{}, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, result)
				assert.Equal(t, 0, total)
			},
		},
		{
			name: "successful search with jobs but no technologies",
			params: &SearchParams{
				Query:  "simple job",
				Limit:  5,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				jobs := []*JobWithCompany{
					{
						Job: Job{
							ID:              3,
							CompanyID:       3,
							Title:           "Simple Job",
							Description:     "Basic position",
							ExperienceLevel: "Entry-Level",
							EmploymentType:  "Part-Time",
							Location:        "Remote",
							WorkMode:        "Remote",
							ApplicationURL:  "https://example.com/apply3",
							IsActive:        true,
							Signature:       "job-signature-3",
							CreatedAt:       now,
							UpdatedAt:       now,
						},
						CompanyName: "Simple Corp",
					},
				}

				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return(jobs, 1, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{3}).
					Return(map[int][]*jobtech.JobTechnologyWithDetails{}, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, result, 1)
				assert.Equal(t, 1, total)

				assert.Equal(t, 3, result[0].ID)
				assert.Equal(t, "Simple Job", result[0].Title)
				assert.Equal(t, "Simple Corp", result[0].CompanyName)
				assert.Empty(t, result[0].Technologies)
			},
		},
		{
			name: "search with all filters applied",
			params: &SearchParams{
				Query:           "senior developer",
				Limit:           5,
				Offset:          10,
				ExperienceLevel: stringPtr("Senior"),
				EmploymentType:  stringPtr("Full-Time"),
				Location:        stringPtr("San Francisco"),
				WorkMode:        stringPtr("Remote"),
				Company:         stringPtr("TechCorp"),
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				jobs := []*JobWithCompany{
					{
						Job: Job{
							ID:              4,
							CompanyID:       4,
							Title:           "Senior Developer",
							Description:     "Senior position with all filters",
							ExperienceLevel: "Senior",
							EmploymentType:  "Full-Time",
							Location:        "San Francisco",
							WorkMode:        "Remote",
							ApplicationURL:  "https://example.com/apply4",
							IsActive:        true,
							Signature:       "job-signature-4",
							CreatedAt:       now,
							UpdatedAt:       now,
						},
						CompanyName: "TechCorp",
					},
				}
				technologiesMap := map[int][]*jobtech.JobTechnologyWithDetails{
					4: {
						{
							JobID:        4,
							TechnologyID: 3,
							TechName:     "React",
							TechCategory: "Frontend Framework",
							IsRequired:   true,
						},
					},
				}

				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return(jobs, 42, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{4}).
					Return(technologiesMap, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, result, 1)
				assert.Equal(t, 42, total)

				assert.Equal(t, 4, result[0].ID)
				assert.Equal(t, "Senior Developer", result[0].Title)
				assert.Equal(t, "Senior", result[0].ExperienceLevel)
				assert.Equal(t, "Full-Time", result[0].EmploymentType)
				assert.Equal(t, "San Francisco", result[0].Location)
				assert.Equal(t, "Remote", result[0].WorkMode)
				assert.Equal(t, "TechCorp", result[0].CompanyName)
				assert.Len(t, result[0].Technologies, 1)
				assert.Equal(t, "React", result[0].Technologies[0].Name)
			},
		},
		{
			name: "error during job search",
			params: &SearchParams{
				Query:  "error query",
				Limit:  10,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return(nil, 0, searchError).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, 0, total)

				var searchErr *httpservice.SearchError
				require.ErrorAs(t, err, &searchErr)
				assert.Equal(t, "search jobs", searchErr.Operation)
				require.ErrorIs(t, searchErr.Err, searchError)
			},
		},
		{
			name: "error during technologies fetch",
			params: &SearchParams{
				Query:  "tech error query",
				Limit:  10,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				jobs := []*JobWithCompany{
					{
						Job: Job{
							ID:              5,
							CompanyID:       5,
							Title:           "Test Job",
							Description:     "Test description",
							ExperienceLevel: "Mid-Level",
							EmploymentType:  "Full-Time",
							Location:        "Remote",
							WorkMode:        "Remote",
							ApplicationURL:  "https://example.com/apply5",
							IsActive:        true,
							Signature:       "job-signature-5",
							CreatedAt:       now,
							UpdatedAt:       now,
						},
						CompanyName: "Test Corp",
					},
				}

				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return(jobs, 1, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{5}).
					Return(nil, technologiesError).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, 0, total)

				var searchErr *httpservice.SearchError
				require.ErrorAs(t, err, &searchErr)
				assert.Equal(t, "fetch job technologies", searchErr.Operation)
				require.ErrorIs(t, searchErr.Err, technologiesError)
			},
		},
		{
			name: "edge case: empty query string",
			params: &SearchParams{
				Query:  "",
				Limit:  10,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return([]*JobWithCompany{}, 0, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{}).
					Return(map[int][]*jobtech.JobTechnologyWithDetails{}, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, result)
				assert.Equal(t, 0, total)
			},
		},
		{
			name: "edge case: maximum limit",
			params: &SearchParams{
				Query:  "max limit test",
				Limit:  100,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				// Create a large slice of jobs to test maximum limit
				jobs := make([]*JobWithCompany, 100)
				jobIDs := make([]int, 100)
				technologiesMap := make(map[int][]*jobtech.JobTechnologyWithDetails)

				for i := 0; i < 100; i++ {
					jobID := i + 1
					jobs[i] = &JobWithCompany{
						Job: Job{
							ID:              jobID,
							CompanyID:       1,
							Title:           "Test Job",
							Description:     "Test description",
							ExperienceLevel: "Mid-Level",
							EmploymentType:  "Full-Time",
							Location:        "Remote",
							WorkMode:        "Remote",
							ApplicationURL:  "https://example.com/apply",
							IsActive:        true,
							Signature:       "job-signature",
							CreatedAt:       now,
							UpdatedAt:       now,
						},
						CompanyName: "Test Corp",
					}
					jobIDs[i] = jobID
					technologiesMap[jobID] = []*jobtech.JobTechnologyWithDetails{}
				}

				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return(jobs, 1000, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), jobIDs).
					Return(technologiesMap, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, result, 100)
				assert.Equal(t, 1000, total)
			},
		},
		{
			name: "edge case: high offset",
			params: &SearchParams{
				Query:  "high offset test",
				Limit:  10,
				Offset: 9990,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return([]*JobWithCompany{}, 10000, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{}).
					Return(map[int][]*jobtech.JobTechnologyWithDetails{}, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, result)
				assert.Equal(t, 10000, total) // Total should still reflect the actual count
			},
		},
		{
			name: "boundary case: single job with many technologies",
			params: &SearchParams{
				Query:  "full stack developer",
				Limit:  1,
				Offset: 0,
			},
			mockSetup: func(mockRepo *MockDataRepository, params *SearchParams) {
				t.Helper()
				jobs := []*JobWithCompany{
					{
						Job: Job{
							ID:              6,
							CompanyID:       6,
							Title:           "Full Stack Developer",
							Description:     "Full stack position",
							ExperienceLevel: "Senior",
							EmploymentType:  "Full-Time",
							Location:        "Remote",
							WorkMode:        "Remote",
							ApplicationURL:  "https://example.com/apply6",
							IsActive:        true,
							Signature:       "job-signature-6",
							CreatedAt:       now,
							UpdatedAt:       now,
						},
						CompanyName: "Full Stack Corp",
					},
				}

				// Many technologies for a single job
				technologiesMap := map[int][]*jobtech.JobTechnologyWithDetails{
					6: {
						{JobID: 6, TechnologyID: 1, TechName: "React", TechCategory: "Frontend", IsRequired: true},
						{JobID: 6, TechnologyID: 2, TechName: "Node.js", TechCategory: "Backend", IsRequired: true},
						{JobID: 6, TechnologyID: 3, TechName: "PostgreSQL", TechCategory: "Database", IsRequired: true},
						{JobID: 6, TechnologyID: 4, TechName: "Docker", TechCategory: "DevOps", IsRequired: false},
						{JobID: 6, TechnologyID: 5, TechName: "AWS", TechCategory: "Cloud", IsRequired: false},
					},
				}

				mockRepo.EXPECT().SearchJobsWithCount(context.Background(), params).
					Return(jobs, 1, nil).Once()

				mockRepo.EXPECT().GetJobTechnologiesBatch(context.Background(), []int{6}).
					Return(technologiesMap, nil).Once()
			},
			checkResults: func(t *testing.T, result JobResponseList, total int, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, result, 1)
				assert.Equal(t, 1, total)

				assert.Equal(t, 6, result[0].ID)
				assert.Equal(t, "Full Stack Developer", result[0].Title)
				assert.Len(t, result[0].Technologies, 5)

				// Verify required technologies
				requiredTechs := 0
				for _, tech := range result[0].Technologies {
					if tech.Required {
						requiredTechs++
					}
				}
				assert.Equal(t, 3, requiredTechs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := NewMockDataRepository(t)
			service := NewSearchService(mockRepo)

			tt.mockSetup(mockRepo, tt.params)

			result, total, err := service.ExecuteSearch(context.Background(), tt.params)
			tt.checkResults(t, result, total, err)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
