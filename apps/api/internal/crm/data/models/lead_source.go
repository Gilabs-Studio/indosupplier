package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeadSource represents the origin of leads (Website, Referral, Cold Call, etc.)
type LeadSource struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index;uniqueIndex:uq_crm_lead_sources_tenant_code" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex:uq_crm_lead_sources_tenant_code" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	Order       int            `gorm:"default:0" json:"order"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for LeadSource
func (LeadSource) TableName() string {
	return "crm_lead_sources"
}

// BeforeCreate hook to generate UUID
func (l *LeadSource) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
