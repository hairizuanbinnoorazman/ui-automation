package endpoint

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStore implements the Store interface using GORM and MySQL.
type MySQLStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStore creates a new MySQL-backed endpoint store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new endpoint in the database.
func (s *MySQLStore) Create(ctx context.Context, endpoint *Endpoint) error {
	if err := endpoint.Validate(); err != nil {
		return err
	}

	if endpoint.Credentials == nil || len(endpoint.Credentials) == 0 {
		endpoint.Credentials = DefaultCredentials()
	}

	result := s.db.WithContext(ctx).Create(endpoint)
	if result.Error != nil {
		s.logger.Error(ctx, "failed to create endpoint", map[string]interface{}{
			"error": result.Error.Error(),
		})
		return result.Error
	}

	return nil
}

// GetByID retrieves an endpoint by its ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uuid.UUID) (*Endpoint, error) {
	var ep Endpoint
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&ep).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEndpointNotFound
		}
		s.logger.Error(ctx, "failed to get endpoint by ID", map[string]interface{}{
			"error":       err.Error(),
			"endpoint_id": id.String(),
		})
		return nil, err
	}

	return &ep, nil
}

// Update updates an endpoint with the given setters.
func (s *MySQLStore) Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error {
	ep, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	for _, setter := range setters {
		if err := setter(ep); err != nil {
			return err
		}
	}

	if err := s.db.WithContext(ctx).Save(ep).Error; err != nil {
		s.logger.Error(ctx, "failed to update endpoint", map[string]interface{}{
			"error":       err.Error(),
			"endpoint_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "endpoint updated", map[string]interface{}{
		"endpoint_id": id.String(),
	})

	return nil
}

// Delete deletes an endpoint (hard delete).
func (s *MySQLStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&Endpoint{})

	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete endpoint", map[string]interface{}{
			"error":       result.Error.Error(),
			"endpoint_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrEndpointNotFound
	}

	s.logger.Info(ctx, "endpoint deleted", map[string]interface{}{
		"endpoint_id": id.String(),
	})

	return nil
}

// ListByCreator retrieves a paginated list of endpoints for a specific creator.
func (s *MySQLStore) ListByCreator(ctx context.Context, createdBy uuid.UUID, limit, offset int) ([]*Endpoint, error) {
	var endpoints []*Endpoint
	err := s.db.WithContext(ctx).
		Where("created_by = ?", createdBy).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&endpoints).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list endpoints by creator", map[string]interface{}{
			"error":      err.Error(),
			"created_by": createdBy.String(),
			"limit":      limit,
			"offset":     offset,
		})
		return nil, err
	}

	return endpoints, nil
}

// CountByCreator returns the total count of endpoints for a specific creator.
func (s *MySQLStore) CountByCreator(ctx context.Context, createdBy uuid.UUID) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&Endpoint{}).
		Where("created_by = ?", createdBy).
		Count(&count).Error

	if err != nil {
		s.logger.Error(ctx, "failed to count endpoints by creator", map[string]interface{}{
			"error":      err.Error(),
			"created_by": createdBy.String(),
		})
		return 0, err
	}

	return int(count), nil
}
