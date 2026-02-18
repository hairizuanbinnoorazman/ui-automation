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
// This now delegates to CreateWithDraft to automatically create both v1 and v0.
func (s *MySQLStore) Create(ctx context.Context, testProcedure *TestProcedure) error {
	result, err := s.CreateWithDraft(ctx, testProcedure)
	if err != nil {
		return err
	}

	// Copy the created v1 ID back to the original pointer
	testProcedure.ID = result.ID
	testProcedure.Version = result.Version
	testProcedure.IsLatest = result.IsLatest
	testProcedure.CreatedAt = result.CreatedAt
	testProcedure.UpdatedAt = result.UpdatedAt

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

// Update updates a test procedure with the given setters.
// This now delegates to UpdateDraft - it only updates the draft (v0), not the committed version.
func (s *MySQLStore) Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error {
	return s.UpdateDraft(ctx, id, setters...)
}

// Delete deletes all versions of a test procedure chain (hard delete due to CASCADE).
func (s *MySQLStore) Delete(ctx context.Context, id uuid.UUID) error {
	proc, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	rootID := id
	if proc.ParentID != nil {
		rootID = *proc.ParentID
	}

	result := s.db.WithContext(ctx).
		Where("id = ? OR parent_id = ?", rootID, rootID).
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
		"root_id":           rootID.String(),
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

// GetDraft retrieves the draft version (version 0) for a procedure.
func (s *MySQLStore) GetDraft(ctx context.Context, procedureID uuid.UUID) (*TestProcedure, error) {
	// First get the procedure to determine root ID
	proc, err := s.GetByID(ctx, procedureID)
	if err != nil {
		return nil, err
	}

	// Determine root ID
	rootID := procedureID
	if proc.ParentID != nil {
		rootID = *proc.ParentID
	}

	// Find version 0 in the chain
	var draft TestProcedure
	err = s.db.WithContext(ctx).
		Where("(id = ? OR parent_id = ?) AND version = ?", rootID, rootID, 0).
		First(&draft).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDraftNotFound
		}
		s.logger.Error(ctx, "failed to get draft", map[string]interface{}{
			"error":        err.Error(),
			"procedure_id": procedureID.String(),
		})
		return nil, err
	}

	return &draft, nil
}

// GetLatestCommitted retrieves the latest committed version (version >= 1, is_latest=true).
func (s *MySQLStore) GetLatestCommitted(ctx context.Context, procedureID uuid.UUID) (*TestProcedure, error) {
	// First get the procedure to determine root ID
	proc, err := s.GetByID(ctx, procedureID)
	if err != nil {
		return nil, err
	}

	// Determine root ID
	rootID := procedureID
	if proc.ParentID != nil {
		rootID = *proc.ParentID
	}

	// Find latest committed version (version >= 1 and is_latest = true)
	var committed TestProcedure
	err = s.db.WithContext(ctx).
		Where("(id = ? OR parent_id = ?) AND version >= ? AND is_latest = ?", rootID, rootID, 1, true).
		First(&committed).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoCommittedVersion
		}
		s.logger.Error(ctx, "failed to get latest committed", map[string]interface{}{
			"error":        err.Error(),
			"procedure_id": procedureID.String(),
		})
		return nil, err
	}

	return &committed, nil
}

// CreateWithDraft creates both a committed version (v1) and a draft (v0).
func (s *MySQLStore) CreateWithDraft(ctx context.Context, tp *TestProcedure) (*TestProcedure, error) {
	if err := tp.Validate(); err != nil {
		return nil, err
	}

	var v1 *TestProcedure

	// Execute in transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create v1 (committed version)
		v1 = &TestProcedure{
			ProjectID:   tp.ProjectID,
			Name:        tp.Name,
			Description: tp.Description,
			Steps:       tp.Steps,
			CreatedBy:   tp.CreatedBy,
			Version:     1,
			IsLatest:    true,
			ParentID:    nil,
		}

		if err := tx.WithContext(ctx).Create(v1).Error; err != nil {
			return fmt.Errorf("failed to create committed version: %w", err)
		}

		// Clone to v0 (draft version)
		v0 := &TestProcedure{
			ProjectID:   v1.ProjectID,
			Name:        v1.Name,
			Description: v1.Description,
			Steps:       v1.Steps,
			CreatedBy:   v1.CreatedBy,
			Version:     0,
			IsLatest:    false,
			ParentID:    &v1.ID,
		}

		if err := tx.WithContext(ctx).Create(v0).Error; err != nil {
			return fmt.Errorf("failed to create draft version: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "failed to create procedure with draft", map[string]interface{}{
			"error":      err.Error(),
			"name":       tp.Name,
			"project_id": tp.ProjectID.String(),
		})
		return nil, err
	}

	s.logger.Info(ctx, "test procedure created with draft", map[string]interface{}{
		"test_procedure_id": v1.ID.String(),
		"name":              v1.Name,
		"project_id":        v1.ProjectID.String(),
	})

	return v1, nil
}

// UpdateDraft updates only the draft version (v0) with the given setters.
func (s *MySQLStore) UpdateDraft(ctx context.Context, procedureID uuid.UUID, setters ...UpdateSetter) error {
	var draftID uuid.UUID

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		draft, err := s.getDraftWithTx(ctx, tx, procedureID)
		if err != nil {
			return err
		}

		for _, setter := range setters {
			if err := setter(draft); err != nil {
				return err
			}
		}

		if err := tx.WithContext(ctx).Save(draft).Error; err != nil {
			return err
		}

		draftID = draft.ID
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "failed to update draft", map[string]interface{}{
			"error":        err.Error(),
			"procedure_id": procedureID.String(),
		})
		return err
	}

	s.logger.Info(ctx, "draft updated", map[string]interface{}{
		"procedure_id": procedureID.String(),
		"draft_id":     draftID.String(),
	})

	return nil
}

// ResetDraft resets the draft (v0) to match the latest committed version.
func (s *MySQLStore) ResetDraft(ctx context.Context, procedureID uuid.UUID) error {
	var draftID uuid.UUID

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		committed, err := s.getLatestCommittedWithTx(ctx, tx, procedureID)
		if err != nil {
			return err
		}

		draft, err := s.getDraftWithTx(ctx, tx, procedureID)
		if err != nil {
			return err
		}

		draft.Name = committed.Name
		draft.Description = committed.Description
		draft.Steps = committed.Steps

		if err := tx.WithContext(ctx).Save(draft).Error; err != nil {
			return err
		}

		draftID = draft.ID
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "failed to reset draft", map[string]interface{}{
			"error":        err.Error(),
			"procedure_id": procedureID.String(),
		})
		return err
	}

	s.logger.Info(ctx, "draft reset to committed version", map[string]interface{}{
		"procedure_id": procedureID.String(),
		"draft_id":     draftID.String(),
	})

	return nil
}

// CommitDraft creates a new committed version from the draft, incrementing version number.
func (s *MySQLStore) CommitDraft(ctx context.Context, procedureID uuid.UUID) (*TestProcedure, error) {
	var newVersion *TestProcedure

	// Execute in transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get draft
		draft, err := s.getDraftWithTx(ctx, tx, procedureID)
		if err != nil {
			return err
		}

		// Determine root ID
		proc, err := s.getByIDWithTx(ctx, tx, procedureID)
		if err != nil {
			return err
		}

		rootID := procedureID
		if proc.ParentID != nil {
			rootID = *proc.ParentID
		}

		// Mark all versions in chain as is_latest=false
		if err := tx.WithContext(ctx).
			Model(&TestProcedure{}).
			Where("(id = ? OR parent_id = ?) AND version >= ?", rootID, rootID, 1).
			Update("is_latest", false).Error; err != nil {
			return fmt.Errorf("failed to update is_latest flags: %w", err)
		}

		// Find max version number in chain
		var maxVersion uint
		err = tx.WithContext(ctx).
			Model(&TestProcedure{}).
			Where("(id = ? OR parent_id = ?) AND version >= ?", rootID, rootID, 1).
			Select("COALESCE(MAX(version), 0)").
			Scan(&maxVersion).Error
		if err != nil {
			return fmt.Errorf("failed to get max version: %w", err)
		}

		// Create new committed version from draft
		newVersion = &TestProcedure{
			ProjectID:   draft.ProjectID,
			Name:        draft.Name,
			Description: draft.Description,
			Steps:       draft.Steps,
			CreatedBy:   draft.CreatedBy,
			Version:     maxVersion + 1,
			IsLatest:    true,
			ParentID:    &rootID,
		}

		if err := tx.WithContext(ctx).Create(newVersion).Error; err != nil {
			return fmt.Errorf("failed to create committed version: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "failed to commit draft", map[string]interface{}{
			"error":        err.Error(),
			"procedure_id": procedureID.String(),
		})
		return nil, err
	}

	s.logger.Info(ctx, "draft committed as new version", map[string]interface{}{
		"procedure_id":   procedureID.String(),
		"new_version_id": newVersion.ID.String(),
		"version":        newVersion.Version,
	})

	return newVersion, nil
}

// getDraftWithTx is a helper to get draft within a transaction.
func (s *MySQLStore) getDraftWithTx(ctx context.Context, tx *gorm.DB, procedureID uuid.UUID) (*TestProcedure, error) {
	// First get the procedure to determine root ID
	proc, err := s.getByIDWithTx(ctx, tx, procedureID)
	if err != nil {
		return nil, err
	}

	// Determine root ID
	rootID := procedureID
	if proc.ParentID != nil {
		rootID = *proc.ParentID
	}

	// Find version 0 in the chain
	var draft TestProcedure
	err = tx.WithContext(ctx).
		Where("(id = ? OR parent_id = ?) AND version = ?", rootID, rootID, 0).
		First(&draft).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDraftNotFound
		}
		return nil, err
	}

	return &draft, nil
}

// getLatestCommittedWithTx is a helper to get the latest committed version within a transaction.
func (s *MySQLStore) getLatestCommittedWithTx(ctx context.Context, tx *gorm.DB, procedureID uuid.UUID) (*TestProcedure, error) {
	proc, err := s.getByIDWithTx(ctx, tx, procedureID)
	if err != nil {
		return nil, err
	}

	rootID := procedureID
	if proc.ParentID != nil {
		rootID = *proc.ParentID
	}

	var committed TestProcedure
	err = tx.WithContext(ctx).
		Where("(id = ? OR parent_id = ?) AND version >= ? AND is_latest = ?", rootID, rootID, 1, true).
		First(&committed).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoCommittedVersion
		}
		return nil, err
	}

	return &committed, nil
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
