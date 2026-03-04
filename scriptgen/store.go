package scriptgen

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for generated script persistence.
type Store interface {
	// Create creates a new generated script record.
	Create(ctx context.Context, script *GeneratedScript) error

	// GetByID retrieves a script by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*GeneratedScript, error)

	// GetByProcedureAndFramework retrieves a script by procedure ID and framework.
	GetByProcedureAndFramework(ctx context.Context, procedureID uuid.UUID, framework Framework) (*GeneratedScript, error)

	// ListByProcedure retrieves all scripts for a test procedure.
	ListByProcedure(ctx context.Context, procedureID uuid.UUID) ([]*GeneratedScript, error)

	// Update updates a script with setter functions.
	Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error

	// Delete deletes a script by its ID.
	Delete(ctx context.Context, id uuid.UUID) error
}

// UpdateSetter returns the column-value pairs to apply in a partial UPDATE.
// Using a map avoids a read-modify-write: the caller never needs to fetch
// the full row before writing.
type UpdateSetter func() map[string]interface{}
