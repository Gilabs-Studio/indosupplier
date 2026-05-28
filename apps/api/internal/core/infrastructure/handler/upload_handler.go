package handler

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func sanitizePathSegment(segment string) string {
	trimmed := strings.TrimSpace(segment)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range strings.ToLower(trimmed) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}

	return strings.Trim(b.String(), "-_")
}

func sanitizeFolderPath(rawFolder string, fallback string) string {
	trimmed := strings.Trim(rawFolder, " /")
	if trimmed == "" {
		return fallback
	}

	parts := strings.Split(trimmed, "/")
	sanitized := make([]string, 0, len(parts))
	for _, part := range parts {
		segment := sanitizePathSegment(part)
		if segment == "" {
			continue
		}
		if segment == "uploads" {
			continue
		}
		sanitized = append(sanitized, segment)
	}

	if len(sanitized) == 0 {
		return fallback
	}

	return strings.Join(sanitized, "/")
}

func resolveUploadNamespace(c *gin.Context) string {
	userID := sanitizePathSegment(c.GetString("user_id"))
	if userID != "" {
		return userID
	}

	if rawUserID := c.Request.Context().Value("user_id"); rawUserID != nil {
		if value, ok := rawUserID.(string); ok {
			userID = sanitizePathSegment(value)
			if userID != "" {
				return userID
			}
		}
	}

	return "public"
}

func buildScopedFolder(c *gin.Context, defaultModule string) string {
	namespace := resolveUploadNamespace(c)
	moduleFolder := sanitizeFolderPath(c.Query("folder"), defaultModule)
	return fmt.Sprintf("uploads/%s/%s", namespace, moduleFolder)
}

// UploadImage handles image upload requests
// Query params: ?folder=attendance,delivery-orders,etc (default: 'images')
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 1. Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		errors.ErrorResponse(c, "INVALID_REQUEST", map[string]interface{}{
			"field": "file",
			"error": "file is required",
		}, nil)
		return
	}
	defer file.Close()

	// 2. Prepare upload config with user namespace and module-based object key.
	folder := buildScopedFolder(c, "images")
	uploadConfig := utils.FileUploadConfig{
		MaxSize: config.AppConfig.Storage.MaxUploadSize,
		Folder:  folder,
	}

	// 3. Save and process file
	uploadedFile, err := utils.SaveUploadedFile(file, header, uploadConfig)
	if err != nil {
		switch err {
		case utils.ErrInvalidFileType:
			errors.ErrorResponse(c, "INVALID_FILE_TYPE", map[string]interface{}{
				"allowed_types": []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
			}, nil)
		case utils.ErrFileTooLarge:
			errors.ErrorResponse(c, "FILE_TOO_LARGE", map[string]interface{}{
				"max_size": config.AppConfig.Storage.MaxUploadSize,
			}, nil)
		case utils.ErrInvalidImage:
			errors.ErrorResponse(c, "INVALID_IMAGE", map[string]interface{}{
				"error": "corrupted or invalid image file",
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, "failed to process upload")
		}
		return
	}

	// 4. Return success response
	resp := map[string]interface{}{
		"filename":      uploadedFile.Filename,
		"original_name": uploadedFile.OriginalName,
		"url":           uploadedFile.URL,
		"size":          uploadedFile.Size,
		"mime_type":     uploadedFile.MimeType,
	}

	response.SuccessResponseCreated(c, resp, nil)
}

// UploadDocument handles document upload requests (PDF, DOCX, XLS, etc.)
// Query params: ?folder=attendance,delivery-orders,etc (default: 'documents')
func (h *UploadHandler) UploadDocument(c *gin.Context) {
	// 1. Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		errors.ErrorResponse(c, "INVALID_REQUEST", map[string]interface{}{
			"field": "file",
			"error": "file is required",
		}, nil)
		return
	}
	defer file.Close()

	// 2. Prepare upload config with user namespace and module-based object key.
	folder := buildScopedFolder(c, "documents")
	uploadConfig := utils.FileUploadConfig{
		MaxSize: config.AppConfig.Storage.MaxUploadSize,
		Folder:  folder,
	}

	// 3. Save document file
	uploadedFile, err := utils.SaveDocumentFile(file, header, uploadConfig)
	if err != nil {
		switch err {
		case utils.ErrInvalidFileType:
			errors.ErrorResponse(c, "INVALID_FILE_TYPE", map[string]interface{}{
				"allowed_types": []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
			}, nil)
		case utils.ErrFileTooLarge:
			errors.ErrorResponse(c, "FILE_TOO_LARGE", map[string]interface{}{
				"max_size": config.AppConfig.Storage.MaxUploadSize,
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, "failed to process upload")
		}
		return
	}

	// 4. Return success response
	resp := map[string]interface{}{
		"filename":      uploadedFile.Filename,
		"original_name": uploadedFile.OriginalName,
		"url":           uploadedFile.URL,
		"size":          uploadedFile.Size,
		"mime_type":     uploadedFile.MimeType,
	}

	response.SuccessResponseCreated(c, resp, nil)
}
