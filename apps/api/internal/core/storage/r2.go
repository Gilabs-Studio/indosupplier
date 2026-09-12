package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
)

// R2StorageProvider implements StorageProvider for Cloudflare R2 (S3-compatible API).
type R2StorageProvider struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewR2StorageProvider initializes a new Cloudflare R2 client.
func NewR2StorageProvider(accountID, accessKeyID, secretKey, bucketName, pubURL string) (*R2StorageProvider, error) {
	if accountID == "" || accessKeyID == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("storage: R2 config incomplete — R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_BUCKET_NAME are required")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	c := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretKey, ""),
		UsePathStyle: true,
	})

	return &R2StorageProvider{
		client:    c,
		bucket:    bucketName,
		publicURL: strings.TrimSuffix(pubURL, "/"),
	}, nil
}

// Upload stores data under the given key in Cloudflare R2.
// Falls back to local filesystem if R2 client fails.
func (p *R2StorageProvider) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	var uploadErr error
	if p.client != nil {
		_, uploadErr = p.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(p.bucket),
			Key:           aws.String(key),
			Body:          bytes.NewReader(data),
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(int64(len(data))),
		})
		if uploadErr == nil {
			return p.URL(key), nil
		}
		log.Printf("[WARN] R2 upload failed for key %q: %v. Falling back to local filesystem storage.", key, uploadErr)
	} else {
		log.Printf("[WARN] R2 storage client is not initialized. Falling back to local filesystem storage.")
	}

	// Local filesystem fallback
	localPath := filepath.Clean(key)
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("storage fallback: failed to create local directories: %w", err)
	}

	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return "", fmt.Errorf("storage fallback: failed to write local file: %w", err)
	}

	port := "8088"
	if config.AppConfig != nil && config.AppConfig.Server.Port != "" {
		port = config.AppConfig.Server.Port
	}

	localURL := fmt.Sprintf("http://localhost:%s/%s", port, key)
	log.Printf("[INFO] Local file fallback saved successfully. URL: %s", localURL)
	return localURL, nil
}

// Delete removes the object identified by key from the R2 bucket.
func (p *R2StorageProvider) Delete(ctx context.Context, key string) error {
	if p.client != nil {
		_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(p.bucket),
			Key:    aws.String(key),
		})
		if err == nil {
			return nil
		}
		log.Printf("[WARN] R2 delete failed for key %q: %v. Attempting local deletion fallback.", key, err)
	}

	// Local filesystem fallback
	localPath := filepath.Clean(key)
	if _, err := os.Stat(localPath); err == nil {
		if err := os.Remove(localPath); err != nil {
			return fmt.Errorf("storage fallback: failed to delete local file: %w", err)
		}
		log.Printf("[INFO] Local file deleted successfully: %s", localPath)
	}
	return nil
}

// DeleteByPrefix removes all objects under a prefix from the R2 bucket.
func (p *R2StorageProvider) DeleteByPrefix(ctx context.Context, prefix string) error {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		return nil
	}

	if p.client == nil {
		log.Printf("[WARN] R2 storage client is not initialized. Attempting local directory deletion fallback.")
		localPath := filepath.Clean(trimmed)
		if _, err := os.Stat(localPath); err == nil {
			if err := os.RemoveAll(localPath); err != nil {
				return fmt.Errorf("storage fallback: failed to delete local directory: %w", err)
			}
			log.Printf("[INFO] Local directory deleted successfully: %s", localPath)
		}
		return nil
	}

	pager := s3.NewListObjectsV2Paginator(p.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucket),
		Prefix: aws.String(strings.TrimPrefix(trimmed, "/")),
	})

	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("storage: R2 list by prefix failed for %q: %w", trimmed, err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		objects := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			if obj.Key == nil || strings.TrimSpace(*obj.Key) == "" {
				continue
			}
			objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
		}

		if len(objects) == 0 {
			continue
		}

		_, err = p.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(p.bucket),
			Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return fmt.Errorf("storage: R2 delete by prefix failed for %q: %w", trimmed, err)
		}
	}

	return nil
}

// URL returns the public URL for a given object key in R2.
func (p *R2StorageProvider) URL(key string) string {
	cleanKey := strings.TrimPrefix(key, "/")
	if p.publicURL == "" {
		return cleanKey
	}
	return p.publicURL + "/" + cleanKey
}

// KeyFromURL extracts the R2 object key from a public URL.
func (p *R2StorageProvider) KeyFromURL(fullURL string) string {
	if p.publicURL == "" || fullURL == "" {
		return ""
	}
	prefix := p.publicURL + "/"
	if strings.HasPrefix(fullURL, prefix) {
		return strings.TrimPrefix(fullURL, prefix)
	}
	return ""
}

// Init initializes the Cloudflare R2 client. Provided for backward compatibility.
func Init(accountID, accessKeyID, secretKey, bucketName, pubURL string) error {
	r2, err := NewR2StorageProvider(accountID, accessKeyID, secretKey, bucketName, pubURL)
	if err != nil {
		return err
	}
	defaultProvider = r2
	return nil
}
