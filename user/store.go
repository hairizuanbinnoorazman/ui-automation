package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")

	// ErrDuplicateEmail is returned when attempting to create a user with an existing email.
	ErrDuplicateEmail = errors.New("email already exists")
)

// Store defines the interface for user persistence operations.
type Store interface {
	// Create creates a new user in the store.
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update updates a user with the given setters.
	Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error

	// Delete soft deletes a user by setting is_active to false.
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a paginated list of active users.
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// UpdateSetter is a function that updates a user field.
type UpdateSetter func(*User) error
