package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OvertimeRequestType represents how the overtime was initiated
type OvertimeRequestType string

const (
	OvertimeTypeAutoDetected OvertimeRequestType = "AUTO_DETECTED" // System detected from clock out
	OvertimeTypeManualClaim  OvertimeRequestType = "MANUAL_CLAIM"  // Employee submitted claim
	OvertimeTypePreApproved  OvertimeRequestType = "PRE_APPROVED"  // Pre-approved before work
)

// OvertimeApprovalStatus represents the approval status
type OvertimeApprovalStatus string

const (
	OvertimeStatusPending  OvertimeApprovalStatus = "PENDING"
	OvertimeStatusApproved OvertimeApprovalStatus = "APPROVED"
	OvertimeStatusRejected OvertimeApprovalStatus = "REJECTED"
	OvertimeStatusCanceled OvertimeApprovalStatus = "CANCELED"
)

// OvertimeRequest represents an overtime request/record
type OvertimeRequest struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID string    `gorm:"type:uuid;not null;index" json:"employee_id"`
	Date       time.Time `gorm:"type:date;not null;index" json:"date"`

	// Request type
	RequestType OvertimeRequestType `gorm:"size:20;not null;default:'AUTO_DETECTED'" json:"request_type"`

	// Time details
	StartTime        time.Time `gorm:"type:timestamptz;not null" json:"start_time"`
	EndTime          time.Time `gorm:"type:timestamptz;not null" json:"end_time"`
	PlannedMinutes   int       `gorm:"default:0" json:"planned_minutes"`  // For pre-approved
	ActualMinutes    int       `gorm:"default:0" json:"actual_minutes"`   // Actual overtime worked
	ApprovedMinutes  int       `gorm:"default:0" json:"approved_minutes"` // Manager can adjust

	// Reason and description
	Reason      string `gorm:"size:500;not null" json:"reason"`
	Description string `gorm:"type:text" json:"description"`
	TaskDetails string `gorm:"type:text" json:"task_details"` // What was worked on

	// Approval workflow
	Status       OvertimeApprovalStatus `gorm:"size:20;not null;default:'PENDING'" json:"status"`
	ApprovedBy   *string                `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt   *time.Time             `gorm:"type:timestamptz" json:"approved_at"`
	RejectedBy   *string                `gorm:"type:uuid" json:"rejected_by"`
	RejectedAt   *time.Time             `gorm:"type:timestamptz" json:"rejected_at"`
	RejectReason string                 `gorm:"size:500" json:"reject_reason"`

	// Link to attendance record
	AttendanceRecordID *string `gorm:"type:uuid;index" json:"attendance_record_id"`

	// Compensation (for payroll integration)
	OvertimeRate       float64 `gorm:"type:decimal(5,2);default:1.50" json:"overtime_rate"`      // e.g., 1.5x for weekday OT
	CompensationAmount float64 `gorm:"type:decimal(15,2);default:0" json:"compensation_amount"` // Calculated amount

	// Notification tracking
	ManagerNotifiedAt *time.Time `gorm:"type:timestamptz" json:"manager_notified_at"`
	IsManagerNotified bool       `gorm:"default:false" json:"is_manager_notified"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (o *OvertimeRequest) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// CalculateActualMinutes calculates actual overtime from start and end time
func (o *OvertimeRequest) CalculateActualMinutes() {
	duration := o.EndTime.Sub(o.StartTime)
	o.ActualMinutes = int(duration.Minutes())
	if o.ActualMinutes < 0 {
		o.ActualMinutes = 0
	}
}
