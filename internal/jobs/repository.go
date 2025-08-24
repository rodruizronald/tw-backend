package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SQL query constants
const (
	// Base query for selecting job fields
	selectJobBaseQuery = `
        SELECT id, company_id, title, description, responsibilities, skill_must_have,
               skill_nice_have, main_technologies, benefits, experience_level, employment_type,
               location, work_mode, application_url, is_active, signature, created_at, updated_at
        FROM jobs
    `

	createJobQuery = `
        INSERT INTO jobs (
            company_id, title, description, responsibilities, skill_must_have, 
            skill_nice_have, main_technologies, benefits, experience_level, employment_type,
            location, work_mode, application_url, is_active, signature
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        RETURNING id, created_at, updated_at
    `

	getJobByIDQuery = selectJobBaseQuery + `
        WHERE id = $1
    `

	getJobBySignatureQuery = selectJobBaseQuery + `
        WHERE signature = $1
    `

	updateJobQuery = `
        UPDATE jobs
        SET company_id = $1, title = $2, description = $3, responsibilities = $4,
            skill_must_have = $5, skill_nice_have = $6, main_technologies = $7, benefits = $8,
            experience_level = $9, employment_type = $10, location = $11, work_mode = $12,
            application_url = $13, is_active = $14, signature = $15, updated_at = NOW()
        WHERE id = $16
        RETURNING updated_at
    `

	deleteJobQuery = `DELETE FROM jobs WHERE id = $1`

	// Full-text search query with company data and total count using window function
	searchJobsWithCountBaseQuery = `
        WITH search_query AS (
            SELECT plainto_tsquery('english', $1) AS query
        )
        SELECT 
            j.id, j.company_id, j.title, j.description, j.responsibilities, j.skill_must_have,
            j.skill_nice_have, j.main_technologies, j.benefits, j.experience_level, j.employment_type,
            j.location, j.work_mode, j.application_url, j.is_active, j.signature, j.created_at, j.updated_at,
            c.name as company_name,
            COUNT(*) OVER() as total_count
        FROM jobs j
        JOIN companies c ON j.company_id = c.id, search_query sq
        WHERE j.is_active = true AND j.search_vector @@ sq.query
    `

	// Count-only search query
	searchJobsCountOnlyBaseQuery = `
        WITH search_query AS (
            SELECT plainto_tsquery('english', $1) AS query
        )
        SELECT COUNT(*)
        FROM jobs j
        JOIN companies c ON j.company_id = c.id, search_query sq
        WHERE j.is_active = true AND j.search_vector @@ sq.query
    `
)

// Constants for pagination
const (
	// Default pagination limit for search requests. Can be overridden by clients.
	DefaultLimit = 20
	MaxLimit     = 100
)

// Database interface to support pgxpool and mocks
type Database interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

// Repository handles database operations for the Job model.
type Repository struct {
	db Database
}

// NewRepository creates a new Repository instance.
func NewRepository(db Database) *Repository {
	return &Repository{db: db}
}

// SearchJobsWithCount performs a full-text search and returns both results and total count
func (r *Repository) SearchJobsWithCount(ctx context.Context, params *SearchParams) ([]*JobWithCompany, int, error) {
	// Build WHERE conditions and arguments
	additionalWhere, args := r.buildWhereConditions(params)

	// Build final search query with ordering and pagination
	argCount := len(args) + 1 // Next argument position
	searchQuery := searchJobsWithCountBaseQuery + additionalWhere +
		fmt.Sprintf(" ORDER BY j.created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)

	// Add pagination parameters
	args = append(args, params.Limit, params.Offset)

	// Execute search query
	rows, err := r.db.Query(ctx, searchQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*JobWithCompany
	var total int

	for rows.Next() {
		job := &JobWithCompany{}
		err = rows.Scan(
			&job.ID,
			&job.CompanyID,
			&job.Title,
			&job.Description,
			&job.Responsibilities,
			&job.SkillMustHave,
			&job.SkillNiceHave,
			&job.MainTechnologies,
			&job.Benefits,
			&job.ExperienceLevel,
			&job.EmploymentType,
			&job.Location,
			&job.WorkMode,
			&job.ApplicationURL,
			&job.IsActive,
			&job.Signature,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.CompanyName,
			&total, // Window function gives us the same total for each row
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job row: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating job rows: %w", err)
	}

	// If no results, total should be 0
	if len(jobs) == 0 {
		total = 0
	}

	return jobs, total, nil
}

// GetSearchCount performs a full-text search and returns only the total count
func (r *Repository) GetSearchCount(ctx context.Context, params *SearchParams) (int, error) {
	// Build WHERE conditions and arguments
	additionalWhere, args := r.buildWhereConditions(params)

	// Build final count query
	countQuery := searchJobsCountOnlyBaseQuery + additionalWhere

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get search count: %w", err)
	}

	return total, nil
}

// Create inserts a new job into the database.
func (r *Repository) Create(ctx context.Context, job *Job) error {
	err := r.db.QueryRow(
		ctx,
		createJobQuery,
		job.CompanyID,
		job.Title,
		job.Description,
		job.Responsibilities,
		job.SkillMustHave,
		job.SkillNiceHave,
		job.MainTechnologies,
		job.Benefits,
		job.ExperienceLevel,
		job.EmploymentType,
		job.Location,
		job.WorkMode,
		job.ApplicationURL,
		job.IsActive,
		job.Signature,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation (duplicate job signature)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Signature: job.Signature}
		}
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetByID retrieves a job by its ID.
func (r *Repository) GetByID(ctx context.Context, id int) (*Job, error) {
	job := &Job{}
	err := r.db.QueryRow(ctx, getJobByIDQuery, id).Scan(
		&job.ID,
		&job.CompanyID,
		&job.Title,
		&job.Description,
		&job.Responsibilities,
		&job.SkillMustHave,
		&job.SkillNiceHave,
		&job.MainTechnologies,
		&job.Benefits,
		&job.ExperienceLevel,
		&job.EmploymentType,
		&job.Location,
		&job.WorkMode,
		&job.ApplicationURL,
		&job.IsActive,
		&job.Signature,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{ID: id}
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

// Update updates an existing job in the database.
func (r *Repository) Update(ctx context.Context, job *Job) error {
	err := r.db.QueryRow(
		ctx,
		updateJobQuery,
		job.CompanyID,
		job.Title,
		job.Description,
		job.Responsibilities,
		job.SkillMustHave,
		job.SkillNiceHave,
		job.MainTechnologies,
		job.Benefits,
		job.ExperienceLevel,
		job.EmploymentType,
		job.Location,
		job.WorkMode,
		job.ApplicationURL,
		job.IsActive,
		job.Signature,
		job.ID,
	).Scan(&job.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &NotFoundError{ID: job.ID}
		}

		// Check for unique constraint violation (duplicate job signature)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Signature: job.Signature}
		}

		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// Delete removes a job from the database.
func (r *Repository) Delete(ctx context.Context, id int) error {
	commandTag, err := r.db.Exec(ctx, deleteJobQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{ID: id}
	}

	return nil
}

// GetBySignature retrieves a job by its signature.
func (r *Repository) GetBySignature(ctx context.Context, signature string) (*Job, error) {
	job := &Job{}
	err := r.db.QueryRow(ctx, getJobBySignatureQuery, signature).Scan(
		&job.ID,
		&job.CompanyID,
		&job.Title,
		&job.Description,
		&job.Responsibilities,
		&job.SkillMustHave,
		&job.SkillNiceHave,
		&job.MainTechnologies,
		&job.Benefits,
		&job.ExperienceLevel,
		&job.EmploymentType,
		&job.Location,
		&job.WorkMode,
		&job.ApplicationURL,
		&job.IsActive,
		&job.Signature,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Signature: signature}
		}
		return nil, fmt.Errorf("failed to get job by signature: %w", err)
	}

	return job, nil
}

// buildWhereConditions builds the WHERE conditions and arguments for search queries
//
//nolint:gocritic
func (r *Repository) buildWhereConditions(params *SearchParams) (string, []any) {
	// Trim whitespace from query
	params.Query = strings.TrimSpace(params.Query)

	whereConditions := []string{}
	args := []any{params.Query}
	argCount := 2 // Starting at 2 because $1 is the search query

	// Add optional filters
	if params.ExperienceLevel != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("j.experience_level = $%d", argCount))
		args = append(args, *params.ExperienceLevel)
		argCount++
	}

	if params.EmploymentType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("j.employment_type = $%d", argCount))
		args = append(args, *params.EmploymentType)
		argCount++
	}

	if params.Location != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("j.location = $%d", argCount))
		args = append(args, *params.Location)
		argCount++
	}

	if params.WorkMode != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("j.work_mode = $%d", argCount))
		args = append(args, *params.WorkMode)
		argCount++
	}

	if params.Company != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(c.name) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+*params.Company+"%")
		argCount++
	}

	if params.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("j.created_at >= $%d", argCount))
		args = append(args, *params.DateFrom)
		argCount++
	}

	if params.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("j.created_at <= $%d", argCount))
		args = append(args, *params.DateTo)
	}

	// Build additional WHERE clause
	additionalWhere := ""
	if len(whereConditions) > 0 {
		additionalWhere = " AND " + strings.Join(whereConditions, " AND ")
	}

	return additionalWhere, args
}
