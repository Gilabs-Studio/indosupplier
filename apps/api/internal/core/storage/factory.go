package storage

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
)

var defaultProvider StorageProvider

// InitStorage initializes the active storage provider based on configuration.
// Supports driver="local" (for development) and driver="r2" (for production).
func InitStorage(cfg config.StorageConfig, serverPort string) error {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		driver = "r2"
	}

	switch driver {
	case "local":
		localDir := strings.TrimSpace(cfg.LocalUploadDir)
		if localDir == "" {
			localDir = "./uploads"
		}
		localURL := strings.TrimSpace(cfg.LocalBaseURL)
		if localURL == "" {
			port := serverPort
			if port == "" {
				port = "8088"
			}
			localURL = fmt.Sprintf("http://localhost:%s/uploads", port)
		}
		defaultProvider = NewLocalStorageProvider(localDir, localURL)
		log.Printf("[INFO] Storage driver initialized: LOCAL (dir=%s, url=%s)", localDir, localURL)
		return nil

	case "r2":
		r2, err := NewR2StorageProvider(
			cfg.R2AccountID,
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			cfg.R2BucketName,
			cfg.R2PublicURL,
		)
		if err != nil {
			log.Printf("[WARN] Failed to initialize R2 storage provider: %v. Falling back to local storage.", err)
			port := serverPort
			if port == "" {
				port = "8088"
			}
			localURL := fmt.Sprintf("http://localhost:%s/uploads", port)
			defaultProvider = NewLocalStorageProvider("./uploads", localURL)
			return nil
		}
		defaultProvider = r2
		log.Printf("[INFO] Storage driver initialized: CLOUDFLARE R2 (bucket=%s, url=%s)", cfg.R2BucketName, cfg.R2PublicURL)
		return nil

	default:
		return fmt.Errorf("storage: unsupported driver %q (must be 'local' or 'r2')", driver)
	}
}

// GetProvider returns the currently configured storage provider.
func GetProvider() StorageProvider {
	if defaultProvider == nil {
		defaultProvider = NewLocalStorageProvider("./uploads", "http://localhost:8088/uploads")
	}
	return defaultProvider
}

// Upload stores data under the given key using the active storage driver.
func Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	return GetProvider().Upload(ctx, key, data, contentType)
}

// Delete removes the object identified by key using the active storage driver.
func Delete(ctx context.Context, key string) error {
	return GetProvider().Delete(ctx, key)
}

// DeleteByPrefix removes all objects matching a given prefix using the active storage driver.
func DeleteByPrefix(ctx context.Context, prefix string) error {
	return GetProvider().DeleteByPrefix(ctx, prefix)
}

// URL returns the public URL for a given object key using the active storage driver.
func URL(key string) string {
	return GetProvider().URL(key)
}

// KeyFromURL extracts the internal object key from a full public URL using the active storage driver.
func KeyFromURL(fullURL string) string {
	return GetProvider().KeyFromURL(fullURL)
}
