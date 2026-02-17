package testrun

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrAssetNotFound is returned when an asset is not found.
	ErrAssetNotFound = errors.New("asset not found")

	// ErrInvalidAssetType is returned when asset type is invalid.
	ErrInvalidAssetType = errors.New("invalid asset type")

	// ErrInvalidTestRunID is returned when test_run_id is not set.
	ErrInvalidTestRunID = errors.New("test_run_id is required")

	// ErrInvalidAssetPath is returned when asset_path is empty.
	ErrInvalidAssetPath = errors.New("asset_path is required")

	// ErrInvalidFileName is returned when file_name is empty.
	ErrInvalidFileName = errors.New("file_name is required")
)

// AssetType represents the type of asset.
type AssetType string

const (
	AssetTypeImage    AssetType = "image"
	AssetTypeVideo    AssetType = "video"
	AssetTypeBinary   AssetType = "binary"
	AssetTypeDocument AssetType = "document"
)

// IsValid checks if the asset type is valid.
func (at AssetType) IsValid() bool {
	switch at {
	case AssetTypeImage, AssetTypeVideo, AssetTypeBinary, AssetTypeDocument:
		return true
	default:
		return false
	}
}

// TestRunAsset represents an asset associated with a test run.
type TestRunAsset struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	TestRunID   uuid.UUID `json:"test_run_id" gorm:"type:char(36);not null;index:idx_test_run_id"`
	AssetType   AssetType `json:"asset_type" gorm:"type:varchar(20);not null;index:idx_asset_type"`
	AssetPath   string    `json:"asset_path" gorm:"type:varchar(512);not null"`
	FileName    string    `json:"file_name" gorm:"type:varchar(255);not null"`
	FileSize    int64     `json:"file_size" gorm:"not null"`
	MimeType    string    `json:"mime_type,omitempty" gorm:"type:varchar(128)"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// BeforeCreate hook to generate UUID before creating a new test run asset
func (a *TestRunAsset) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// Validate checks if the asset has valid required fields.
func (a *TestRunAsset) Validate() error {
	if a.TestRunID == uuid.Nil {
		return ErrInvalidTestRunID
	}
	if !a.AssetType.IsValid() {
		return ErrInvalidAssetType
	}
	if a.AssetPath == "" {
		return ErrInvalidAssetPath
	}
	if a.FileName == "" {
		return ErrInvalidFileName
	}
	return nil
}
