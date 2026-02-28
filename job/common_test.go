package job

import (
	"testing"

	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/hairizuanbinnoorazman/ui-automation/testutil"
	"gorm.io/gorm"
)

// setupTestStore creates a test database and job store for testing.
func setupTestStore(t *testing.T) (*gorm.DB, Store) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &Job{})

	log := logger.NewTestLogger()
	store := NewMySQLStore(db, log)

	return db, store
}
