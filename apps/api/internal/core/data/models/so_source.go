package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SOSource represents a sales order source (channel)
type SOSource struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code        string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SOSource
func (SOSource) TableName() string {
	return "so_sources"
}

// BeforeCreate hook to generate UUID
func (s *SOSource) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
