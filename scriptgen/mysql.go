package scriptgen

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

// NewMySQLStore creates a new MySQL-backed generated script store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new generated script record in the database.
func (s *MySQLStore) Create(ctx context.Context, script *GeneratedScript) error {
	// Ensure default status is set before validation
	if script.GenerationStatus == "" {
		script.GenerationStatus = StatusPending
	}

	if err := script.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(script).Error; err != nil {
		// Check for unique constraint violation (MySQL and SQLite)
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			s.logger.Warn(ctx, "script already exists for procedure and framework", map[string]interface{}{
				"test_procedure_id": script.TestProcedureID.String(),
				"framework":         script.Framework,
			})
			return ErrScriptAlreadyExists
		}

		s.logger.Error(ctx, "failed to create generated script", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": script.TestProcedureID.String(),
			"framework":         script.Framework,
		})
		return err
	}

	s.logger.Info(ctx, "generated script created", map[string]interface{}{
		"script_id":         script.ID.String(),
		"test_procedure_id": script.TestProcedureID.String(),
		"framework":         script.Framework,
	})

	return nil
}

// GetByID retrieves a script by its ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uuid.UUID) (*GeneratedScript, error) {
	var script GeneratedScript
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&script).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScriptNotFound
		}
		s.logger.Error(ctx, "failed to get script by ID", map[string]interface{}{
			"error":     err.Error(),
			"script_id": id.String(),
		})
		return nil, err
	}

	return &script, nil
}

// GetByProcedureAndFramework retrieves a script by procedure ID and framework.
func (s *MySQLStore) GetByProcedureAndFramework(ctx context.Context, procedureID uuid.UUID, framework Framework) (*GeneratedScript, error) {
	var script GeneratedScript
	err := s.db.WithContext(ctx).
		Where("test_procedure_id = ? AND framework = ?", procedureID, framework).
		First(&script).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScriptNotFound
		}
		s.logger.Error(ctx, "failed to get script by procedure and framework", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
			"framework":         framework,
		})
		return nil, err
	}

	return &script, nil
}

// ListByProcedure retrieves all scripts for a test procedure.
func (s *MySQLStore) ListByProcedure(ctx context.Context, procedureID uuid.UUID) ([]*GeneratedScript, error) {
	var scripts []*GeneratedScript
	err := s.db.WithContext(ctx).
		Where("test_procedure_id = ?", procedureID).
		Order("generated_at DESC").
		Find(&scripts).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list scripts by procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
		})
		return nil, err
	}

	return scripts, nil
}

// Update updates a script with the given setters.
func (s *MySQLStore) Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error {
	// First, fetch the script
	script, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply all setters
	for _, setter := range setters {
		if err := setter(script); err != nil {
			return err
		}
	}

	// Save the updated script
	if err := s.db.WithContext(ctx).Save(script).Error; err != nil {
		s.logger.Error(ctx, "failed to update script", map[string]interface{}{
			"error":     err.Error(),
			"script_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "script updated", map[string]interface{}{
		"script_id": id.String(),
	})

	return nil
}

// Delete deletes a script by its ID.
func (s *MySQLStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&GeneratedScript{})

	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete script", map[string]interface{}{
			"error":     result.Error.Error(),
			"script_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrScriptNotFound
	}

	s.logger.Info(ctx, "script deleted", map[string]interface{}{
		"script_id": id.String(),
	})

	return nil
}
