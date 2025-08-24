package techalias

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SQL query constants
const (
	createTechnologyAliasQuery = `
        INSERT INTO technology_aliases (technology_id, alias)
        VALUES ($1, $2)
        RETURNING id, created_at
    `

	getTechnologyAliasByIDQuery = `
        SELECT id, technology_id, alias, created_at
        FROM technology_aliases
        WHERE id = $1
    `

	getTechnologyAliasByAliasQuery = `
        SELECT id, technology_id, alias, created_at
        FROM technology_aliases
        WHERE LOWER(alias) = LOWER($1)
    `

	updateTechnologyAliasQuery = `
        UPDATE technology_aliases
        SET alias = $1
        WHERE id = $2
    `

	deleteTechnologyAliasQuery = `DELETE FROM technology_aliases WHERE id = $1`

	listTechnologyAliasesByTechnologyIDQuery = `
        SELECT id, technology_id, alias, created_at
        FROM technology_aliases
        WHERE technology_id = $1
        ORDER BY alias
    `
)

// Database interface to support pgxpool and mocks
type Database interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

// Repository handles database operations for the TechnologyAlias model.
type Repository struct {
	db Database
}

// NewRepository creates a new Repository instance.
func NewRepository(db Database) *Repository {
	return &Repository{db: db}
}

// Create inserts a new technology alias into the database.
func (r *Repository) Create(ctx context.Context, alias *TechnologyAlias) error {
	err := r.db.QueryRow(
		ctx,
		createTechnologyAliasQuery,
		alias.TechnologyID,
		alias.Alias,
	).Scan(&alias.ID, &alias.CreatedAt)

	if err != nil {
		// Check for unique constraint violation (duplicate alias)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Alias: alias.Alias}
		}
		return fmt.Errorf("failed to create technology alias: %w", err)
	}

	return nil
}

// GetByID retrieves a technology alias by its ID.
func (r *Repository) GetByID(ctx context.Context, id int) (*TechnologyAlias, error) {
	alias := &TechnologyAlias{}
	err := r.db.QueryRow(ctx, getTechnologyAliasByIDQuery, id).Scan(
		&alias.ID,
		&alias.TechnologyID,
		&alias.Alias,
		&alias.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{ID: id}
		}
		return nil, fmt.Errorf("failed to get technology alias: %w", err)
	}

	return alias, nil
}

// GetByAlias retrieves a technology alias by its alias value.
func (r *Repository) GetByAlias(ctx context.Context, aliasValue string) (*TechnologyAlias, error) {
	alias := &TechnologyAlias{}
	err := r.db.QueryRow(ctx, getTechnologyAliasByAliasQuery, aliasValue).Scan(
		&alias.ID,
		&alias.TechnologyID,
		&alias.Alias,
		&alias.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Alias: aliasValue}
		}
		return nil, fmt.Errorf("failed to get technology alias: %w", err)
	}

	return alias, nil
}

// Update updates an existing technology alias in the database.
func (r *Repository) Update(ctx context.Context, alias *TechnologyAlias) error {
	commandTag, err := r.db.Exec(
		ctx,
		updateTechnologyAliasQuery,
		alias.Alias,
		alias.ID,
	)

	if err != nil {
		// Check for unique constraint violation (duplicate alias)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateError{Alias: alias.Alias}
		}
		return fmt.Errorf("failed to update technology alias: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{ID: alias.ID}
	}

	return nil
}

// Delete removes a technology alias from the database.
func (r *Repository) Delete(ctx context.Context, id int) error {
	commandTag, err := r.db.Exec(ctx, deleteTechnologyAliasQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete technology alias: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{ID: id}
	}

	return nil
}

// ListByTechnologyID retrieves all aliases for a specific technology.
func (r *Repository) ListByTechnologyID(ctx context.Context, technologyID int) ([]*TechnologyAlias, error) {
	rows, err := r.db.Query(ctx, listTechnologyAliasesByTechnologyIDQuery, technologyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list technology aliases: %w", err)
	}
	defer rows.Close()

	var aliases []*TechnologyAlias
	for rows.Next() {
		alias := &TechnologyAlias{}
		err = rows.Scan(
			&alias.ID,
			&alias.TechnologyID,
			&alias.Alias,
			&alias.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan technology alias row: %w", err)
		}
		aliases = append(aliases, alias)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating technology alias rows: %w", err)
	}

	return aliases, nil
}
