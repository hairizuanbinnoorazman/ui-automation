package testprocedure

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for test procedure persistence operations.
type Store interface {
	// Create creates a new test procedure in the store.
	Create(ctx context.Context, testProcedure *TestProcedure) error

	// GetByID retrieves a test procedure by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*TestProcedure, error)

	// Update updates a test procedure with the given setters (in-place, doesn't create version).
	Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error

	// Delete deletes a test procedure (hard delete due to CASCADE).
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByProject retrieves a paginated list of latest test procedures for a specific project.
	ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*TestProcedure, error)

	// CreateVersion creates a new version of an existing test procedure.
	// This creates an immutable copy with incremented version number.
	CreateVersion(ctx context.Context, originalID uuid.UUID) (*TestProcedure, error)

	// GetVersionHistory retrieves all versions of a test procedure.
	GetVersionHistory(ctx context.Context, testProcedureID uuid.UUID) ([]*TestProcedure, error)
}

// UpdateSetter is a function that updates a test procedure field.
type UpdateSetter func(*TestProcedure) error
