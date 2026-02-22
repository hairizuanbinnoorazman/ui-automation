package testrun

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStepNoteStore implements StepNoteStore using GORM and MySQL.
type MySQLStepNoteStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStepNoteStore creates a new MySQL-backed step note store.
func NewMySQLStepNoteStore(db *gorm.DB, log logger.Logger) *MySQLStepNoteStore {
	return &MySQLStepNoteStore{
		db:     db,
		logger: log,
	}
}

// Upsert creates or updates a step note for a given (test_run_id, step_index).
func (s *MySQLStepNoteStore) Upsert(ctx context.Context, note *StepNote) error {
	existing, err := s.GetByRunAndStep(ctx, note.TestRunID, note.StepIndex)
	if err != nil && !errors.Is(err, ErrStepNoteNotFound) {
		return err
	}

	if existing != nil {
		existing.Notes = note.Notes
		if err := s.db.WithContext(ctx).Save(existing).Error; err != nil {
			s.logger.Error(ctx, "failed to update step note", map[string]interface{}{
				"error":       err.Error(),
				"test_run_id": note.TestRunID.String(),
				"step_index":  note.StepIndex,
			})
			return err
		}
		*note = *existing
		return nil
	}

	if err := s.db.WithContext(ctx).Create(note).Error; err != nil {
		s.logger.Error(ctx, "failed to create step note", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": note.TestRunID.String(),
			"step_index":  note.StepIndex,
		})
		return err
	}

	return nil
}

// ListByTestRun retrieves all step notes for a specific test run, ordered by step_index.
func (s *MySQLStepNoteStore) ListByTestRun(ctx context.Context, testRunID uuid.UUID) ([]*StepNote, error) {
	var notes []*StepNote
	err := s.db.WithContext(ctx).
		Where("test_run_id = ?", testRunID).
		Order("step_index ASC").
		Find(&notes).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list step notes by test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": testRunID.String(),
		})
		return nil, err
	}

	return notes, nil
}

// GetByRunAndStep retrieves a step note for a specific run and step index.
func (s *MySQLStepNoteStore) GetByRunAndStep(ctx context.Context, testRunID uuid.UUID, stepIndex int) (*StepNote, error) {
	var note StepNote
	err := s.db.WithContext(ctx).
		Where("test_run_id = ? AND step_index = ?", testRunID, stepIndex).
		First(&note).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStepNoteNotFound
		}
		s.logger.Error(ctx, "failed to get step note", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": testRunID.String(),
			"step_index":  stepIndex,
		})
		return nil, err
	}

	return &note, nil
}
