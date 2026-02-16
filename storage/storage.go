package storage

import (
	"context"
	"io"
)

// BlobStorage defines the interface for storing and retrieving binary data.
type BlobStorage interface {
	// Upload stores data from the reader at the specified path.
	Upload(ctx context.Context, path string, reader io.Reader) error

	// Download retrieves data from the specified path.
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes the data at the specified path.
	Delete(ctx context.Context, path string) error

	// Exists checks if data exists at the specified path.
	Exists(ctx context.Context, path string) (bool, error)

	// GetURL returns a URL for accessing the data at the specified path.
	// For local storage, this returns a file:// URL or relative path.
	GetURL(ctx context.Context, path string) (string, error)
}
