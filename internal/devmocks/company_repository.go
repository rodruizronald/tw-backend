package devmocks

import (
	"context"

	"github.com/rodruizronald/tw-backend/internal/company"
)

// CompanyRepository provides mock data for company development and testing
type CompanyRepository struct {
	companies []*company.Company
	nextID    int
}

// NewCompanyRepository creates a new mock company repository with sample data
func NewCompanyRepository() *CompanyRepository {
	return &CompanyRepository{
		companies: []*company.Company{
			{
				ID:       1,
				Name:     "Tech Corp",
				IsActive: true,
			},
			{
				ID:       2,
				Name:     "Innovation Labs",
				IsActive: true,
			},
			{
				ID:       3,
				Name:     "Digital Solutions",
				IsActive: false,
			},
		},
		nextID: 4,
	}
}

// Create adds a new company to the mock repository
func (r *CompanyRepository) Create(_ context.Context, comp *company.Company) error {
	// Check for duplicate name
	for _, existing := range r.companies {
		if existing.Name == comp.Name {
			return &company.DuplicateError{Name: comp.Name}
		}
	}

	comp.ID = r.nextID
	r.nextID++
	r.companies = append(r.companies, comp)
	return nil
}

// Update modifies an existing company in the mock repository
func (r *CompanyRepository) Update(_ context.Context, comp *company.Company) error {
	for i, existing := range r.companies {
		if existing.ID == comp.ID {
			// Check for duplicate name (excluding current company)
			for _, other := range r.companies {
				if other.ID != comp.ID && other.Name == comp.Name {
					return &company.DuplicateError{Name: comp.Name}
				}
			}
			r.companies[i] = comp
			return nil
		}
	}
	return &company.NotFoundError{ID: comp.ID}
}

// Delete removes a company from the mock repository
func (r *CompanyRepository) Delete(_ context.Context, id int) error {
	for i, existing := range r.companies {
		if existing.ID == id {
			r.companies = append(r.companies[:i], r.companies[i+1:]...)
			return nil
		}
	}
	return &company.NotFoundError{ID: id}
}

// List returns all companies from the mock repository
func (r *CompanyRepository) List(_ context.Context) ([]*company.Company, error) {
	return r.companies, nil
}
