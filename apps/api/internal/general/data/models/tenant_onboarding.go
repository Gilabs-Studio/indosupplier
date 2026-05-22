package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantOnboarding stores the one-time setup state for a tenant.
// One record per tenant; created lazily on first access.
type TenantOnboarding struct {
	ID           string         `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID     string         `gorm:"type:uuid;not null;uniqueIndex" json:"tenant_id"`
	BusinessType string         `gorm:"type:varchar(50);not null;default:''" json:"business_type"` // e.g. "fnb", "retail", "other"
	Completed    bool           `gorm:"default:false" json:"completed"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TenantOnboarding) TableName() string {
	return "tenant_onboardings"
}

func (t *TenantOnboarding) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
