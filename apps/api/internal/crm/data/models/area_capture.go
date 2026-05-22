package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AreaCapture represents a GPS data point captured during field visits.
// Used for area mapping, coverage analysis, and visit heatmaps.
type AreaCapture struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	VisitReportID *string        `gorm:"type:uuid;index" json:"visit_report_id"`                          // FK to crm visit_reports
	CaptureType   string         `gorm:"type:varchar(20);not null" json:"capture_type"`                   // check_in, check_out, manual
	Latitude      float64        `gorm:"type:decimal(10,7);not null" json:"latitude"`                     // GPS latitude
	Longitude     float64        `gorm:"type:decimal(10,7);not null" json:"longitude"`                    // GPS longitude
	Address       string         `gorm:"type:text" json:"address"`                                        // Reverse-geocoded address
	Accuracy      float64        `gorm:"type:decimal(10,2)" json:"accuracy"`                              // GPS accuracy in meters
	AreaID        *string        `gorm:"type:uuid;index" json:"area_id"`                                  // FK to areas (resolved territory)
	CapturedAt    time.Time      `gorm:"not null" json:"captured_at"`                                     // Time of capture
	CapturedBy    *string        `gorm:"type:uuid;index" json:"captured_by"`                              // FK to employees
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	VisitReport *VisitReport `gorm:"foreignKey:VisitReportID" json:"visit_report,omitempty"`
}

// TableName specifies the table name for AreaCapture
func (AreaCapture) TableName() string {
	return "crm_area_captures"
}

// BeforeCreate hook to generate UUID
func (ac *AreaCapture) BeforeCreate(tx *gorm.DB) error {
	if ac.ID == "" {
		ac.ID = uuid.New().String()
	}
	return nil
}
