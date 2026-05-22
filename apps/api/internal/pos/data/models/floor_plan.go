package models

import (
	"time"

	"gorm.io/gorm"
)

// FloorPlanStatus enum values
const (
	FloorPlanStatusDraft     = "draft"
	FloorPlanStatusPublished = "published"
)

// FloorPlan represents a single floor layout for an outlet
type FloorPlan struct {
	ID          string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	OutletID    string         `gorm:"type:uuid;not null;index:idx_floor_plans_outlet;uniqueIndex:uidx_floor_plans_outlet_floor_active,priority:1,where:deleted_at IS NULL" json:"outlet_id"`
	CompanyID   *string        `gorm:"type:uuid;index:idx_floor_plans_company" json:"company_id,omitempty"`
	Name        string         `gorm:"type:varchar(200);not null" json:"name"`
	FloorNumber int            `gorm:"type:int;not null;default:1;uniqueIndex:uidx_floor_plans_outlet_floor_active,priority:2,where:deleted_at IS NULL" json:"floor_number"`
	Status      string         `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	GridSize    int            `gorm:"type:int;not null;default:20" json:"grid_size"`
	SnapToGrid  bool           `gorm:"type:boolean;not null;default:true" json:"snap_to_grid"`
	Width       int            `gorm:"type:int;not null;default:1200" json:"width"`
	Height      int            `gorm:"type:int;not null;default:800" json:"height"`
	LayoutData  string         `gorm:"type:jsonb;default:'[]'" json:"layout_data"`
	Version     int            `gorm:"type:int;not null;default:0" json:"version"`
	PublishedAt *time.Time     `gorm:"type:timestamptz" json:"published_at"`
	PublishedBy *string        `gorm:"type:varchar(100)" json:"published_by"`
	CreatedBy   *string        `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName overrides the default table name
func (FloorPlan) TableName() string {
	return "pos_floor_plans"
}
