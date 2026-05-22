package utils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gilabs/gims/api/internal/core/storage"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
)

// SaveSignatureFile validates and uploads a signature image to Cloudflare R2.
// Signatures are stored under the "signatures/" key prefix for logical grouping.
func SaveSignatureFile(file multipart.File, header *multipart.FileHeader, config FileUploadConfig) (*UploadedFile, error) {
// 1. Validate file type and size
if err := ValidateImageFile(file, header, config.MaxSize); err != nil {
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

// 3. Generate filename: use OriginalName if provided (e.g., "John Doe signature.png"),
// otherwise fall back to a UUID-based name.
var filename, originalFilename string
if config.OriginalName != "" {
filename = fmt.Sprintf("%s.%s", sanitizeFilenameBase(config.OriginalName), kind.Extension)
originalFilename = filename
} else {
filename = fmt.Sprintf("%s.%s", uuid.New().String(), kind.Extension)
originalFilename = header.Filename
}

// 4. Read full file content
fileContent, err := io.ReadAll(file)
if err != nil {
return nil, ErrFileProcessing
}

	// 5. Determine folder: use config.Folder if set, otherwise "signatures"
	folder := config.Folder
	if folder == "" {
		folder = "signatures"
	}
	key := fmt.Sprintf("%s/%s", strings.Trim(folder, "/"), filename)
url, err := storage.Upload(context.Background(), key, fileContent, kind.MIME.Value)
if err != nil {
return nil, ErrFileProcessing
}

return &UploadedFile{
Filename:     filename,
OriginalName: originalFilename,
Path:         key, // R2 object key — used for deletion
URL:          url,
Size:         int64(len(fileContent)),
MimeType:     kind.MIME.Value,
}, nil
}
