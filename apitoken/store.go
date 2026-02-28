package apitoken

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for API token persistence operations.
type Store interface {
	// Create creates a new API token in the store.
	Create(ctx context.Context, token *APIToken) error

	// GetByID retrieves an API token by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*APIToken, error)

	// GetByTokenHash retrieves an active, non-expired token by its hash.
	GetByTokenHash(ctx context.Context, hash string) (*APIToken, error)

	// ListByUser retrieves active tokens for a user, ordered by created_at DESC.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*APIToken, error)

	// CountActiveByUser returns the count of active tokens for a user.
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// Revoke sets a token's is_active to false.
	Revoke(ctx context.Context, id uuid.UUID) error

	// Delete hard-deletes a token.
	Delete(ctx context.Context, id uuid.UUID) error
}
