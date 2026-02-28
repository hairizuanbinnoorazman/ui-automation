package endpoint

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for endpoint persistence operations.
type Store interface {
	// Create creates a new endpoint in the store.
	Create(ctx context.Context, endpoint *Endpoint) error

	// GetByID retrieves an endpoint by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Endpoint, error)

	// Update updates an endpoint with the given setters.
	Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error

	// Delete deletes an endpoint (hard delete).
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByCreator retrieves a paginated list of endpoints for a specific creator.
	ListByCreator(ctx context.Context, createdBy uuid.UUID, limit, offset int) ([]*Endpoint, error)

	// CountByCreator returns the total count of endpoints for a specific creator.
	CountByCreator(ctx context.Context, createdBy uuid.UUID) (int, error)
}

// UpdateSetter is a function that updates an endpoint field.
type UpdateSetter func(*Endpoint) error
