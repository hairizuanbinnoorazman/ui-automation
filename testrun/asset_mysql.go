package testrun

import (
	"context"
	"errors"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLAssetStore implements the AssetStore interface using GORM and MySQL.
type MySQLAssetStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLAssetStore creates a new MySQL-backed asset store.
func NewMySQLAssetStore(db *gorm.DB, log logger.Logger) *MySQLAssetStore {
	return &MySQLAssetStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new asset in the database.
func (s *MySQLAssetStore) Create(ctx context.Context, asset *TestRunAsset) error {
	if err := asset.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(asset).Error; err != nil {
		s.logger.Error(ctx, "failed to create asset", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": asset.TestRunID,
			"file_name":   asset.FileName,
		})
		return err
	}

	s.logger.Info(ctx, "asset created", map[string]interface{}{
		"asset_id":    asset.ID,
		"test_run_id": asset.TestRunID,
		"file_name":   asset.FileName,
	})

	return nil
}

// GetByID retrieves an asset by its ID.
func (s *MySQLAssetStore) GetByID(ctx context.Context, id uint) (*TestRunAsset, error) {
	var asset TestRunAsset
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&asset).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssetNotFound
		}
		s.logger.Error(ctx, "failed to get asset by ID", map[string]interface{}{
			"error":    err.Error(),
			"asset_id": id,
		})
		return nil, err
	}

	return &asset, nil
}

// ListByTestRun retrieves all assets for a specific test run.
func (s *MySQLAssetStore) ListByTestRun(ctx context.Context, testRunID uint) ([]*TestRunAsset, error) {
	var assets []*TestRunAsset
	err := s.db.WithContext(ctx).
		Where("test_run_id = ?", testRunID).
		Order("uploaded_at ASC").
		Find(&assets).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list assets by test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": testRunID,
		})
		return nil, err
	}

	return assets, nil
}

// Delete deletes an asset by ID.
func (s *MySQLAssetStore) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&TestRunAsset{})

	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete asset", map[string]interface{}{
			"error":    result.Error.Error(),
			"asset_id": id,
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrAssetNotFound
	}

	s.logger.Info(ctx, "asset deleted", map[string]interface{}{
		"asset_id": id,
	})

	return nil
}
