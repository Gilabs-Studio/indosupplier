package models

import "time"

// LayoutVersion stores immutable snapshots created when a floor plan is published
type LayoutVersion struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TenantID       string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	FloorPlanID    string    `gorm:"type:uuid;not null;index:idx_layout_versions_floor_plan" json:"floor_plan_id"`
	Version        int       `gorm:"type:int;not null" json:"version"`
	LayoutData     string    `gorm:"type:jsonb;not null" json:"layout_data"`
	PublishedAt    time.Time `gorm:"type:timestamptz;not null" json:"published_at"`
	PublishedBy    string    `gorm:"type:varchar(100);not null" json:"published_by"`
	// PublishedByName is populated via JOIN with the users table and never persisted.
	PublishedByName string `gorm:"->:false;column:published_by_name" json:"-"`
}

// TableName overrides the default table name
func (LayoutVersion) TableName() string {
	return "pos_layout_versions"
}
