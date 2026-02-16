package testrun

import (
	"context"
	"errors"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStore implements the Store interface using GORM and MySQL.
type MySQLStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStore creates a new MySQL-backed test run store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new test run in the database.
func (s *MySQLStore) Create(ctx context.Context, testRun *TestRun) error {
	// Ensure default status is set before validation
	if testRun.Status == "" {
		testRun.Status = StatusPending
	}

	if err := testRun.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(testRun).Error; err != nil {
		s.logger.Error(ctx, "failed to create test run", map[string]interface{}{
			"error":               err.Error(),
			"test_procedure_id":   testRun.TestProcedureID,
			"executed_by":         testRun.ExecutedBy,
		})
		return err
	}

	s.logger.Info(ctx, "test run created", map[string]interface{}{
		"test_run_id":         testRun.ID,
		"test_procedure_id":   testRun.TestProcedureID,
	})

	return nil
}

// GetByID retrieves a test run by its ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uint) (*TestRun, error) {
	var testRun TestRun
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&testRun).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTestRunNotFound
		}
		s.logger.Error(ctx, "failed to get test run by ID", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		return nil, err
	}

	return &testRun, nil
}

// Update updates a test run with the given setters.
func (s *MySQLStore) Update(ctx context.Context, id uint, setters ...UpdateSetter) error {
	// First, fetch the test run
	testRun, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply all setters
	for _, setter := range setters {
		if err := setter(testRun); err != nil {
			return err
		}
	}

	// Save the updated test run
	if err := s.db.WithContext(ctx).Save(testRun).Error; err != nil {
		s.logger.Error(ctx, "failed to update test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		return err
	}

	s.logger.Info(ctx, "test run updated", map[string]interface{}{
		"test_run_id": id,
	})

	return nil
}

// ListByTestProcedure retrieves a paginated list of test runs for a specific test procedure.
func (s *MySQLStore) ListByTestProcedure(ctx context.Context, testProcedureID uint, limit, offset int) ([]*TestRun, error) {
	var testRuns []*TestRun
	err := s.db.WithContext(ctx).
		Where("test_procedure_id = ?", testProcedureID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&testRuns).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list test runs by test procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": testProcedureID,
			"limit":             limit,
			"offset":            offset,
		})
		return nil, err
	}

	return testRuns, nil
}

// Start marks a test run as started (sets started_at, changes status to running).
func (s *MySQLStore) Start(ctx context.Context, id uint) error {
	// Fetch the test run
	testRun, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Call the domain method
	if err := testRun.Start(); err != nil {
		return err
	}

	// Save the updated test run
	if err := s.db.WithContext(ctx).Save(testRun).Error; err != nil {
		s.logger.Error(ctx, "failed to start test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		return err
	}

	s.logger.Info(ctx, "test run started", map[string]interface{}{
		"test_run_id": id,
	})

	return nil
}

// Complete marks a test run as completed (sets completed_at, final status, optional notes).
func (s *MySQLStore) Complete(ctx context.Context, id uint, status Status, notes string) error {
	// Fetch the test run
	testRun, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Call the domain method
	if err := testRun.Complete(status, notes); err != nil {
		return err
	}

	// Save the updated test run
	if err := s.db.WithContext(ctx).Save(testRun).Error; err != nil {
		s.logger.Error(ctx, "failed to complete test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		return err
	}

	s.logger.Info(ctx, "test run completed", map[string]interface{}{
		"test_run_id": id,
		"status":      status,
	})

	return nil
}
