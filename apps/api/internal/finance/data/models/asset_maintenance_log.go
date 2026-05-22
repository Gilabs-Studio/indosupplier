package models

import (
	"time"

	"github.com/google/uuid"
)

// MaintenanceType enumeration
type MaintenanceType string

const (
	MaintenanceTypePreventive MaintenanceType = "PREVENTIVE"
	MaintenanceTypeCorrective MaintenanceType = "CORRECTIVE"
	MaintenanceTypeEmergency  MaintenanceType = "EMERGENCY"
)

// AssetMaintenanceLog model
type AssetMaintenanceLog struct {
	ID      uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	AssetID uuid.UUID `gorm:"type:uuid;not null;index" json:"asset_id"`

	MaintenanceType MaintenanceType `gorm:"type:varchar(50);index" json:"maintenance_type"`
	MaintenanceDate time.Time       `gorm:"type:date;not null;index" json:"maintenance_date"`

	Description *string  `gorm:"type:text" json:"description"`
	Cost        *float64 `gorm:"type:numeric(18,2)" json:"cost"`

	ServiceProvider *string `gorm:"type:varchar(200)" json:"service_provider"`
	PartsReplaced   *string `gorm:"type:text" json:"parts_replaced"`

	DurationHours       *float64   `gorm:"type:numeric(10,2)" json:"duration_hours"`
	NextMaintenanceDate *time.Time `gorm:"type:date" json:"next_maintenance_date"`

	RecordedBy *uuid.UUID `gorm:"type:uuid" json:"recorded_by"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relations
	Asset *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// TableName specifies the table name
func (AssetMaintenanceLog) TableName() string {
	return "asset_maintenance_logs"
}
