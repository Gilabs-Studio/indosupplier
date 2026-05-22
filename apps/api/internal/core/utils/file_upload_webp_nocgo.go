//go:build !cgo

package utils

import (
	"io"
	"log"
	"mime/multipart"
)

func init() {
	log.Println("[WARN] WebP conversion disabled (CGO not available). Images will be saved in original format.")
}

// ConvertToWebP is a no-op fallback when CGO is unavailable.
// It returns the original file bytes without conversion.
func ConvertToWebP(file multipart.File) ([]byte, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, ErrFileProcessing
	}

	if _, err := file.Seek(0, 0); err != nil {
		return nil, ErrFileProcessing
	}

	return data, nil
}

// WebPEnabled reports whether WebP conversion is available.
const WebPEnabled = false
