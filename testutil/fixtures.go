package testutil

import (
	"testing"

	"gorm.io/gorm"
)

// CreateFixture creates a fixture in the database.
func CreateFixture(t *testing.T, db *gorm.DB, model interface{}) {
	if err := db.Create(model).Error; err != nil {
		t.Fatalf("failed to create fixture: %v", err)
	}
}

// CreateFixtures creates multiple fixtures in the database.
func CreateFixtures(t *testing.T, db *gorm.DB, models ...interface{}) {
	for _, model := range models {
		CreateFixture(t, db, model)
	}
}
