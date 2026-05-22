package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetAttachment represents a file attachment for a fixed asset
type AssetAttachment struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       *uuid.UUID     `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	AssetID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"asset_id"`
	AttachableType *string        `gorm:"type:varchar(50);index" json:"attachable_type,omitempty"`
	AttachableID   *uuid.UUID     `gorm:"type:uuid;index" json:"attachable_id,omitempty"`
	FileName       string         `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath       string         `gorm:"type:varchar(500);not null" json:"file_path"`
	FileURL        string         `gorm:"type:varchar(500);not null" json:"file_url"`
	FileType       string         `gorm:"type:varchar(20);not null" json:"file_type"` // invoice, warranty, photo, manual, other
	FileSize       *int           `gorm:"type:integer" json:"file_size,omitempty"`
	MimeType       *string        `gorm:"type:varchar(100)" json:"mime_type,omitempty"`
	Description    *string        `gorm:"type:text" json:"description,omitempty"`
	UploadedBy     *uuid.UUID     `gorm:"type:uuid;index" json:"uploaded_by,omitempty"`
	UploadedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"uploaded_at"`
	CreatedAt      time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Asset Asset `gorm:"foreignKey:AssetID" json:"-"`
	User  *User `gorm:"foreignKey:UploadedBy" json:"uploaded_by_user,omitempty"`
}

// TableName specifies the table name for AssetAttachment
func (AssetAttachment) TableName() string {
	return "asset_attachments"
}

// BeforeCreate hook untuk generate UUID
func (a *AssetAttachment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
