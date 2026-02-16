package testrun

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssetType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		assetType AssetType
		want      bool
	}{
		{"image is valid", AssetTypeImage, true},
		{"video is valid", AssetTypeVideo, true},
		{"binary is valid", AssetTypeBinary, true},
		{"document is valid", AssetTypeDocument, true},
		{"invalid type", AssetType("invalid"), false},
		{"empty type", AssetType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.assetType.IsValid())
		})
	}
}

func TestTestRunAsset_Validate(t *testing.T) {
	tests := []struct {
		name    string
		asset   TestRunAsset
		wantErr error
	}{
		{
			name: "valid asset",
			asset: TestRunAsset{
				TestRunID: 1,
				AssetType: AssetTypeImage,
				AssetPath: "path/to/file.png",
				FileName:  "file.png",
				FileSize:  1024,
			},
			wantErr: nil,
		},
		{
			name: "missing test_run_id",
			asset: TestRunAsset{
				AssetType: AssetTypeImage,
				AssetPath: "path/to/file.png",
				FileName:  "file.png",
				FileSize:  1024,
			},
			wantErr: ErrInvalidTestRunID,
		},
		{
			name: "invalid asset type",
			asset: TestRunAsset{
				TestRunID: 1,
				AssetType: AssetType("invalid"),
				AssetPath: "path/to/file.png",
				FileName:  "file.png",
				FileSize:  1024,
			},
			wantErr: ErrInvalidAssetType,
		},
		{
			name: "missing asset path",
			asset: TestRunAsset{
				TestRunID: 1,
				AssetType: AssetTypeImage,
				FileName:  "file.png",
				FileSize:  1024,
			},
			wantErr: ErrInvalidAssetPath,
		},
		{
			name: "missing file name",
			asset: TestRunAsset{
				TestRunID: 1,
				AssetType: AssetTypeImage,
				AssetPath: "path/to/file.png",
				FileSize:  1024,
			},
			wantErr: ErrInvalidFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
