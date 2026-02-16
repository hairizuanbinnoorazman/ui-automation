package testutil

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing.
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return db
}

// AutoMigrate runs GORM auto-migrations for the given models.
func AutoMigrate(t *testing.T, db *gorm.DB, models ...interface{}) {
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
}
