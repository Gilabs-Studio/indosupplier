package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BreakTime represents a single break period
type BreakTime struct {
	StartTime string `json:"start_time"` // Format: "12:00"
	EndTime   string `json:"end_time"`   // Format: "13:00"
}

// Breaks is a custom type for handling JSON array of break times
type Breaks []BreakTime

// Value implements the driver.Valuer interface for database storage
func (b Breaks) Value() (driver.Value, error) {
	if b == nil {
		return nil, nil
	}
	return json.Marshal(b)
}

// Scan implements the sql.Scanner interface for database retrieval
func (b *Breaks) Scan(value interface{}) error {
	if value == nil {
		*b = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan type into Breaks")
	}

	return json.Unmarshal(bytes, b)
}

// GormDataType returns the gorm data type for this type
func (Breaks) GormDataType() string {
	return "jsonb"
}

// WorkSchedule represents a work schedule configuration
// Following international ERP standards with flexible hours support
type WorkSchedule struct {
	ID          string  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string  `gorm:"size:100;not null" json:"name"`
	Description string  `gorm:"size:255" json:"description"`
	DivisionID  *string `gorm:"type:uuid;index" json:"division_id"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`

	// Standard work hours
	StartTime string `gorm:"size:5;not null" json:"start_time"` // Format: "08:00"
	EndTime   string `gorm:"size:5;not null" json:"end_time"`   // Format: "17:00"

	// Flexible hours configuration
	IsFlexible        bool   `gorm:"default:false" json:"is_flexible"`
	FlexibleStartTime string `gorm:"size:5" json:"flexible_start_time"` // e.g., "07:00" - can clock in from this time
	FlexibleEndTime   string `gorm:"size:5" json:"flexible_end_time"`   // e.g., "09:00" - must clock in before this time

	// Break times (stored as JSON array)
	Breaks Breaks `gorm:"type:jsonb" json:"breaks"`

	// Working days (bitmask: 1=Mon, 2=Tue, 4=Wed, 8=Thu, 16=Fri, 32=Sat, 64=Sun)
	// Example: 31 = Mon-Fri (1+2+4+8+16)
	WorkingDays int `gorm:"default:31" json:"working_days"`

	// Working hours per day (calculated from start_time and end_time)
	WorkingHoursPerDay float64 `gorm:"type:decimal(4,2);default:8.00" json:"working_hours_per_day"`

	// Tolerance settings (in minutes)
	LateToleranceMinutes       int `gorm:"default:0" json:"late_tolerance_minutes"`
	EarlyLeaveToleranceMinutes int `gorm:"default:0" json:"early_leave_tolerance_minutes"`

	// GPS Settings
	RequireGPS      bool    `gorm:"default:false" json:"require_gps"`
	GPSRadiusMeter  float64 `gorm:"type:decimal(10,2);default:100.00" json:"gps_radius_meter"` // Tolerance radius in meters
	OfficeLatitude  float64 `gorm:"type:decimal(10,8)" json:"office_latitude"`
	OfficeLongitude float64 `gorm:"type:decimal(11,8)" json:"office_longitude"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (w *WorkSchedule) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// IsWorkingDay checks if a given weekday is a working day
// weekday: 0=Sunday, 1=Monday, ..., 6=Saturday
func (w *WorkSchedule) IsWorkingDay(weekday int) bool {
	// Convert to our bitmask format (Monday=1, Sunday=64)
	var dayBit int
	switch weekday {
	case 0: // Sunday
		dayBit = 64
	case 1: // Monday
		dayBit = 1
	case 2: // Tuesday
		dayBit = 2
	case 3: // Wednesday
		dayBit = 4
	case 4: // Thursday
		dayBit = 8
	case 5: // Friday
		dayBit = 16
	case 6: // Saturday
		dayBit = 32
	}
	return w.WorkingDays&dayBit != 0
}

// CalculateWorkingHours calculates working hours from start and end time
func (w *WorkSchedule) CalculateWorkingHours() float64 {
	if w.StartTime == "" || w.EndTime == "" {
		return 8.0
	}

	// Parse times
	startHour := int(w.StartTime[0]-'0')*10 + int(w.StartTime[1]-'0')
	startMin := int(w.StartTime[3]-'0')*10 + int(w.StartTime[4]-'0')
	endHour := int(w.EndTime[0]-'0')*10 + int(w.EndTime[1]-'0')
	endMin := int(w.EndTime[3]-'0')*10 + int(w.EndTime[4]-'0')

	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	diffMinutes := endMinutes - startMinutes
	if diffMinutes < 0 {
		diffMinutes += 24 * 60 // Handle overnight shifts
	}

	return float64(diffMinutes) / 60.0
}
