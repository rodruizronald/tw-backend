package company

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rodruizronald/tw-backend/internal/jobs"
)

// SQL query constants
const (
	createCompanyQuery = `
        INSERT INTO companies (name, is_active)
        VALUES ($1, $2)
        RETURNING id
    `

	getCompanyByNameQuery = `
        SELECT id, name, is_active, created_at, updated_at
        FROM companies
        WHERE name = $1
    `

	updateCompanyQuery = `
        UPDATE companies
        SET name = $1, is_active = $2, updated_at = NOW()
        WHERE id = $3
        RETURNING updated_at
    `

	deleteCompanyQuery = `DELETE FROM companies WHERE id = $1`

	listCompaniesQuery = `
        SELECT id, name, is_active, created_at, updated_at
        FROM companies
        ORDER BY name
    `

	getCompanyJobsQuery = `
        SELECT id, company_id, title, description, experience_level, employment_type,
               location, work_mode, application_url, is_active, signature, created_at, updated_at
        FROM jobs
        WHERE company_id = $1 AND is_active = true
        ORDER BY created_at DESC
    `
)

// Database interface to support pgxpool and mocks
type Database interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

// Repository handles database operations for the Company model.
type Repository struct {
	db Database
}

// NewRepository creates a new Repository instance.
func NewRepository(db Database) *Repository {
	return &Repository{db: db}
}

// Create inserts a new company into the database.
func (r *Repository) Create(ctx context.Context, company *Company) error {
	err := r.db.QueryRow(
		ctx,
		createCompanyQuery,
		company.Name,
		company.IsActive,
	).Scan(&company.ID)

	if err != nil {
		// Check for unique constraint violation (duplicate company name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Name: company.Name}
		}
		return fmt.Errorf("failed to create company: %w", err)
	}

	return nil
}

// GetByName retrieves a company by its name.
func (r *Repository) GetByName(ctx context.Context, name string) (*Company, error) {
	company := &Company{}
	err := r.db.QueryRow(ctx, getCompanyByNameQuery, name).Scan(
		&company.ID,
		&company.Name,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Name: name}
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return company, nil
}

// Update updates an existing company in the database.
func (r *Repository) Update(ctx context.Context, company *Company) error {
	err := r.db.QueryRow(
		ctx,
		updateCompanyQuery,
		company.Name,
		company.IsActive,
		company.ID,
	).Scan(&company.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &NotFoundError{ID: company.ID}
		}

		// Check for unique constraint violation (duplicate company name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Name: company.Name}
		}

		return fmt.Errorf("failed to update company: %w", err)
	}

	return nil
}

// Delete removes a company from the database.
func (r *Repository) Delete(ctx context.Context, id int) error {
	commandTag, err := r.db.Exec(ctx, deleteCompanyQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete company: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{ID: id}
	}

	return nil
}

// List retrieves all companies from the database.
func (r *Repository) List(ctx context.Context) ([]*Company, error) {
	rows, err := r.db.Query(ctx, listCompaniesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}
	defer rows.Close()

	var companies []*Company
	for rows.Next() {
		company := &Company{}
		err = rows.Scan(
			&company.ID,
			&company.Name,
			&company.IsActive,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company row: %w", err)
		}
		companies = append(companies, company)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating company rows: %w", err)
	}

	return companies, nil
}

// GetWithJobs retrieves a company by name including its jobs.
func (r *Repository) GetWithJobs(ctx context.Context, name string) (*Company, error) {
	company, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, getCompanyJobsQuery, company.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company jobs: %w", err)
	}
	defer rows.Close()

	var gotJobs []jobs.Job
	for rows.Next() {
		gotJob := jobs.Job{}
		err = rows.Scan(
			&gotJob.ID,
			&gotJob.CompanyID,
			&gotJob.Title,
			&gotJob.Description,
			&gotJob.ExperienceLevel,
			&gotJob.EmploymentType,
			&gotJob.Location,
			&gotJob.WorkMode,
			&gotJob.ApplicationURL,
			&gotJob.IsActive,
			&gotJob.Signature,
			&gotJob.CreatedAt,
			&gotJob.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job row: %w", err)
		}
		gotJobs = append(gotJobs, gotJob)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating job rows: %w", err)
	}

	company.Jobs = gotJobs
	return company, nil
}
