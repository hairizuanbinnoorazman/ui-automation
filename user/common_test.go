package user

import (
	"testing"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/testutil"
	"gorm.io/gorm"
)

// setupTestStore creates a test database and user store for testing.
func setupTestStore(t *testing.T) (*gorm.DB, Store) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &User{})

	log := logger.NewTestLogger()
	store := NewMySQLStore(db, log)

	return db, store
}

// createTestUser creates a test user with default values.
func createTestUser(email, username, password string) *User {
	user := &User{
		Email:    email,
		Username: username,
		IsActive: true,
	}
	user.SetPassword(password)
	return user
}
