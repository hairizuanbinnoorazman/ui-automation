package testrun

import (
	"context"

	"github.com/google/uuid"
)

// StepNoteStore defines the interface for step note persistence operations.
type StepNoteStore interface {
	// Upsert creates or updates a step note for a given (test_run_id, step_index).
	Upsert(ctx context.Context, note *StepNote) error

	// ListByTestRun retrieves all step notes for a specific test run, ordered by step_index.
	ListByTestRun(ctx context.Context, testRunID uuid.UUID) ([]*StepNote, error)

	// GetByRunAndStep retrieves a step note for a specific run and step index.
	GetByRunAndStep(ctx context.Context, testRunID uuid.UUID, stepIndex int) (*StepNote, error)
}
