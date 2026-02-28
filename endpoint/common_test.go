package endpoint

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/hairizuanbinnoorazman/ui-automation/testutil"
	"gorm.io/gorm"
)

// setupTestStore creates a test database and endpoint store for testing.
func setupTestStore(t *testing.T) (*gorm.DB, Store) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &Endpoint{})

	log := logger.NewTestLogger()
	store := NewMySQLStore(db, log)

	return db, store
}

// createTestEndpoint creates an endpoint with default values.
func createTestEndpoint(name, url string, createdBy uuid.UUID, creds Credentials) *Endpoint {
	return &Endpoint{
		Name:        name,
		URL:         url,
		CreatedBy:   createdBy,
		Credentials: creds,
	}
}
