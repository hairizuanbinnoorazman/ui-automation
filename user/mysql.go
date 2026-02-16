package user

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStore implements the Store interface using GORM and MySQL.
type MySQLStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStore creates a new MySQL-backed user store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new user in the database.
func (s *MySQLStore) Create(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		// Check for duplicate key error (MySQL and SQLite)
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateEmail
		}
		s.logger.Error(ctx, "failed to create user", map[string]interface{}{
			"error": err.Error(),
			"email": user.Email,
		})
		return err
	}

	s.logger.Info(ctx, "user created", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})

	return nil
}

// GetByID retrieves a user by their ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := s.db.WithContext(ctx).
		Where("id = ? AND is_active = ?", id, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error(ctx, "failed to get user by ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id.String(),
		})
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email address.
func (s *MySQLStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.WithContext(ctx).
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error(ctx, "failed to get user by email", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		})
		return nil, err
	}

	return &user, nil
}

// Update updates a user with the given setters.
func (s *MySQLStore) Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error {
	// First, fetch the user
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply all setters
	for _, setter := range setters {
		if err := setter(user); err != nil {
			return err
		}
	}

	// Save the updated user
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		// Check for duplicate key error (MySQL and SQLite)
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateEmail
		}
		s.logger.Error(ctx, "failed to update user", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "user updated", map[string]interface{}{
		"user_id": id.String(),
	})

	return nil
}

// Delete soft deletes a user by setting is_active to false.
func (s *MySQLStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ? AND is_active = ?", id, true).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete user", map[string]interface{}{
			"error":   result.Error.Error(),
			"user_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	s.logger.Info(ctx, "user deleted", map[string]interface{}{
		"user_id": id.String(),
	})

	return nil
}

// List retrieves a paginated list of active users.
func (s *MySQLStore) List(ctx context.Context, limit, offset int) ([]*User, error) {
	var users []*User
	err := s.db.WithContext(ctx).
		Where("is_active = ?", true).
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list users", map[string]interface{}{
			"error":  err.Error(),
			"limit":  limit,
			"offset": offset,
		})
		return nil, err
	}

	return users, nil
}
