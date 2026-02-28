package project

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/hairizuanbinnoorazman/ui-automation/testutil"
	"gorm.io/gorm"
)

// setupTestStore creates a test database and project store for testing.
func setupTestStore(t *testing.T) (*gorm.DB, Store) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &Project{})

	log := logger.NewTestLogger()
	store := NewMySQLStore(db, log)

	return db, store
}

// createTestProject creates a test project with default values.
func createTestProject(name, description string, ownerID uuid.UUID) *Project {
	return &Project{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		IsActive:    true,
	}
}
