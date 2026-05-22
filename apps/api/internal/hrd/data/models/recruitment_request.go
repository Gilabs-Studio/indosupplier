package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecruitmentStatus represents the status of a recruitment request
type RecruitmentStatus string

const (
	RecruitmentStatusDraft     RecruitmentStatus = "DRAFT"
	RecruitmentStatusPending   RecruitmentStatus = "PENDING"
	RecruitmentStatusApproved  RecruitmentStatus = "APPROVED"
	RecruitmentStatusRejected  RecruitmentStatus = "REJECTED"
	RecruitmentStatusOpen      RecruitmentStatus = "OPEN"
	RecruitmentStatusClosed    RecruitmentStatus = "CLOSED"
	RecruitmentStatusCancelled RecruitmentStatus = "CANCELLED"
)

// RecruitmentPriority represents the urgency of a recruitment request
type RecruitmentPriority string

const (
	RecruitmentPriorityLow    RecruitmentPriority = "LOW"
	RecruitmentPriorityMedium RecruitmentPriority = "MEDIUM"
	RecruitmentPriorityHigh   RecruitmentPriority = "HIGH"
	RecruitmentPriorityUrgent RecruitmentPriority = "URGENT"
)

// RecruitmentEmploymentType represents the employment type for the position
type RecruitmentEmploymentType string

const (
	RecruitmentEmploymentFullTime RecruitmentEmploymentType = "FULL_TIME"
	RecruitmentEmploymentPartTime RecruitmentEmploymentType = "PART_TIME"
	RecruitmentEmploymentContract RecruitmentEmploymentType = "CONTRACT"
	RecruitmentEmploymentIntern   RecruitmentEmploymentType = "INTERN"
)

// RecruitmentRequest represents a request to hire new employees
type RecruitmentRequest struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	RequestCode string `gorm:"type:varchar(50);not null;uniqueIndex" json:"request_code"`

	// Requester
	RequestedByID string    `gorm:"type:uuid;not null;index:idx_recruitment_requester" json:"requested_by_id"`
	RequestDate   time.Time `gorm:"type:date;not null;index:idx_recruitment_date" json:"request_date"`

	// Position details
	DivisionID string `gorm:"type:uuid;not null;index:idx_recruitment_division" json:"division_id"`
	PositionID string `gorm:"type:uuid;not null;index:idx_recruitment_position" json:"position_id"`

	// Requirements
	RequiredCount     int                       `gorm:"not null;default:1" json:"required_count"`
	FilledCount       int                       `gorm:"not null;default:0" json:"filled_count"`
	EmploymentType    RecruitmentEmploymentType `gorm:"type:varchar(20);not null;default:'FULL_TIME'" json:"employment_type"`
	ExpectedStartDate time.Time                 `gorm:"type:date;not null" json:"expected_start_date"`
	SalaryRangeMin    *float64                  `gorm:"type:decimal(15,2)" json:"salary_range_min"`
	SalaryRangeMax    *float64                  `gorm:"type:decimal(15,2)" json:"salary_range_max"`

	// Description
	JobDescription string  `gorm:"type:text;not null" json:"job_description"`
	Qualifications string  `gorm:"type:text;not null" json:"qualifications"`
	Notes          *string `gorm:"type:text" json:"notes"`

	// Priority & Status
	Priority RecruitmentPriority `gorm:"type:varchar(20);not null;default:'MEDIUM';index:idx_recruitment_priority" json:"priority"`
	Status   RecruitmentStatus   `gorm:"type:varchar(20);not null;default:'DRAFT';index:idx_recruitment_status" json:"status"`

	// Approval
	ApprovedByID   *string    `gorm:"type:uuid" json:"approved_by_id"`
	ApprovedAt     *time.Time `gorm:"type:timestamptz" json:"approved_at"`
	RejectedByID   *string    `gorm:"type:uuid" json:"rejected_by_id"`
	RejectedAt     *time.Time `gorm:"type:timestamptz" json:"rejected_at"`
	RejectionNotes *string    `gorm:"type:text" json:"rejection_notes"`
	ClosedAt       *time.Time `gorm:"type:timestamptz" json:"closed_at"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Audit
	CreatedBy *string `gorm:"type:uuid" json:"created_by"`
	UpdatedBy *string `gorm:"type:uuid" json:"updated_by"`
}

// TableName specifies the table name
func (RecruitmentRequest) TableName() string {
	return "recruitment_requests"
}

// BeforeCreate hook to generate UUID and validate
func (r *RecruitmentRequest) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	if r.RequiredCount < 1 {
		return errors.New("required_count must be at least 1")
	}

	if !r.ExpectedStartDate.After(r.RequestDate) && !r.ExpectedStartDate.Equal(r.RequestDate) {
		return errors.New("expected_start_date must be on or after request_date")
	}

	if r.SalaryRangeMin != nil && r.SalaryRangeMax != nil && *r.SalaryRangeMin > *r.SalaryRangeMax {
		return errors.New("salary_range_min must not exceed salary_range_max")
	}

	if r.Status == "" {
		r.Status = RecruitmentStatusDraft
	}

	return nil
}

// OpenPositions returns the number of unfilled positions
func (r *RecruitmentRequest) OpenPositions() int {
	open := r.RequiredCount - r.FilledCount
	if open < 0 {
		return 0
	}
	return open
}

// IsEditable returns whether the recruitment request can be edited
// WHY: DRAFT and REJECTED requests can be modified; REJECTED allows fixing issues before resubmission
func (r *RecruitmentRequest) IsEditable() bool {
	return r.Status == RecruitmentStatusDraft || r.Status == RecruitmentStatusRejected
}

// CanBeSubmitted returns whether the request can be submitted for approval
func (r *RecruitmentRequest) CanBeSubmitted() bool {
	return r.Status == RecruitmentStatusDraft
}

// CanBeApproved returns whether the request can be approved
func (r *RecruitmentRequest) CanBeApproved() bool {
	return r.Status == RecruitmentStatusPending
}

// CanBeRejected returns whether the request can be rejected
func (r *RecruitmentRequest) CanBeRejected() bool {
	return r.Status == RecruitmentStatusPending
}

// CanBeOpened returns whether the request can be opened for hiring
func (r *RecruitmentRequest) CanBeOpened() bool {
	return r.Status == RecruitmentStatusApproved
}

// CanBeClosed returns whether the request can be closed
func (r *RecruitmentRequest) CanBeClosed() bool {
	return r.Status == RecruitmentStatusOpen
}

// CanBeCancelled returns whether the request can be cancelled
func (r *RecruitmentRequest) CanBeCancelled() bool {
	return r.Status == RecruitmentStatusDraft || r.Status == RecruitmentStatusPending
}

// ValidStatusTransitions defines allowed status transitions
var ValidRecruitmentStatusTransitions = map[RecruitmentStatus][]RecruitmentStatus{
	RecruitmentStatusDraft:    {RecruitmentStatusPending, RecruitmentStatusCancelled},
	RecruitmentStatusPending:  {RecruitmentStatusApproved, RecruitmentStatusRejected, RecruitmentStatusCancelled},
	RecruitmentStatusApproved: {RecruitmentStatusOpen},
	RecruitmentStatusOpen:     {RecruitmentStatusClosed},
	RecruitmentStatusRejected: {RecruitmentStatusPending}, // Resubmit after fixing
}

// IsValidTransition checks if a status transition is valid
func (r *RecruitmentRequest) IsValidTransition(newStatus RecruitmentStatus) bool {
	validTargets, ok := ValidRecruitmentStatusTransitions[r.Status]
	if !ok {
		return false
	}
	for _, target := range validTargets {
		if target == newStatus {
			return true
		}
	}
	return false
}
