package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactRole represents role classification for contacts (Director, PIC, Manager, etc.)
type ContactRole struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code        string         `gorm:"type:varchar(50);not null;index" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	BadgeColor  string         `gorm:"type:varchar(20)" json:"badge_color"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for ContactRole
func (ContactRole) TableName() string {
	return "crm_contact_roles"
}

// BeforeCreate hook to generate UUID
func (c *ContactRole) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
