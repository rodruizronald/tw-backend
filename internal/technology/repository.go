package technology

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rodruizronald/ticos-in-tech/internal/jobtech"
	"github.com/rodruizronald/ticos-in-tech/internal/techalias"
)

// SQL query constants
const (
	createTechnologyQuery = `
        INSERT INTO technologies (name, category, parent_id)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `

	getTechnologyByIDQuery = `
        SELECT id, name, category, parent_id, created_at
        FROM technologies
        WHERE id = $1
    `

	getTechnologyByNameQuery = `
        SELECT id, name, category, parent_id, created_at
        FROM technologies
        WHERE LOWER(name) = LOWER($1)
    `

	updateTechnologyQuery = `
        UPDATE technologies
        SET name = $1, category = $2, parent_id = $3
        WHERE id = $4
    `

	deleteTechnologyQuery = `DELETE FROM technologies WHERE id = $1`

	getTechnologyAliasesQuery = `
        SELECT id, technology_id, alias, created_at
        FROM technology_aliases
        WHERE technology_id = $1
        ORDER BY alias
    `

	getTechnologyJobsQuery = `
        SELECT id, job_id, technology_id, is_primary, is_required, created_at
        FROM job_technologies
        WHERE technology_id = $1
        ORDER BY created_at DESC
    `
)

// Database interface to support pgxpool and mocks
type Database interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

// Repository handles database operations for the Technology model.
type Repository struct {
	db Database
}

// NewRepository creates a new Repository instance.
func NewRepository(db Database) *Repository {
	return &Repository{db: db}
}

// Create inserts a new technology into the database.
func (r *Repository) Create(ctx context.Context, tech *Technology) error {
	err := r.db.QueryRow(
		ctx,
		createTechnologyQuery,
		tech.Name,
		tech.Category,
		tech.ParentID,
	).Scan(&tech.ID, &tech.CreatedAt)

	if err != nil {
		// Check for unique constraint violation (duplicate technology name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Name: tech.Name}
		}
		return fmt.Errorf("failed to create technology: %w", err)
	}

	return nil
}

// GetByID retrieves a technology by its ID.
func (r *Repository) GetByID(ctx context.Context, id int) (*Technology, error) {
	tech := &Technology{}
	err := r.db.QueryRow(ctx, getTechnologyByIDQuery, id).Scan(
		&tech.ID,
		&tech.Name,
		&tech.Category,
		&tech.ParentID,
		&tech.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{ID: id}
		}
		return nil, fmt.Errorf("failed to get technology: %w", err)
	}

	return tech, nil
}

// GetByName retrieves a technology by its name.
func (r *Repository) GetByName(ctx context.Context, name string) (*Technology, error) {
	tech := &Technology{}
	err := r.db.QueryRow(ctx, getTechnologyByNameQuery, name).Scan(
		&tech.ID,
		&tech.Name,
		&tech.Category,
		&tech.ParentID,
		&tech.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Name: name}
		}
		return nil, fmt.Errorf("failed to get technology: %w", err)
	}

	return tech, nil
}

// Update updates an existing technology in the database.
func (r *Repository) Update(ctx context.Context, tech *Technology) error {
	commandTag, err := r.db.Exec(
		ctx,
		updateTechnologyQuery,
		tech.Name,
		tech.Category,
		tech.ParentID,
		tech.ID,
	)

	if err != nil {
		// Check for unique constraint violation (duplicate technology name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Name: tech.Name}
		}
		return fmt.Errorf("failed to update technology: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{ID: tech.ID}
	}

	return nil
}

// Delete removes a technology from the database.
func (r *Repository) Delete(ctx context.Context, id int) error {
	commandTag, err := r.db.Exec(ctx, deleteTechnologyQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete technology: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{ID: id}
	}

	return nil
}

// GetWithAliases retrieves a technology by ID including its aliases.
func (r *Repository) GetWithAliases(ctx context.Context, id int) (*Technology, error) {
	tech, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, getTechnologyAliasesQuery, tech.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get technology aliases: %w", err)
	}
	defer rows.Close()

	var aliases []techalias.TechnologyAlias
	for rows.Next() {
		alias := techalias.TechnologyAlias{}
		err = rows.Scan(
			&alias.ID,
			&alias.TechnologyID,
			&alias.Alias,
			&alias.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alias row: %w", err)
		}
		aliases = append(aliases, alias)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alias rows: %w", err)
	}

	tech.Aliases = aliases
	return tech, nil
}

// GetWithJobs retrieves a technology by ID including its job associations.
func (r *Repository) GetWithJobs(ctx context.Context, id int) (*Technology, error) {
	tech, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, getTechnologyJobsQuery, tech.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get technology jobs: %w", err)
	}
	defer rows.Close()

	var jobs []jobtech.JobTechnology
	for rows.Next() {
		job := jobtech.JobTechnology{}
		err = rows.Scan(
			&job.ID,
			&job.JobID,
			&job.TechnologyID,
			&job.IsRequired,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job technology row: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating job technology rows: %w", err)
	}

	tech.Jobs = jobs
	return tech, nil
}
