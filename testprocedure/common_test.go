package testprocedure

import (
	"testing"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/testutil"
	"gorm.io/gorm"
)

// setupTestStore creates a test database and test procedure store for testing.
func setupTestStore(t *testing.T) (*gorm.DB, Store) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestProcedure{})

	log := logger.NewTestLogger()
	store := NewMySQLStore(db, log)

	return db, store
}

// createTestProcedure creates a test procedure with default values.
func createTestProcedure(name, description string, projectID, createdBy uint, steps Steps) *TestProcedure {
	return &TestProcedure{
		Name:        name,
		Description: description,
		ProjectID:   projectID,
		CreatedBy:   createdBy,
		Steps:       steps,
	}
}
