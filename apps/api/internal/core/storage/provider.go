package storage

import (
	"context"
)

// StorageProvider defines the unified interface for file storage operations.
// Supports both Local filesystem (development) and Cloudflare R2 / S3 (production).
type StorageProvider interface {
	// Upload stores data under the given key and returns the accessible public URL.
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)

	// Delete removes a single object identified by key.
	Delete(ctx context.Context, key string) error

	// DeleteByPrefix removes all objects matching a given prefix.
	DeleteByPrefix(ctx context.Context, prefix string) error

	// URL returns the public URL for an object key.
	URL(key string) string

	// KeyFromURL extracts the internal object key from a full public URL.
	KeyFromURL(fullURL string) string
}
