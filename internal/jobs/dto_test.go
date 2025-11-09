package jobs

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodruizronald/tw-backend/internal/httpservice"
)

func TestSearchRequest_ToSearchParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		request      *SearchRequest
		checkResults func(t *testing.T, result httpservice.SearchParams, err error)
	}{
		{
			name: "successful conversion with all fields",
			request: &SearchRequest{
				Query:           "golang developer",
				Limit:           25,
				Offset:          10,
				ExperienceLevel: "Senior",
				EmploymentType:  "Full-Time",
				Location:        "Costa Rica",
				WorkMode:        "Remote",
				Company:         "Tech Corp",
				DateFrom:        "2024-01-01",
				DateTo:          "2024-12-31",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "golang developer", searchParams.Query)
				assert.Equal(t, 25, searchParams.Limit)
				assert.Equal(t, 10, searchParams.Offset)
				assert.NotNil(t, searchParams.ExperienceLevel)
				assert.Equal(t, "Senior", *searchParams.ExperienceLevel)
				assert.NotNil(t, searchParams.EmploymentType)
				assert.Equal(t, "Full-Time", *searchParams.EmploymentType)
				assert.NotNil(t, searchParams.Location)
				assert.Equal(t, "Costa Rica", *searchParams.Location)
				assert.NotNil(t, searchParams.WorkMode)
				assert.Equal(t, "Remote", *searchParams.WorkMode)
				assert.NotNil(t, searchParams.Company)
				assert.Equal(t, "Tech Corp", *searchParams.Company)
				assert.NotNil(t, searchParams.DateFrom)
				assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), *searchParams.DateFrom)
				assert.NotNil(t, searchParams.DateTo)
				assert.Equal(t, time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), *searchParams.DateTo)
			},
		},
		{
			name: "successful conversion with minimal fields",
			request: &SearchRequest{
				Query:  "python",
				Limit:  0,
				Offset: 0,
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "python", searchParams.Query)
				assert.Equal(t, DefaultLimit, searchParams.Limit) // Should use default
				assert.Equal(t, 0, searchParams.Offset)
				assert.Nil(t, searchParams.ExperienceLevel)
				assert.Nil(t, searchParams.EmploymentType)
				assert.Nil(t, searchParams.Location)
				assert.Nil(t, searchParams.WorkMode)
				assert.Nil(t, searchParams.Company)
				assert.Nil(t, searchParams.DateFrom)
				assert.Nil(t, searchParams.DateTo)
			},
		},
		{
			name: "limit exceeds maximum - should be capped",
			request: &SearchRequest{
				Query:  "javascript",
				Limit:  150, // Exceeds MaxLimit (100)
				Offset: 5,
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "javascript", searchParams.Query)
				assert.Equal(t, MaxLimit, searchParams.Limit) // Should be capped at MaxLimit
				assert.Equal(t, 5, searchParams.Offset)
			},
		},
		{
			name: "negative limit - should use default",
			request: &SearchRequest{
				Query:  "react",
				Limit:  -10,
				Offset: 0,
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "react", searchParams.Query)
				assert.Equal(t, DefaultLimit, searchParams.Limit) // Should use default
				assert.Equal(t, 0, searchParams.Offset)
			},
		},
		{
			name: "negative offset - should be set to zero",
			request: &SearchRequest{
				Query:  "vue",
				Limit:  10,
				Offset: -5,
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "vue", searchParams.Query)
				assert.Equal(t, 10, searchParams.Limit)
				assert.Equal(t, 0, searchParams.Offset) // Should be set to 0
			},
		},
		{
			name: "empty optional fields - should be nil pointers",
			request: &SearchRequest{
				Query:           "node.js",
				Limit:           15,
				Offset:          0,
				ExperienceLevel: "",
				EmploymentType:  "",
				Location:        "",
				WorkMode:        "",
				Company:         "",
				DateFrom:        "",
				DateTo:          "",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "node.js", searchParams.Query)
				assert.Equal(t, 15, searchParams.Limit)
				assert.Equal(t, 0, searchParams.Offset)
				assert.Nil(t, searchParams.ExperienceLevel)
				assert.Nil(t, searchParams.EmploymentType)
				assert.Nil(t, searchParams.Location)
				assert.Nil(t, searchParams.WorkMode)
				assert.Nil(t, searchParams.Company)
				assert.Nil(t, searchParams.DateFrom)
				assert.Nil(t, searchParams.DateTo)
			},
		},
		{
			name: "only date_from provided - should not set dates",
			request: &SearchRequest{
				Query:    "docker",
				Limit:    20,
				Offset:   0,
				DateFrom: "2024-01-01",
				DateTo:   "",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "docker", searchParams.Query)
				assert.Nil(t, searchParams.DateFrom)
				assert.Nil(t, searchParams.DateTo)
			},
		},
		{
			name: "only date_to provided - should not set dates",
			request: &SearchRequest{
				Query:    "kubernetes",
				Limit:    20,
				Offset:   0,
				DateFrom: "",
				DateTo:   "2024-12-31",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "kubernetes", searchParams.Query)
				assert.Nil(t, searchParams.DateFrom)
				assert.Nil(t, searchParams.DateTo)
			},
		},
		{
			name: "invalid date_from format",
			request: &SearchRequest{
				Query:    "aws",
				Limit:    10,
				Offset:   0,
				DateFrom: "invalid-date",
				DateTo:   "2024-12-31",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.Error(t, err)
				assert.Nil(t, result)

				var convErr *httpservice.ConversionError
				require.ErrorAs(t, err, &convErr)
				assert.Equal(t, "date_from", convErr.Field)
				assert.Equal(t, "invalid-date", convErr.Value)
			},
		},
		{
			name: "invalid date_to format",
			request: &SearchRequest{
				Query:    "gcp",
				Limit:    10,
				Offset:   0,
				DateFrom: "2024-01-01",
				DateTo:   "not-a-date",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.Error(t, err)
				assert.Nil(t, result)

				var convErr *httpservice.ConversionError
				require.ErrorAs(t, err, &convErr)
				assert.Equal(t, "date_to", convErr.Field)
				assert.Equal(t, "not-a-date", convErr.Value)
			},
		},
		{
			name: "edge case: maximum valid limit",
			request: &SearchRequest{
				Query:  "machine learning",
				Limit:  MaxLimit, // Exactly at the limit
				Offset: 0,
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "machine learning", searchParams.Query)
				assert.Equal(t, MaxLimit, searchParams.Limit)
				assert.Equal(t, 0, searchParams.Offset)
			},
		},
		{
			name: "edge case: very high offset",
			request: &SearchRequest{
				Query:  "data science",
				Limit:  10,
				Offset: 999999,
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "data science", searchParams.Query)
				assert.Equal(t, 10, searchParams.Limit)
				assert.Equal(t, 999999, searchParams.Offset)
			},
		},
		{
			name: "boundary case: leap year date",
			request: &SearchRequest{
				Query:    "blockchain",
				Limit:    10,
				Offset:   0,
				DateFrom: "2024-02-29", // Leap year
				DateTo:   "2024-03-01",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "blockchain", searchParams.Query)
				assert.NotNil(t, searchParams.DateFrom)
				assert.Equal(t, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), *searchParams.DateFrom)
				assert.NotNil(t, searchParams.DateTo)
				assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), *searchParams.DateTo)
			},
		},
		{
			name: "boundary case: year boundaries",
			request: &SearchRequest{
				Query:    "devops",
				Limit:    10,
				Offset:   0,
				DateFrom: "2023-12-31",
				DateTo:   "2024-01-01",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "devops", searchParams.Query)
				assert.NotNil(t, searchParams.DateFrom)
				assert.Equal(t, time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), *searchParams.DateFrom)
				assert.NotNil(t, searchParams.DateTo)
				assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), *searchParams.DateTo)
			},
		},
		{
			name: "edge case: same date for from and to",
			request: &SearchRequest{
				Query:    "cybersecurity",
				Limit:    10,
				Offset:   0,
				DateFrom: "2024-06-15",
				DateTo:   "2024-06-15",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "cybersecurity", searchParams.Query)
				assert.NotNil(t, searchParams.DateFrom)
				assert.NotNil(t, searchParams.DateTo)
				assert.Equal(t, *searchParams.DateFrom, *searchParams.DateTo)
			},
		},
		{
			name: "whitespace in optional fields - should be set as non-nil",
			request: &SearchRequest{
				Query:           "ai",
				Limit:           10,
				Offset:          0,
				ExperienceLevel: "   Senior   ",
				EmploymentType:  " Full-Time ",
				Location:        "  Costa Rica  ",
				WorkMode:        " Remote ",
				Company:         "  Tech Corp  ",
			},
			checkResults: func(t *testing.T, result httpservice.SearchParams, err error) {
				t.Helper()
				require.NoError(t, err)

				searchParams := result.(*SearchParams)
				assert.Equal(t, "ai", searchParams.Query)
				// Note: The function doesn't trim whitespace, so these should be set as-is
				assert.NotNil(t, searchParams.ExperienceLevel)
				assert.Equal(t, "   Senior   ", *searchParams.ExperienceLevel)
				assert.NotNil(t, searchParams.EmploymentType)
				assert.Equal(t, " Full-Time ", *searchParams.EmploymentType)
				assert.NotNil(t, searchParams.Location)
				assert.Equal(t, "  Costa Rica  ", *searchParams.Location)
				assert.NotNil(t, searchParams.WorkMode)
				assert.Equal(t, " Remote ", *searchParams.WorkMode)
				assert.NotNil(t, searchParams.Company)
				assert.Equal(t, "  Tech Corp  ", *searchParams.Company)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.request.ToSearchParams()
			tt.checkResults(t, result, err)
		})
	}
}

func TestSearchRequest_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		request      *SearchRequest
		checkResults func(t *testing.T, err error)
	}{
		{
			name: "valid request with all fields",
			request: &SearchRequest{
				Query:           "golang developer",
				Limit:           25,
				Offset:          10,
				ExperienceLevel: experienceLevelSenior,
				EmploymentType:  employmentTypeFullTime,
				Location:        locationCostaRica,
				WorkMode:        workModeRemote,
				Company:         "Tech Corp",
				DateFrom:        "2024-01-01",
				DateTo:          "2024-12-31",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
		{
			name: "valid request with empty optional fields",
			request: &SearchRequest{
				Query:  "javascript",
				Limit:  20,
				Offset: 5,
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
		{
			name: "valid date range",
			request: &SearchRequest{
				Query:    "developer",
				DateFrom: "2024-01-01",
				DateTo:   "2024-12-31",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
		{
			name: "valid same date for from and to",
			request: &SearchRequest{
				Query:    "engineer",
				DateFrom: "2024-06-15",
				DateTo:   "2024-06-15",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
		{
			name: "empty query",
			request: &SearchRequest{
				Query:  "",
				Limit:  10,
				Offset: 0,
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "search query cannot be empty")
			},
		},
		{
			name: "whitespace only query",
			request: &SearchRequest{
				Query:  "   \t\n   ",
				Limit:  10,
				Offset: 0,
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "search query cannot be empty")
			},
		},
		{
			name: "query too short",
			request: &SearchRequest{
				Query: "a", // Only 1 character, below MinQueryLength (2)
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "search query must be at least 2 characters")
			},
		},
		{
			name: "query too long",
			request: &SearchRequest{
				Query: strings.Repeat("a", 101), // 101 characters, exceeds MaxQueryLength (100)
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "search query cannot exceed 100 characters")
			},
		},
		{
			name: "invalid experience level",
			request: &SearchRequest{
				Query:           "developer",
				ExperienceLevel: "Invalid-Level",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "invalid value for field: 'experience_level'")
			},
		},
		{
			name: "invalid employment type",
			request: &SearchRequest{
				Query:          "engineer",
				EmploymentType: "Invalid-Type",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "invalid value for field: 'employment_type'")
			},
		},
		{
			name: "invalid location",
			request: &SearchRequest{
				Query:    "developer",
				Location: "Invalid-Location",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "invalid value for field: 'location'")
			},
		},
		{
			name: "invalid work mode",
			request: &SearchRequest{
				Query:    "engineer",
				WorkMode: "Invalid-Mode",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "invalid value for field: 'work_mode'")
			},
		},
		{
			name: "only date_from provided",
			request: &SearchRequest{
				Query:    "developer",
				DateFrom: "2024-01-01",
				DateTo:   "",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "both date_from and date_to must be provided together")
			},
		},
		{
			name: "only date_to provided",
			request: &SearchRequest{
				Query:    "engineer",
				DateFrom: "",
				DateTo:   "2024-12-31",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "both date_from and date_to must be provided together")
			},
		},
		{
			name: "invalid date_from format",
			request: &SearchRequest{
				Query:    "developer",
				DateFrom: "invalid-date",
				DateTo:   "2024-12-31",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "date_from must be in YYYY-MM-DD format")
			},
		},
		{
			name: "invalid date_to format",
			request: &SearchRequest{
				Query:    "engineer",
				DateFrom: "2024-01-01",
				DateTo:   "not-a-date",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "date_to must be in YYYY-MM-DD format")
			},
		},
		{
			name: "date_from after date_to",
			request: &SearchRequest{
				Query:    "developer",
				DateFrom: "2024-12-31",
				DateTo:   "2024-01-01",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "date_from cannot be after date_to")
			},
		},
		{
			name: "boundary case: invalid leap year date",
			request: &SearchRequest{
				Query:    "developer",
				DateFrom: "2023-02-29", // Invalid - 2023 is not a leap year
				DateTo:   "2023-03-01",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "date_from must be in YYYY-MM-DD format")
			},
		},
		{
			name: "boundary case: date format without leading zeros",
			request: &SearchRequest{
				Query:    "developer",
				DateFrom: "2024-1-1", // Invalid format
				DateTo:   "2024-1-31",
			},
			checkResults: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)

				var validationErr *httpservice.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Errors, "date_from must be in YYYY-MM-DD format")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.request.Validate()
			tt.checkResults(t, err)
		})
	}
}
