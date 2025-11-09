package technology

import (
	"time"

	"github.com/rodruizronald/tw-backend/internal/jobtech"
	"github.com/rodruizronald/tw-backend/internal/techalias"
)

// Technology represents a technology skill (programming language, framework, tool, etc.)
// used in job postings.
type Technology struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Category  string    `json:"category" db:"category"`
	ParentID  *int      `json:"parent_id,omitempty" db:"parent_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relationships (not stored in database)
	Aliases []techalias.TechnologyAlias `json:"aliases,omitempty" db:"-"`
	Jobs    []jobtech.JobTechnology     `json:"jobs,omitempty" db:"-"`
}
