package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DashboardLayout stores a user's personalized dashboard widget configuration.
// One record per (user_id, dashboard_type) pair, enforced by a unique index.
type DashboardLayout struct {
	ID            string         `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	UserID        string         `gorm:"type:uuid;not null;uniqueIndex:idx_dashboard_layout_user_type" json:"user_id"`
	DashboardType string         `gorm:"type:varchar(50);not null;default:'general';uniqueIndex:idx_dashboard_layout_user_type" json:"dashboard_type"`
	LayoutJSON    string         `gorm:"type:jsonb;not null" json:"layout_json"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate auto-generates the UUID primary key if not set.
func (d *DashboardLayout) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
