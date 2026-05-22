package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CheckInType represents the type of check-in
type CheckInType string

const (
	CheckInTypeNormal    CheckInType = "NORMAL"
	CheckInTypeWFH       CheckInType = "WFH"        // Work From Home
	CheckInTypeFieldWork CheckInType = "FIELD_WORK" // Working on field/client site
)

// AttendanceStatus represents the attendance status
type AttendanceStatus string

const (
	AttendanceStatusPresent  AttendanceStatus = "PRESENT"
	AttendanceStatusAbsent   AttendanceStatus = "ABSENT"
	AttendanceStatusLate     AttendanceStatus = "LATE"
	AttendanceStatusHalfDay  AttendanceStatus = "HALF_DAY"
	AttendanceStatusHoliday  AttendanceStatus = "HOLIDAY"
	AttendanceStatusLeave    AttendanceStatus = "LEAVE"
	AttendanceStatusWFH      AttendanceStatus = "WFH"
	AttendanceStatusOffDay   AttendanceStatus = "OFF_DAY" // Non-working day per schedule
)

// AttendanceRecord represents a single attendance entry for an employee
type AttendanceRecord struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID string    `gorm:"type:uuid;not null;index" json:"employee_id"`
	Date       time.Time `gorm:"type:date;not null;index" json:"date"`

	// Check-in details
	CheckInTime      *time.Time  `gorm:"type:timestamptz" json:"check_in_time"`
	CheckInType      CheckInType `gorm:"size:20;default:'NORMAL'" json:"check_in_type"`
	CheckInLatitude  *float64    `gorm:"type:decimal(10,8)" json:"check_in_latitude"`
	CheckInLongitude *float64    `gorm:"type:decimal(11,8)" json:"check_in_longitude"`
	CheckInAddress   string      `gorm:"size:500" json:"check_in_address"`
	CheckInNote      string      `gorm:"size:500" json:"check_in_note"`

	// Check-out details
	CheckOutTime      *time.Time `gorm:"type:timestamptz" json:"check_out_time"`
	CheckOutLatitude  *float64   `gorm:"type:decimal(10,8)" json:"check_out_latitude"`
	CheckOutLongitude *float64   `gorm:"type:decimal(11,8)" json:"check_out_longitude"`
	CheckOutAddress   string     `gorm:"size:500" json:"check_out_address"`
	CheckOutNote      string     `gorm:"size:500" json:"check_out_note"`

	// Calculated fields
	Status           AttendanceStatus `gorm:"size:20;not null;default:'ABSENT'" json:"status"`
	WorkingMinutes   int              `gorm:"default:0" json:"working_minutes"`   // Total worked minutes
	OvertimeMinutes  int              `gorm:"default:0" json:"overtime_minutes"`  // Overtime in minutes (auto-calculated)
	LateMinutes      int              `gorm:"default:0" json:"late_minutes"`      // Minutes late
	EarlyLeaveMinutes int             `gorm:"default:0" json:"early_leave_minutes"` // Minutes left early
	BreakMinutes     int              `gorm:"default:0" json:"break_minutes"`     // Actual break taken

	// Reference to work schedule used
	WorkScheduleID string `gorm:"type:uuid;index" json:"work_schedule_id"`

	// If on leave, reference the leave request
	LeaveRequestID *string `gorm:"type:uuid;index" json:"leave_request_id"`

	// Late reason — required when employee clocks in late (NORMAL type)
	LateReason string `gorm:"size:500" json:"late_reason"`
	// Photo proof URL — required for WFH / FIELD_WORK clock-in (camera capture)
	PhotoURL string `gorm:"size:500" json:"photo_url"`

	// Notes and flags
	Notes             string `gorm:"type:text" json:"notes"`
	IsManualEntry     bool   `gorm:"default:false" json:"is_manual_entry"`     // If entry was manually created by admin
	ManualEntryReason string `gorm:"size:500" json:"manual_entry_reason"`
	ApprovedBy        *string `gorm:"type:uuid" json:"approved_by"` // For manual entries requiring approval

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (a *AttendanceRecord) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// CalculateWorkingMinutes calculates the total working minutes
func (a *AttendanceRecord) CalculateWorkingMinutes() {
	if a.CheckInTime == nil || a.CheckOutTime == nil {
		a.WorkingMinutes = 0
		return
	}
	duration := a.CheckOutTime.Sub(*a.CheckInTime)
	a.WorkingMinutes = int(duration.Minutes()) - a.BreakMinutes
	if a.WorkingMinutes < 0 {
		a.WorkingMinutes = 0
	}
}
