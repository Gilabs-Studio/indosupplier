package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmployeeSignature represents an employee's digital signature image
// This model is managed by HR/Admin through the employee management module
type EmployeeSignature struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID string    `gorm:"type:uuid;not null;uniqueIndex" json:"employee_id"`

	// File information
	FilePath string `gorm:"type:varchar(500);not null" json:"file_path"`
	FileName string `gorm:"type:varchar(255);not null" json:"file_name"`
	FileSize int64  `gorm:"not null" json:"file_size"`
	FileHash string `gorm:"type:varchar(64);not null" json:"file_hash"` // SHA-256 for integrity check
	MimeType string `gorm:"type:varchar(50);not null" json:"mime_type"`

	// Image dimensions
	Width  int `json:"width"`
	Height int `json:"height"`

	// Upload information (HR/Admin who uploaded)
	UploadedBy string    `gorm:"type:uuid;not null" json:"uploaded_by"`
	UploadedAt time.Time `json:"uploaded_at"`

	// Soft delete untuk history (requirement #3)
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for EmployeeSignature
func (EmployeeSignature) TableName() string {
	return "employee_signatures"
}

// IsActive checks if the signature is active (not soft deleted)
func (es *EmployeeSignature) IsActive() bool {
	return es.DeletedAt.Time.IsZero()
}
