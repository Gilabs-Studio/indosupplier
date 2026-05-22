package models

import (
	"errors"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeaveStatus represents the status of a leave request
type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "PENDING"
	LeaveStatusApproved  LeaveStatus = "APPROVED"
	LeaveStatusRejected  LeaveStatus = "REJECTED"
	LeaveStatusCancelled LeaveStatus = "CANCELLED"
)

// LeaveDuration represents the duration type of leave
type LeaveDuration string

const (
	LeaveDurationFullDay  LeaveDuration = "FULL_DAY"
	LeaveDurationHalfDay  LeaveDuration = "HALF_DAY"
	LeaveDurationMultiDay LeaveDuration = "MULTI_DAY"
)

// MaxCarryOverDays is the maximum number of leave days that can be carried over to next year
const MaxCarryOverDays = 5

// CarryOverExpiryMonth is the month (1-12) when carry-over leave expires
const CarryOverExpiryMonth = 3 // March 31

// LeaveRequest represents a leave request submitted by an employee
type LeaveRequest struct {
	ID          string        `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID  string        `gorm:"type:uuid;not null;index:idx_leave_employee" json:"employee_id"`
	LeaveTypeID string        `gorm:"type:uuid;not null;index:idx_leave_type" json:"leave_type_id"`
	StartDate   time.Time     `gorm:"type:date;not null;index:idx_leave_dates" json:"start_date"`
	EndDate     time.Time     `gorm:"type:date;not null;index:idx_leave_dates" json:"end_date"`
	Duration    LeaveDuration `gorm:"size:20;not null;default:'FULL_DAY'" json:"duration"`
	TotalDays   float64       `gorm:"type:decimal(4,2);not null" json:"total_days"` // Total calendar days (inclusive)
	Reason      string        `gorm:"type:text;not null" json:"reason"`
	Status      LeaveStatus   `gorm:"size:20;not null;default:'PENDING';index:idx_leave_status" json:"status"`

	// Attachment for supporting documents (e.g., medical certificate)
	AttachmentURL *string `gorm:"size:500" json:"attachment_url"`

	// Approval/Rejection details
	ApprovedBy     *string    `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt     *time.Time `gorm:"type:timestamptz" json:"approved_at"`
	RejectedBy     *string    `gorm:"type:uuid" json:"rejected_by"`
	RejectionNote  *string    `gorm:"type:text" json:"rejection_note"`
	ApprovalNotes  *string    `gorm:"type:text" json:"approval_notes"`
	RejectionNotes *string    `gorm:"type:text" json:"rejection_notes"`
	RejectedAt     *time.Time `gorm:"type:timestamptz" json:"rejected_at"`

	// Carry-over tracking
	// WHY: Track leave from previous year that hasn't expired (max 5 days until March 31)
	IsCarryOver         bool       `gorm:"default:false" json:"is_carry_over"`
	RemainingCarryOver  float64    `gorm:"type:decimal(4,2);default:0" json:"remaining_carry_over"`
	CarryOverExpiryDate *time.Time `gorm:"type:date" json:"carry_over_expiry_date"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // Soft delete for audit trail

	// Audit fields
	CreatedBy *string `gorm:"type:uuid" json:"created_by"`
	UpdatedBy *string `gorm:"type:uuid" json:"updated_by"`
}

// TableName specifies the table name
func (LeaveRequest) TableName() string {
	return "leave_requests"
}

// BeforeCreate hook to generate UUID and validate dates
// WHY: Ensure data integrity before insertion
func (lr *LeaveRequest) BeforeCreate(tx *gorm.DB) error {
	if lr.ID == "" {
		lr.ID = uuid.New().String()
	}

	// Validate date range
	if !lr.EndDate.After(lr.StartDate) && !lr.EndDate.Equal(lr.StartDate) {
		return errors.New("end_date must be on or after start_date")
	}

	// Validate TotalDays based on Duration
	if lr.Duration == LeaveDurationHalfDay && lr.TotalDays != 0.5 {
		return errors.New("half_day leave must have total_days = 0.5")
	}

	if lr.Duration == LeaveDurationFullDay && lr.TotalDays != 1.0 {
		return errors.New("full_day leave must have total_days = 1.0")
	}

	// Set default status if not set
	if lr.Status == "" {
		lr.Status = LeaveStatusPending
	}

	return nil
}

// BeforeUpdate hook to validate state transitions
// WHY: Ensure only valid status transitions are allowed
func (lr *LeaveRequest) BeforeUpdate(tx *gorm.DB) error {
	// Get the old value from database
	var oldLeave LeaveRequest
	if err := tx.Where("id = ?", lr.ID).First(&oldLeave).Error; err != nil {
		return err
	}

	// Validate status transitions
	// WHY: Prevent invalid workflow transitions (e.g., APPROVED → PENDING)
	if oldLeave.Status == LeaveStatusApproved && lr.Status == LeaveStatusPending {
		return errors.New("cannot revert approved leave to pending")
	}

	// Allow REJECTED → APPROVED and CANCELLED → APPROVED for re-approve flow
	// These transitions are valid when HR re-approves an accidentally rejected/cancelled leave

	// If status is changing to APPROVED, ensure ApprovedBy and ApprovedAt are set
	if lr.Status == LeaveStatusApproved && oldLeave.Status != LeaveStatusApproved {
		if lr.ApprovedBy == nil || lr.ApprovedAt == nil {
			return errors.New("approved_by and approved_at must be set when approving")
		}
	}

	return nil
}

// IsEditable checks if the leave request can be edited
// WHY: Only PENDING leaves can be edited by employee (REJECTED cannot be edited, must create new request)
func (lr *LeaveRequest) IsEditable() bool {
	return lr.Status == LeaveStatusPending
}

// CanBeApproved checks if the leave request can be approved
// WHY: PENDING, REJECTED, and CANCELLED leaves can be (re-)approved
func (lr *LeaveRequest) CanBeApproved() bool {
	return lr.Status == LeaveStatusPending || lr.Status == LeaveStatusRejected || lr.Status == LeaveStatusCancelled
}

// CanBeCancelled checks if the leave request can be cancelled
// WHY: Only PENDING or APPROVED (if not yet started) leaves can be cancelled
func (lr *LeaveRequest) CanBeCancelled() bool {
	if lr.Status == LeaveStatusPending {
		return true
	}
	if lr.Status == LeaveStatusApproved {
		// Can cancel if leave hasn't started yet
		return apptime.Now().Before(lr.StartDate)
	}
	return false
}
