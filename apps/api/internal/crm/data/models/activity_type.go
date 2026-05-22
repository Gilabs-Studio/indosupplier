package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityType represents dynamic activity classification (Visit, Call, Email, Meeting, etc.)
type ActivityType struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index;uniqueIndex:uq_crm_activity_types_tenant_code" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex:uq_crm_activity_types_tenant_code" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	BadgeColor  string         `gorm:"type:varchar(20)" json:"badge_color"`
	Order       int            `gorm:"default:0" json:"order"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for ActivityType
func (ActivityType) TableName() string {
	return "crm_activity_types"
}

// BeforeCreate hook to generate UUID
func (a *ActivityType) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}
