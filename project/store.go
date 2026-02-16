package project

import (
	"context"
)

// Store defines the interface for project persistence operations.
type Store interface {
	// Create creates a new project in the store.
	Create(ctx context.Context, project *Project) error

	// GetByID retrieves a project by its ID.
	GetByID(ctx context.Context, id uint) (*Project, error)

	// Update updates a project with the given setters.
	Update(ctx context.Context, id uint, setters ...UpdateSetter) error

	// Delete soft deletes a project by setting is_active to false.
	Delete(ctx context.Context, id uint) error

	// ListByOwner retrieves a paginated list of active projects for a specific owner.
	ListByOwner(ctx context.Context, ownerID uint, limit, offset int) ([]*Project, error)
}

// UpdateSetter is a function that updates a project field.
type UpdateSetter func(*Project) error
