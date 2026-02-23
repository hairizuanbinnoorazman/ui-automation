package testrun

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for test run persistence operations.
type Store interface {
	// Create creates a new test run in the store.
	Create(ctx context.Context, testRun *TestRun) error

	// GetByID retrieves a test run by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*TestRun, error)

	// Update updates a test run with the given setters.
	Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error

	// ListByTestProcedure retrieves a paginated list of test runs for a specific test procedure.
	ListByTestProcedure(ctx context.Context, testProcedureID uuid.UUID, limit, offset int) ([]*TestRun, error)

	// CountByTestProcedure returns the total count of test runs for a specific test procedure.
	CountByTestProcedure(ctx context.Context, testProcedureID uuid.UUID) (int, error)

	// ListByTestProcedures retrieves a paginated list of test runs for multiple procedure versions.
	ListByTestProcedures(ctx context.Context, testProcedureIDs []uuid.UUID, limit, offset int) ([]*TestRun, error)

	// CountByTestProcedures returns the total count of test runs for multiple procedure versions.
	CountByTestProcedures(ctx context.Context, testProcedureIDs []uuid.UUID) (int, error)

	// Start marks a test run as started (sets started_at, changes status to running).
	Start(ctx context.Context, id uuid.UUID) error

	// Complete marks a test run as completed (sets completed_at, final status, optional notes).
	Complete(ctx context.Context, id uuid.UUID, status Status, notes string) error
}

// UpdateSetter is a function that updates a test run field.
type UpdateSetter func(*TestRun) error
