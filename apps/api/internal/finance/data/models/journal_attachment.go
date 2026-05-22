package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JournalAttachment stores uploaded supporting evidence for journal entries.
type JournalAttachment struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	JournalEntryID string        `gorm:"type:uuid;not null;index" json:"journal_entry_id"`
	JournalEntry   *JournalEntry `gorm:"foreignKey:JournalEntryID" json:"journal_entry,omitempty"`

	FileName string `gorm:"type:varchar(255);not null" json:"file_name"`
	FileURL  string `gorm:"type:varchar(500);not null" json:"file_url"`
	MimeType string `gorm:"type:varchar(120)" json:"mime_type"`
	FileSize int64  `gorm:"type:bigint;default:0" json:"file_size"`

	UploadedBy *string `gorm:"type:uuid" json:"uploaded_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (JournalAttachment) TableName() string {
	return "journal_attachments"
}

func (ja *JournalAttachment) BeforeCreate(tx *gorm.DB) error {
	if ja.ID == "" {
		ja.ID = uuid.New().String()
	}
	return nil
}
