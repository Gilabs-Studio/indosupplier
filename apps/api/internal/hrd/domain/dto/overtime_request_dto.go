package dto

// OvertimeRequest DTOs

// CreateOvertimeRequestDTO represents the request to create an overtime request
type CreateOvertimeRequestDTO struct {
	EmployeeID  string `json:"employee_id" binding:"omitempty,uuid"` // Optional, for admin/HR to create for other employees
	Date        string `json:"date" binding:"required"`              // YYYY-MM-DD
	StartTime   string `json:"start_time" binding:"required"`        // HH:MM
	EndTime     string `json:"end_time" binding:"required"`          // HH:MM
	Reason      string `json:"reason" binding:"omitempty,max=500"`
	Description string `json:"description"`
	TaskDetails string `json:"task_details"`
	RequestType string `json:"request_type" binding:"required,oneof=AUTO_DETECTED MANUAL_CLAIM PRE_APPROVED"`
}

// UpdateOvertimeRequestDTO represents the request to update an overtime request
type UpdateOvertimeRequestDTO struct {
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Reason      *string `json:"reason" binding:"omitempty,max=500"`
	Description *string `json:"description"`
	TaskDetails *string `json:"task_details"`
}

// ApproveOvertimeRequest represents the request to approve overtime
type ApproveOvertimeRequest struct {
	ApprovedMinutes int `json:"approved_minutes" binding:"required,gte=0"` // Can adjust minutes
}

// RejectOvertimeRequest represents the request to reject overtime
type RejectOvertimeRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// ListOvertimeRequestsRequest represents the request to list overtime requests
type ListOvertimeRequestsRequest struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PerPage     int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	EmployeeID  string `form:"employee_id" binding:"omitempty,uuid"`
	Status      string `form:"status" binding:"omitempty,oneof=PENDING APPROVED REJECTED CANCELED"`
	RequestType string `form:"request_type" binding:"omitempty,oneof=AUTO_DETECTED MANUAL_CLAIM PRE_APPROVED"`
	DateFrom    string `form:"date_from"`
	DateTo      string `form:"date_to"`
	SortBy      string `form:"sort_by"`
	SortOrder   string `form:"sort_order" binding:"omitempty,oneof=asc desc ASC DESC"`
}

// OvertimeRequestResponse represents the response for an overtime request
type OvertimeRequestResponse struct {
	ID                 string  `json:"id"`
	EmployeeID         string  `json:"employee_id"`
	EmployeeName       string  `json:"employee_name"`
	EmployeeCode       string  `json:"employee_code"`
	DivisionName       string  `json:"division_name,omitempty"`
	Date               string  `json:"date"`
	RequestType        string  `json:"request_type"`
	StartTime          string  `json:"start_time"`
	EndTime            string  `json:"end_time"`
	PlannedMinutes     int     `json:"planned_minutes"`
	PlannedHours       string  `json:"planned_hours"` // Formatted
	ActualMinutes      int     `json:"actual_minutes"`
	ActualHours        string  `json:"actual_hours"` // Formatted
	ApprovedMinutes    int     `json:"approved_minutes"`
	ApprovedHours      string  `json:"approved_hours"` // Formatted
	Reason             string  `json:"reason"`
	Description        string  `json:"description"`
	TaskDetails        string  `json:"task_details"`
	Status             string  `json:"status"`
	ApprovedBy         *string `json:"approved_by,omitempty"`
	ApprovedByName     string  `json:"approved_by_name,omitempty"`
	ApprovedAt         *string `json:"approved_at,omitempty"`
	RejectedBy         *string `json:"rejected_by,omitempty"`
	RejectedByName     string  `json:"rejected_by_name,omitempty"`
	RejectedAt         *string `json:"rejected_at,omitempty"`
	RejectReason       string  `json:"reject_reason"`
	AttendanceRecordID *string `json:"attendance_record_id,omitempty"`
	OvertimeRate       float64 `json:"overtime_rate"`
	CompensationAmount float64 `json:"compensation_amount"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

// OvertimeSummaryResponse represents overtime summary for an employee
type OvertimeSummaryResponse struct {
	EmployeeID            string `json:"employee_id"`
	Year                  int    `json:"year"`
	Month                 int    `json:"month"`
	TotalRequestedMinutes int    `json:"total_requested_minutes"`
	TotalApprovedMinutes  int    `json:"total_approved_minutes"`
	TotalRejectedMinutes  int    `json:"total_rejected_minutes"`
	PendingRequests       int    `json:"pending_requests"`
	ApprovedRequests      int    `json:"approved_requests"`
	RejectedRequests      int    `json:"rejected_requests"`
}

// PendingOvertimeNotification represents a pending overtime for notification
type PendingOvertimeNotification struct {
	OvertimeRequest OvertimeRequestResponse `json:"overtime_request"`
	EmployeeName    string                  `json:"employee_name"`
	DivisionName    string                  `json:"division_name"`
}
