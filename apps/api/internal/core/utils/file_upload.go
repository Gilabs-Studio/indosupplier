package utils

import (
	"context"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gilabs/indosupplier/api/internal/core/storage"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
)

var (
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileTooLarge    = errors.New("file too large")
	ErrInvalidImage    = errors.New("invalid image file")
	ErrFileProcessing  = errors.New("error processing file")
)

// AllowedImageTypes defines allowed MIME types for image uploads
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// AllowedDocumentTypes defines allowed MIME types for document uploads
var AllowedDocumentTypes = map[string]bool{
	"application/pdf":    true, // PDF
	"application/msword": true, // DOC
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // DOCX
	"application/vnd.ms-excel": true, // XLS
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // XLSX
}

// FileUploadConfig holds configuration for file uploads
type FileUploadConfig struct {
	MaxSize      int64  // Maximum file size in bytes
	OriginalName string // Optional custom name for file (e.g., "Employee Name signature")
	Folder       string // R2 folder prefix (e.g., "avatars", "products", "visits") — if empty, uses root
}

// UploadedFile represents an uploaded file
type UploadedFile struct {
	Filename     string // Generated filename (UUID-based)
	OriginalName string // Original filename from upload
	Path         string // Full path to saved file
	URL          string // Public URL to access file
	Size         int64  // File size in bytes
	MimeType     string // MIME type
}

// ValidateImageFile validates file type, size, and content
func ValidateImageFile(file multipart.File, header *multipart.FileHeader, maxSize int64) error {
	// 1. Check file size
	if header.Size > maxSize {
		return ErrFileTooLarge
	}

	// 2. Read first 512 bytes for magic number detection
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return ErrFileProcessing
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return ErrFileProcessing
	}

	// 3. Detect file type using magic bytes
	kind, err := filetype.Match(buf[:n])
	if err != nil {
		return ErrInvalidFileType
	}

	// 4. Validate MIME type
	if !AllowedImageTypes[kind.MIME.Value] {
		return ErrInvalidFileType
	}

	// 5. Validate file extension matches MIME type
	// Accept both .jpg and .jpeg as valid JPEG extensions since they are equivalent
	ext := strings.ToLower(filepath.Ext(header.Filename))
	expectedExt := "." + kind.Extension
	isJpegAlias := (ext == ".jpg" && expectedExt == ".jpeg") || (ext == ".jpeg" && expectedExt == ".jpg")
	if ext != expectedExt && !isJpegAlias {
		return ErrInvalidFileType
	}

	return nil
}

// SaveUploadedFile validates, converts to WebP, and uploads an image to Cloudflare R2.
func SaveUploadedFile(file multipart.File, header *multipart.FileHeader, config FileUploadConfig) (*UploadedFile, error) {
	// 1. Validate file
	if err := ValidateImageFile(file, header, config.MaxSize); err != nil {
		return nil, err
	}

	// 2. Convert to WebP (or passthrough if CGO unavailable)
	webpData, err := ConvertToWebP(file)
	if err != nil {
		return nil, err
	}

	// 3. Generate secure UUID filename
	ext := ".webp"
	mimeType := "image/webp"
	if !WebPEnabled {
		ext = strings.ToLower(filepath.Ext(header.Filename))
		mimeType = header.Header.Get("Content-Type")
	}
	basename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// 4. Build full R2 key with folder prefix (structured: avatars/uuid.webp, products/productId/uuid.webp, etc)
	var key string
	if config.Folder != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(config.Folder, "/"), basename)
	} else {
		key = basename
	}

	// 5. Upload to R2
	url, err := storage.Upload(context.Background(), key, webpData, mimeType)
	if err != nil {
		return nil, ErrFileProcessing
	}

	return &UploadedFile{
		Filename:     basename,
		OriginalName: header.Filename,
		Path:         key, // Full R2 object key (includes folder) — used for deletion
		URL:          url,
		Size:         int64(len(webpData)),
		MimeType:     mimeType,
	}, nil
}

// ValidateDocumentFile validates document file type and size
func ValidateDocumentFile(file multipart.File, header *multipart.FileHeader, maxSize int64) error {
	// 1. Check file size
	if header.Size > maxSize {
		return ErrFileTooLarge
	}

	// 2. Read first 512 bytes for magic number detection
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return ErrFileProcessing
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return ErrFileProcessing
	}

	// 3. Detect file type using magic bytes
	kind, err := filetype.Match(buf[:n])
	if err != nil {
		return ErrInvalidFileType
	}

	// 4. Validate MIME type
	if !AllowedDocumentTypes[kind.MIME.Value] {
		return ErrInvalidFileType
	}

	return nil
}

// SaveDocumentFile validates and uploads a document file (PDF, DOCX, XLS) to Cloudflare R2.
func SaveDocumentFile(file multipart.File, header *multipart.FileHeader, config FileUploadConfig) (*UploadedFile, error) {
	// 1. Validate file
	if err := ValidateDocumentFile(file, header, config.MaxSize); err != nil {
		return nil, err
	}

	// 2. Detect file type for extension
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, ErrFileProcessing
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, ErrFileProcessing
	}

	kind, err := filetype.Match(buf[:n])
	if err != nil {
		return nil, ErrInvalidFileType
	}

	// 3. Generate secure filename: UUID + sanitized original name for better UX
	base := strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
	base = sanitizeFilenameBase(base)
	if base == "" {
		base = "document"
	}
	basename := fmt.Sprintf("%s_%s.%s", uuid.New().String(), base, kind.Extension)

	// 4. Read full file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, ErrFileProcessing
	}

	// 5. Build full R2 key with folder prefix (structured path)
	var key string
	if config.Folder != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(config.Folder, "/"), basename)
	} else {
		key = basename
	}

	// 6. Upload to R2
	url, err := storage.Upload(context.Background(), key, fileContent, kind.MIME.Value)
	if err != nil {
		return nil, ErrFileProcessing
	}

	return &UploadedFile{
		Filename:     basename,
		OriginalName: header.Filename,
		Path:         key, // Full R2 object key (includes folder) — used for deletion
		URL:          url,
		Size:         int64(len(fileContent)),
		MimeType:     kind.MIME.Value,
	}, nil
}

func sanitizeFilenameBase(input string) string {
	s := strings.TrimSpace(input)
	if s == "" {
		return ""
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		isAllowed := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_'
		if isAllowed {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteRune('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}
