package testrun

import (
	"testing"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/testutil"
	"gorm.io/gorm"
)

// setupTestStore creates a test database and test run store for testing.
func setupTestStore(t *testing.T) (*gorm.DB, Store, AssetStore) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestRun{}, &TestRunAsset{})

	log := logger.NewTestLogger()
	store := NewMySQLStore(db, log)
	assetStore := NewMySQLAssetStore(db, log)

	return db, store, assetStore
}

// createTestRun creates a test run with default values.
func createTestRun(testProcedureID, executedBy uint, status Status, notes string) *TestRun {
	return &TestRun{
		TestProcedureID: testProcedureID,
		ExecutedBy:      executedBy,
		Status:          status,
		Notes:           notes,
	}
}

// createTestAsset creates a test run asset with default values.
func createTestAsset(testRunID uint, assetType AssetType, path, fileName string, size int64) *TestRunAsset {
	return &TestRunAsset{
		TestRunID: testRunID,
		AssetType: assetType,
		AssetPath: path,
		FileName:  fileName,
		FileSize:  size,
		MimeType:  "application/octet-stream",
	}
}
