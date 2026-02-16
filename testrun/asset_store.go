package testrun

import (
	"context"
)

// AssetStore defines the interface for test run asset persistence operations.
type AssetStore interface {
	// Create creates a new asset in the store.
	Create(ctx context.Context, asset *TestRunAsset) error

	// GetByID retrieves an asset by its ID.
	GetByID(ctx context.Context, id uint) (*TestRunAsset, error)

	// ListByTestRun retrieves all assets for a specific test run.
	ListByTestRun(ctx context.Context, testRunID uint) ([]*TestRunAsset, error)

	// Delete deletes an asset by ID.
	Delete(ctx context.Context, id uint) error
}
