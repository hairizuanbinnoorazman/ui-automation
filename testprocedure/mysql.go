package testprocedure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStore implements the Store interface using GORM and MySQL.
type MySQLStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStore creates a new MySQL-backed test procedure store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new test procedure in the database.
func (s *MySQLStore) Create(ctx context.Context, testProcedure *TestProcedure) error {
	if err := testProcedure.Validate(); err != nil {
		return err
	}

	// Ensure initial version values
	testProcedure.Version = 1
	testProcedure.IsLatest = true
	testProcedure.ParentID = nil

	if err := s.db.WithContext(ctx).Create(testProcedure).Error; err != nil {
		s.logger.Error(ctx, "failed to create test procedure", map[string]interface{}{
			"error":      err.Error(),
			"name":       testProcedure.Name,
			"project_id": testProcedure.ProjectID,
		})
		return err
	}

	s.logger.Info(ctx, "test procedure created", map[string]interface{}{
		"test_procedure_id": testProcedure.ID.String(),
		"name":              testProcedure.Name,
		"project_id":        testProcedure.ProjectID.String(),
	})

	return nil
}

// GetByID retrieves a test procedure by its ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uuid.UUID) (*TestProcedure, error) {
	var testProcedure TestProcedure
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&testProcedure).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTestProcedureNotFound
		}
		s.logger.Error(ctx, "failed to get test procedure by ID", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id.String(),
		})
		return nil, err
	}

	return &testProcedure, nil
}

// Update updates a test procedure with the given setters (in-place, doesn't create version).
func (s *MySQLStore) Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error {
	// First, fetch the test procedure
	testProcedure, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply all setters
	for _, setter := range setters {
		if err := setter(testProcedure); err != nil {
			return err
		}
	}

	// Save the updated test procedure
	if err := s.db.WithContext(ctx).Save(testProcedure).Error; err != nil {
		s.logger.Error(ctx, "failed to update test procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "test procedure updated", map[string]interface{}{
		"test_procedure_id": id.String(),
	})

	return nil
}

// Delete deletes a test procedure (hard delete due to CASCADE).
func (s *MySQLStore) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&TestProcedure{})

	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete test procedure", map[string]interface{}{
			"error":             result.Error.Error(),
			"test_procedure_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrTestProcedureNotFound
	}

	s.logger.Info(ctx, "test procedure deleted", map[string]interface{}{
		"test_procedure_id": id.String(),
	})

	return nil
}

// ListByProject retrieves a paginated list of latest test procedures for a specific project.
func (s *MySQLStore) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*TestProcedure, error) {
	var testProcedures []*TestProcedure
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND is_latest = ?", projectID, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&testProcedures).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list test procedures by project", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID.String(),
			"limit":      limit,
			"offset":     offset,
		})
		return nil, err
	}

	return testProcedures, nil
}

// CountByProject returns the total count of latest test procedures for a specific project.
func (s *MySQLStore) CountByProject(ctx context.Context, projectID uuid.UUID) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&TestProcedure{}).
		Where("project_id = ? AND is_latest = ?", projectID, true).
		Count(&count).Error

	if err != nil {
		s.logger.Error(ctx, "failed to count test procedures by project", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID.String(),
		})
		return 0, err
	}

	return int(count), nil
}

// CreateVersion creates a new version of an existing test procedure.
// This creates an immutable copy with incremented version number.
func (s *MySQLStore) CreateVersion(ctx context.Context, originalID uuid.UUID) (*TestProcedure, error) {
	var newVersion *TestProcedure

	// Execute in transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Load original test procedure
		original, err := s.getByIDWithTx(ctx, tx, originalID)
		if err != nil {
			return err
		}

		// 2. Determine root ID (original.ParentID ?? original.ID)
		rootID := originalID
		if original.ParentID != nil {
			rootID = *original.ParentID
		}

		// 3. Mark all versions in chain as is_latest=false
		if err := tx.WithContext(ctx).
			Model(&TestProcedure{}).
			Where("id = ? OR parent_id = ?", rootID, rootID).
			Update("is_latest", false).Error; err != nil {
			return fmt.Errorf("failed to update is_latest flags: %w", err)
		}

		// 4. Query max version number in chain
		var maxVersion uint
		err = tx.WithContext(ctx).
			Model(&TestProcedure{}).
			Where("id = ? OR parent_id = ?", rootID, rootID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&maxVersion).Error
		if err != nil {
			return fmt.Errorf("failed to get max version: %w", err)
		}

		// 5. Create new record with version=max+1, is_latest=true, parent_id=root
		newVersion = &TestProcedure{
			ProjectID:   original.ProjectID,
			Name:        original.Name,
			Description: original.Description,
			Steps:       original.Steps,
			CreatedBy:   original.CreatedBy,
			Version:     maxVersion + 1,
			IsLatest:    true,
			ParentID:    &rootID,
		}

		if err := tx.WithContext(ctx).Create(newVersion).Error; err != nil {
			return fmt.Errorf("failed to create new version: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "failed to create test procedure version", map[string]interface{}{
			"error":       err.Error(),
			"original_id": originalID.String(),
		})
		return nil, err
	}

	s.logger.Info(ctx, "test procedure version created", map[string]interface{}{
		"new_version_id": newVersion.ID.String(),
		"version":        newVersion.Version,
		"original_id":    originalID.String(),
	})

	return newVersion, nil
}

// GetVersionHistory retrieves all versions of a test procedure.
func (s *MySQLStore) GetVersionHistory(ctx context.Context, testProcedureID uuid.UUID) ([]*TestProcedure, error) {
	// First get the test procedure to determine root ID
	testProcedure, err := s.GetByID(ctx, testProcedureID)
	if err != nil {
		return nil, err
	}

	// Determine root ID
	rootID := testProcedureID
	if testProcedure.ParentID != nil {
		rootID = *testProcedure.ParentID
	}

	// Get all versions in the chain
	var versions []*TestProcedure
	err = s.db.WithContext(ctx).
		Where("id = ? OR parent_id = ?", rootID, rootID).
		Order("version DESC").
		Find(&versions).Error

	if err != nil {
		s.logger.Error(ctx, "failed to get version history", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": testProcedureID.String(),
		})
		return nil, err
	}

	return versions, nil
}

// getByIDWithTx is a helper to get by ID within a transaction.
func (s *MySQLStore) getByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*TestProcedure, error) {
	var testProcedure TestProcedure
	err := tx.WithContext(ctx).
		Where("id = ?", id).
		First(&testProcedure).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTestProcedureNotFound
		}
		return nil, err
	}

	return &testProcedure, nil
}
