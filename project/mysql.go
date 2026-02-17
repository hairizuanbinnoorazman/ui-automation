package project

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

// NewMySQLStore creates a new MySQL-backed project store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new project in the database.
func (s *MySQLStore) Create(ctx context.Context, project *Project) error {
	if err := project.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(project).Error; err != nil {
		s.logger.Error(ctx, "failed to create project", map[string]interface{}{
			"error":    err.Error(),
			"name":     project.Name,
			"owner_id": project.OwnerID.String(),
		})
		return err
	}

	s.logger.Info(ctx, "project created", map[string]interface{}{
		"project_id": project.ID.String(),
		"name":       project.Name,
		"owner_id":   project.OwnerID.String(),
	})

	return nil
}

// GetByID retrieves a project by its ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	var project Project
	err := s.db.WithContext(ctx).
		Where("id = ? AND is_active = ?", id, true).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		s.logger.Error(ctx, "failed to get project by ID", map[string]interface{}{
			"error":      err.Error(),
			"project_id": id.String(),
		})
		return nil, err
	}

	return &project, nil
}

// Update updates a project with the given setters.
func (s *MySQLStore) Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error {
	// First, fetch the project
	project, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply all setters
	for _, setter := range setters {
		if err := setter(project); err != nil {
			return err
		}
	}

	// Save the updated project
	if err := s.db.WithContext(ctx).Save(project).Error; err != nil {
		s.logger.Error(ctx, "failed to update project", map[string]interface{}{
			"error":      err.Error(),
			"project_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "project updated", map[string]interface{}{
		"project_id": id.String(),
	})

	return nil
}

// Delete soft deletes a project by setting is_active to false.
func (s *MySQLStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Model(&Project{}).
		Where("id = ? AND is_active = ?", id, true).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete project", map[string]interface{}{
			"error":      result.Error.Error(),
			"project_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	s.logger.Info(ctx, "project deleted", map[string]interface{}{
		"project_id": id.String(),
	})

	return nil
}

// ListByOwner retrieves a paginated list of active projects for a specific owner.
func (s *MySQLStore) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Project, error) {
	var projects []*Project
	err := s.db.WithContext(ctx).
		Where("owner_id = ? AND is_active = ?", ownerID, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list projects by owner", map[string]interface{}{
			"error":    err.Error(),
			"owner_id": ownerID.String(),
			"limit":    limit,
			"offset":   offset,
		})
		return nil, err
	}

	return projects, nil
}

// CountByOwner returns the total count of active projects for a specific owner.
func (s *MySQLStore) CountByOwner(ctx context.Context, ownerID uuid.UUID) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&Project{}).
		Where("owner_id = ? AND is_active = ?", ownerID, true).
		Count(&count).Error

	if err != nil {
		s.logger.Error(ctx, "failed to count projects by owner", map[string]interface{}{
			"error":    err.Error(),
			"owner_id": ownerID.String(),
		})
		return 0, err
	}

	return int(count), nil
}
