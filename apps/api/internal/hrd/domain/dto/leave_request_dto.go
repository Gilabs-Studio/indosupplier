package dto

import "time"

// CreateLeaveRequestDTO represents the request to create a new leave request
type CreateLeaveRequestDTO struct {
	EmployeeID    string  `json:"employee_id" binding:"required,uuid"`
	LeaveTypeID   string  `json:"leave_type_id" binding:"required,uuid"`
	StartDate     string  `json:"start_date" binding:"required"` // Format: YYYY-MM-DD
	EndDate       string  `json:"end_date" binding:"required"`   // Format: YYYY-MM-DD
	Duration      string  `json:"duration" binding:"required,oneof=FULL_DAY HALF_DAY MULTI_DAY"`
	Reason        string  `json:"reason" binding:"required,min=10,max=1000"`
	AttachmentURL *string `json:"attachment_url" binding:"omitempty,url"`
}

// CreateMyLeaveRequestDTO represents self-service leave request payload.
// EmployeeID is resolved from authenticated user context.
type CreateMyLeaveRequestDTO struct {
	LeaveTypeID   string  `json:"leave_type_id" binding:"required,uuid"`
	StartDate     string  `json:"start_date" binding:"required"` // Format: YYYY-MM-DD
	EndDate       string  `json:"end_date" binding:"required"`   // Format: YYYY-MM-DD
	Duration      string  `json:"duration" binding:"required,oneof=FULL_DAY HALF_DAY MULTI_DAY"`
	Reason        string  `json:"reason" binding:"required,min=10,max=1000"`
	AttachmentURL *string `json:"attachment_url" binding:"omitempty,url"`
}

// UpdateLeaveRequestDTO represents the request to update a leave request
// WHY: All fields are optional (pointers) to support partial updates
type UpdateLeaveRequestDTO struct {
	LeaveTypeID   *string `json:"leave_type_id" binding:"omitempty,uuid"`
	StartDate     *string `json:"start_date" binding:"omitempty"`
	EndDate       *string `json:"end_date" binding:"omitempty"`
	Duration      *string `json:"duration" binding:"omitempty,oneof=FULL_DAY HALF_DAY MULTI_DAY"`
	Reason        *string `json:"reason" binding:"omitempty,min=10,max=1000"`
	AttachmentURL *string `json:"attachment_url" binding:"omitempty,url"`
}

// LeaveRequestResponseDTO represents the response for a leave request (list view)
type LeaveRequestResponseDTO struct {
	ID           string  `json:"id"`
	EmployeeName string  `json:"employee_name"`
	LeaveType    string  `json:"leave_type"`
	StartDate    string  `json:"start_date"` // Format: YYYY-MM-DD
	EndDate      string  `json:"end_date"`   // Format: YYYY-MM-DD
	Duration     string  `json:"duration"`
	TotalDays    float64 `json:"total_days"`
	Reason       string  `json:"reason"`
	Status       string  `json:"status"`

	// Rejection details (for REJECTED status)
	RejectedBy     *string `json:"rejected_by,omitempty"`
	RejectedByName string  `json:"rejected_by_name,omitempty"`
	RejectionNote  *string `json:"rejection_note,omitempty"`

	// Timestamps
	CreatedAt string `json:"created_at"` // Format: RFC3339
	UpdatedAt string `json:"updated_at"` // Format: RFC3339
}

// LeaveRequestDetailResponseDTO represents the detailed response for a single leave request
type LeaveRequestDetailResponseDTO struct {
	ID            string  `json:"id"`
	StartDate     string  `json:"start_date"` // Format: YYYY-MM-DD
	EndDate       string  `json:"end_date"`   // Format: YYYY-MM-DD
	Duration      string  `json:"duration"`
	TotalDays     float64 `json:"total_days"`
	Reason        string  `json:"reason"`
	Status        string  `json:"status"`
	AttachmentURL *string `json:"attachment_url,omitempty"`

	// Employee details
	Employee EmployeeDetailDTO `json:"employee"`

	// Leave Type details
	LeaveType LeaveTypeDetailDTO `json:"leave_type"`

	// Approval details
	ApprovedBy    *string `json:"approved_by,omitempty"`
	ApprovedAt    *string `json:"approved_at,omitempty"` // Format: RFC3339
	RejectedBy    *string `json:"rejected_by,omitempty"`
	RejectionNote *string `json:"rejection_note,omitempty"`

	// Carry-over details
	IsCarryOver         bool    `json:"is_carry_over"`
	RemainingCarryOver  float64 `json:"remaining_carry_over"`
	CarryOverExpiryDate *string `json:"carry_over_expiry_date,omitempty"` // Format: YYYY-MM-DD

	// Timestamps
	CreatedAt string  `json:"created_at"` // Format: RFC3339
	UpdatedAt string  `json:"updated_at"` // Format: RFC3339
	CreatedBy *string `json:"created_by,omitempty"`
	UpdatedBy *string `json:"updated_by,omitempty"`
}

// EmployeeDetailDTO represents employee information in leave request detail
type EmployeeDetailDTO struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Phone        *string `json:"phone,omitempty"`
	EmployeeCode string  `json:"employee_code"`
	JobPosition  *string `json:"job_position,omitempty"`
	Division     *string `json:"division,omitempty"`
}

// LeaveTypeDetailDTO represents leave type information in leave request detail
type LeaveTypeDetailDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Code             string `json:"code"`
	Description      string `json:"description"`
	MaxDays          int    `json:"max_days"`
	IsPaid           bool   `json:"is_paid"`
	IsCutAnnualLeave bool   `json:"is_cut_annual_leave"`
}

// LeaveBalanceResponseDTO represents the employee's leave balance
type LeaveBalanceResponseDTO struct {
	EmployeeID          string  `json:"employee_id"`
	TotalQuota          int     `json:"total_quota"`            // From Employee.TotalLeaveQuota
	UsedDays            int     `json:"used_days"`              // Sum of approved leave where IsCutAnnualLeave = true
	RemainingBalance    int     `json:"remaining_balance"`      // TotalQuota - UsedDays
	CarryOverBalance    float64 `json:"carry_over_balance"`     // Remaining from previous year (max 5 days)
	CarryOverExpiryDate *string `json:"carry_over_expiry_date"` // Format: YYYY-MM-DD (March 31)
	TotalAvailableLeave int     `json:"total_available_leave"`  // RemainingBalance + CarryOverBalance
	PendingRequestsDays int     `json:"pending_requests_days"`  // Days in pending requests
	CalculatedAt        string  `json:"calculated_at"`          // Format: RFC3339
}

// ApproveLeaveRequestDTO represents the request to approve a leave request
type ApproveLeaveRequestDTO struct {
	ApprovedBy *string `json:"approved_by" binding:"omitempty,uuid"` // Optional, will use current user if not provided
}

// RejectLeaveRequestDTO represents the request to reject a leave request
type RejectLeaveRequestDTO struct {
	RejectionNote string  `json:"rejection_note" binding:"required,min=10,max=500"`
	RejectedBy    *string `json:"rejected_by" binding:"omitempty,uuid"` // Optional, will use current user if not provided
}

// CancelLeaveRequestDTO represents the request to cancel a leave request
type CancelLeaveRequestDTO struct {
	CancellationNote *string `json:"cancellation_note" binding:"omitempty,min=10,max=500"` // Optional note for cancellation
	CancelledBy      *string `json:"cancelled_by" binding:"omitempty,uuid"`                // Optional, will use current user if not provided
}

// LeaveRequestListFilterDTO represents filters for listing leave requests
type LeaveRequestListFilterDTO struct {
	EmployeeID *string `form:"employee_id" binding:"omitempty,uuid"`
	Status     *string `form:"status" binding:"omitempty"`     // PENDING, APPROVED, REJECTED, CANCELLED (case-insensitive)
	StartDate  *string `form:"start_date" binding:"omitempty"` // Format: YYYY-MM-DD
	EndDate    *string `form:"end_date" binding:"omitempty"`   // Format: YYYY-MM-DD
	Search     *string `form:"search" binding:"omitempty"`     // Search by employee name, leave type, or reason
	Page       int     `form:"page" binding:"omitempty,min=1"`
	PerPage    int     `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// LeaveRequestAuditTrailUser represents actor info in audit trail entries.
type LeaveRequestAuditTrailUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// LeaveRequestAuditTrailEntry represents one audit trail row for leave request detail.
type LeaveRequestAuditTrailEntry struct {
	ID             string                      `json:"id"`
	Action         string                      `json:"action"`
	PermissionCode string                      `json:"permission_code"`
	TargetID       string                      `json:"target_id"`
	Metadata       map[string]interface{}      `json:"metadata"`
	User           *LeaveRequestAuditTrailUser `json:"user"`
	CreatedAt      time.Time                   `json:"created_at"`
}

// FormDataResponseDTO represents data for leave request form
type FormDataResponseDTO struct {
	Employees  []FormEmployeeDTO  `json:"employees"`
	LeaveTypes []FormLeaveTypeDTO `json:"leave_types"`
}

// FormEmployeeDTO represents employee option for form dropdown with their leave balance
type FormEmployeeDTO struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	EmployeeCode     string  `json:"employee_code"`
	RemainingBalance float64 `json:"remaining_balance"` // Available leave balance for this employee
}

// FormLeaveTypeDTO represents leave type option for form dropdown
type FormLeaveTypeDTO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	MaxDays int    `json:"max_days"`
}
