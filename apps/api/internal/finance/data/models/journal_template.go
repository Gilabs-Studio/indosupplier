package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JournalTemplate struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID    string      `gorm:"type:uuid;not null;index" json:"company_id"`
	TemplateName string      `gorm:"type:varchar(255);not null;index" json:"template_name"`
	JournalType  JournalType `gorm:"type:varchar(30);not null;index" json:"journal_type"`
	Description  string      `gorm:"type:text" json:"description"`
	Lines        string      `gorm:"type:jsonb;not null" json:"lines"`
	CreatedBy    *string     `gorm:"type:uuid" json:"created_by"`
	LastUsedAt   *time.Time  `json:"last_used_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (JournalTemplate) TableName() string {
	return "journal_templates"
}

func (m *JournalTemplate) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
