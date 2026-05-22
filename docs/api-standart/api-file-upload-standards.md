# API File Upload Standards

This document outlines the standards for handling file uploads, storage, and processing in the API.

## 1. Supported File Types
- **Images Only**: Currently, only image uploads are supported.
- **Allowed Formats**: JPEG, PNG, GIF, WebP.
- **Output Format**: All images MUST be converted to **WebP** for storage to ensure optimized size and performance.

## 2. Security Standards
### 2.1 Validation
- **MIME Type**: Must be validated against the allowed list.
- **Extension**: Must match the MIME type.
- **Magic Bytes**: The first 512 bytes MUST be checked using a library (e.g., `h2non/filetype`) to verify the actual file content matches the extension.
- **Max Size**: Strictly enforced (default: **10MB**).

### 2.2 Storage Security
- **Filename**: MUST use a generated **UUID** to prevent path traversal and collision. (e.g., `550e8400-e29b-41d4-a716-446655440000.webp`)
- **Location**: Files MUST be stored outside the application code directory or in a dedicated `uploads` directory.
- **Permissions**: Uploaded files MUST NOT be executable (Set max permission to `0644`).
- **Path Traversal**: Validate final path is within the intended upload directory.

## 3. Implementation Guidelines
### 3.1 Upload Utility
- Use a central utility for handling uploads to ensure consistent security checks.
- Do not handle file system operations directly in handlers.

```go
// Example: Centralized Upload Utility
func SaveUploadedFile(file multipart.File, header *multipart.FileHeader, config Config) (*UploadedFile, error) {
    // 1. Validate Size
    // 2. Validate Magic Bytes
    // 3. Convert to WebP
    // 4. Generate UUID Filename
    // 5. Save with 0644 permissions
}
```

### 3.2 Response Format
Upload endpoints should return standard metadata about the saved file.

```json
{
  "success": true,
  "data": {
    "filename": "uuid.webp",
    "original_name": "photo.jpg",
    "url": "/uploads/uuid.webp",
    "size": 12345,
    "mime_type": "image/webp"
  }
}
```

## 4. Scalability (Horizontal Scaling)
- **Local Storage**: For development or single-instance deployments, local disk is acceptable.
- **Production**: MUST use a shared storage solution if running multiple replicas:
  - **Shared Volume**: NFS / EFS mounted to all instances.
  - **Object Storage**: S3 / R2 / MinIO (preferred for large scale).

## 5. Compression
- **Library**: Use efficient Go libraries (e.g., `github.com/kolesa-team/go-webp`).
- **Quality**: Target 85% quality for lossy compression which offers a good balance between size and visual fidelity.
