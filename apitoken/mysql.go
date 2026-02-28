package apitoken

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStore implements the Store interface using GORM and MySQL.
type MySQLStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStore creates a new MySQL-backed API token store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new API token in the database.
// Enforces the maximum tokens per user limit.
func (s *MySQLStore) Create(ctx context.Context, token *APIToken) error {
	if err := token.Validate(); err != nil {
		return err
	}

	// Check max tokens limit
	count, err := s.CountActiveByUser(ctx, token.UserID)
	if err != nil {
		return err
	}
	if count >= MaxTokensPerUser {
		return ErrMaxTokensReached
	}

	result := s.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		s.logger.Error(ctx, "failed to create api token", map[string]interface{}{
			"error": result.Error.Error(),
		})
		return result.Error
	}

	return nil
}

// GetByID retrieves an API token by its ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uuid.UUID) (*APIToken, error) {
	var token APIToken
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		s.logger.Error(ctx, "failed to get api token by ID", map[string]interface{}{
			"error":    err.Error(),
			"token_id": id.String(),
		})
		return nil, err
	}

	return &token, nil
}

// GetByTokenHash retrieves an active, non-expired token by its hash.
func (s *MySQLStore) GetByTokenHash(ctx context.Context, hash string) (*APIToken, error) {
	var token APIToken
	err := s.db.WithContext(ctx).
		Where("token_hash = ? AND is_active = ? AND expires_at > ?", hash, true, time.Now()).
		First(&token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		s.logger.Error(ctx, "failed to get api token by hash", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return &token, nil
}

// ListByUser retrieves active tokens for a user, ordered by created_at DESC.
func (s *MySQLStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]*APIToken, error) {
	var tokens []*APIToken
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at DESC").
		Find(&tokens).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list api tokens by user", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		return nil, err
	}

	return tokens, nil
}

// CountActiveByUser returns the count of active tokens for a user.
func (s *MySQLStore) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&APIToken{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&count).Error

	if err != nil {
		s.logger.Error(ctx, "failed to count active api tokens", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		return 0, err
	}

	return int(count), nil
}

// Revoke sets a token's is_active to false.
func (s *MySQLStore) Revoke(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Model(&APIToken{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error(ctx, "failed to revoke api token", map[string]interface{}{
			"error":    result.Error.Error(),
			"token_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}

	s.logger.Info(ctx, "api token revoked", map[string]interface{}{
		"token_id": id.String(),
	})

	return nil
}

// Delete hard-deletes a token.
func (s *MySQLStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&APIToken{})

	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete api token", map[string]interface{}{
			"error":    result.Error.Error(),
			"token_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}

	s.logger.Info(ctx, "api token deleted", map[string]interface{}{
		"token_id": id.String(),
	})

	return nil
}
