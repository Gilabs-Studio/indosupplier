package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorageProvider implements StorageProvider using the local filesystem.
// Dedicated for local development per API File Upload Standards.
type LocalStorageProvider struct {
	baseDir string
	baseURL string
}

// NewLocalStorageProvider creates a new instance of LocalStorageProvider.
func NewLocalStorageProvider(baseDir, baseURL string) *LocalStorageProvider {
	trimmedDir := strings.TrimSpace(baseDir)
	if trimmedDir == "" {
		trimmedDir = "./uploads"
	}
	trimmedURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedURL == "" {
		trimmedURL = "http://localhost:8088/uploads"
	}

	// Ensure base directory exists
	if err := os.MkdirAll(trimmedDir, 0755); err != nil {
		log.Printf("[WARN] Failed to create storage base directory %q: %v", trimmedDir, err)
	}

	return &LocalStorageProvider{
		baseDir: trimmedDir,
		baseURL: trimmedURL,
	}
}

// normalizeKey strips leading slashes and any redundant "uploads/" prefix
// so the relative key is consistent within the storage base directory.
func (l *LocalStorageProvider) normalizeKey(key string) string {
	clean := filepath.ToSlash(filepath.Clean(key))
	clean = strings.TrimPrefix(clean, "/")
	clean = strings.TrimPrefix(clean, "uploads/")
	return strings.TrimPrefix(clean, "/")
}

// resolvePath returns the secure absolute filesystem path for a given key,
// verifying that path traversal cannot escape the configured base directory.
func (l *LocalStorageProvider) resolvePath(key string) (string, error) {
	relKey := l.normalizeKey(key)
	if relKey == "" || relKey == "." {
		return "", fmt.Errorf("storage local: invalid empty key")
	}

	absBase, err := filepath.Abs(l.baseDir)
	if err != nil {
		return "", fmt.Errorf("storage local: failed to resolve base dir: %w", err)
	}

	targetPath := filepath.Join(absBase, filepath.FromSlash(relKey))
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("storage local: failed to resolve target path: %w", err)
	}

	if !strings.HasPrefix(absTarget, absBase) {
		return "", fmt.Errorf("storage local: directory traversal detected")
	}

	return absTarget, nil
}

// Upload stores file data into the local filesystem under baseDir.
func (l *LocalStorageProvider) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	targetPath, err := l.resolvePath(key)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("storage local: failed to create directories: %w", err)
	}

	// Write file with non-executable permissions (0644)
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return "", fmt.Errorf("storage local: failed to write file: %w", err)
	}

	relKey := l.normalizeKey(key)
	url := fmt.Sprintf("%s/%s", l.baseURL, relKey)
	log.Printf("[INFO] Local file saved successfully: %s -> %s", targetPath, url)
	return url, nil
}

// Delete removes a file from the local filesystem.
func (l *LocalStorageProvider) Delete(ctx context.Context, key string) error {
	targetPath, err := l.resolvePath(key)
	if err != nil {
		return err
	}

	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("storage local: failed to delete file: %w", err)
		}
		log.Printf("[INFO] Local file deleted successfully: %s", targetPath)
	}
	return nil
}

// DeleteByPrefix removes an entire directory or prefix from the local filesystem.
func (l *LocalStorageProvider) DeleteByPrefix(ctx context.Context, prefix string) error {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		return nil
	}

	targetPath, err := l.resolvePath(trimmed)
	if err != nil {
		return err
	}

	if _, err := os.Stat(targetPath); err == nil {
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("storage local: failed to delete directory: %w", err)
		}
		log.Printf("[INFO] Local directory deleted successfully: %s", targetPath)
	}
	return nil
}

// URL returns the public URL for a given object key.
func (l *LocalStorageProvider) URL(key string) string {
	relKey := l.normalizeKey(key)
	return fmt.Sprintf("%s/%s", l.baseURL, relKey)
}

// KeyFromURL extracts the relative object key from a full public URL.
func (l *LocalStorageProvider) KeyFromURL(fullURL string) string {
	if fullURL == "" {
		return ""
	}
	prefix := l.baseURL + "/"
	if strings.HasPrefix(fullURL, prefix) {
		return strings.TrimPrefix(fullURL, prefix)
	}
	return ""
}
