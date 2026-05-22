package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeadStatus represents lead lifecycle states (New, Contacted, Qualified, Converted, Lost)
type LeadStatus struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index;uniqueIndex:uq_crm_lead_statuses_tenant_code" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex:uq_crm_lead_statuses_tenant_code" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	Score       int            `gorm:"default:0" json:"score"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	Order       int            `gorm:"default:0" json:"order"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	IsConverted bool           `gorm:"default:false" json:"is_converted"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for LeadStatus
func (LeadStatus) TableName() string {
	return "crm_lead_statuses"
}

// BeforeCreate hook to generate UUID
func (l *LeadStatus) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
