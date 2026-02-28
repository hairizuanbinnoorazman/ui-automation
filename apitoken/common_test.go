package apitoken

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/hairizuanbinnoorazman/ui-automation/testutil"
	"gorm.io/gorm"
)

// setupTestStore creates a test database and API token store for testing.
func setupTestStore(t *testing.T) (*gorm.DB, Store) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &APIToken{})

	log := logger.NewTestLogger()
	store := NewMySQLStore(db, log)

	return db, store
}

// createTestToken creates an API token with default values for testing.
func createTestToken(name string, userID uuid.UUID, scope string, hash string) *APIToken {
	return &APIToken{
		Name:      name,
		UserID:    userID,
		Scope:     scope,
		TokenHash: hash,
		IsActive:  true,
	}
}
