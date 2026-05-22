package dto

import "time"

// ---- RecruitmentRequest DTOs ----

// CreateRecruitmentRequestDTO represents the request to create a recruitment request
type CreateRecruitmentRequestDTO struct {
	DivisionID        string   `json:"division_id" binding:"required,uuid"`
	PositionID        string   `json:"position_id" binding:"required,uuid"`
	RequiredCount     int      `json:"required_count" binding:"required,min=1"`
	EmploymentType    string   `json:"employment_type" binding:"required,oneof=FULL_TIME PART_TIME CONTRACT INTERN"`
	ExpectedStartDate string   `json:"expected_start_date" binding:"required"` // YYYY-MM-DD
	SalaryRangeMin    *float64 `json:"salary_range_min" binding:"omitempty,gte=0"`
	SalaryRangeMax    *float64 `json:"salary_range_max" binding:"omitempty,gte=0"`
	JobDescription    string   `json:"job_description" binding:"required,max=5000"`
	Qualifications    string   `json:"qualifications" binding:"required,max=5000"`
	Priority          string   `json:"priority" binding:"omitempty,oneof=LOW MEDIUM HIGH URGENT"`
	Notes             *string  `json:"notes" binding:"omitempty,max=2000"`
}

// UpdateRecruitmentRequestDTO represents the request to update a recruitment request
type UpdateRecruitmentRequestDTO struct {
	DivisionID        *string  `json:"division_id" binding:"omitempty,uuid"`
	PositionID        *string  `json:"position_id" binding:"omitempty,uuid"`
	RequiredCount     *int     `json:"required_count" binding:"omitempty,min=1"`
	EmploymentType    *string  `json:"employment_type" binding:"omitempty,oneof=FULL_TIME PART_TIME CONTRACT INTERN"`
	ExpectedStartDate *string  `json:"expected_start_date" binding:"omitempty"` // YYYY-MM-DD
	SalaryRangeMin    *float64 `json:"salary_range_min" binding:"omitempty,gte=0"`
	SalaryRangeMax    *float64 `json:"salary_range_max" binding:"omitempty,gte=0"`
	JobDescription    *string  `json:"job_description" binding:"omitempty,max=5000"`
	Qualifications    *string  `json:"qualifications" binding:"omitempty,max=5000"`
	Priority          *string  `json:"priority" binding:"omitempty,oneof=LOW MEDIUM HIGH URGENT"`
	Notes             *string  `json:"notes" binding:"omitempty,max=2000"`
}

// UpdateRecruitmentStatusDTO represents the request to change recruitment status
type UpdateRecruitmentStatusDTO struct {
	Status string  `json:"status" binding:"required,oneof=PENDING APPROVED REJECTED OPEN CLOSED CANCELLED"`
	Notes  *string `json:"notes" binding:"omitempty,max=2000"`
}

// RejectRecruitmentRequestDTO represents the request body for reject action
type RejectRecruitmentRequestDTO struct {
	Notes *string `json:"notes" binding:"omitempty,max=2000"`
}

// UpdateFilledCountDTO represents the request to update filled positions count
type UpdateFilledCountDTO struct {
	FilledCount int `json:"filled_count" binding:"required,min=0"`
}

// RecruitmentRequestResponse represents the recruitment request list response
type RecruitmentRequestResponse struct {
	ID                string                  `json:"id"`
	RequestCode       string                  `json:"request_code"`
	RequestedByID     string                  `json:"requested_by_id"`
	RequestedBy       *EmployeeSimpleResponse `json:"requested_by,omitempty"`
	RequestDate       string                  `json:"request_date"`
	DivisionID        string                  `json:"division_id"`
	DivisionName      string                  `json:"division_name,omitempty"`
	PositionID        string                  `json:"position_id"`
	PositionName      string                  `json:"position_name,omitempty"`
	RequiredCount     int                     `json:"required_count"`
	FilledCount       int                     `json:"filled_count"`
	OpenPositions     int                     `json:"open_positions"`
	EmploymentType    string                  `json:"employment_type"`
	ExpectedStartDate string                  `json:"expected_start_date"`
	SalaryRangeMin    *float64                `json:"salary_range_min"`
	SalaryRangeMax    *float64                `json:"salary_range_max"`
	JobDescription    string                  `json:"job_description"`
	Qualifications    string                  `json:"qualifications"`
	Priority          string                  `json:"priority"`
	Status            string                  `json:"status"`
	Notes             *string                 `json:"notes"`
	ApprovedByID      *string                 `json:"approved_by_id"`
	ApprovedBy        *EmployeeSimpleResponse `json:"approved_by,omitempty"`
	ApprovedAt        *time.Time              `json:"approved_at"`
	RejectedByID      *string                 `json:"rejected_by_id"`
	RejectedAt        *time.Time              `json:"rejected_at"`
	RejectionNotes    *string                 `json:"rejection_notes"`
	ClosedAt          *time.Time              `json:"closed_at"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
}

// RecruitmentFormDataResponse represents the form data for dropdowns
type RecruitmentFormDataResponse struct {
	Employees       []EmployeeFormOption              `json:"employees"`
	Divisions       []DivisionFormOption              `json:"divisions"`
	JobPositions    []JobPositionFormOption           `json:"job_positions"`
	EmploymentTypes []RecruitmentEmploymentTypeOption `json:"employment_types"`
	Priorities      []RecruitmentPriorityOption       `json:"priorities"`
	Statuses        []RecruitmentStatusOption         `json:"statuses"`
}

// DivisionFormOption represents a division option for dropdowns
type DivisionFormOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JobPositionFormOption represents a job position option for dropdowns
type JobPositionFormOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecruitmentEmploymentTypeOption represents an employment type option for dropdowns
type RecruitmentEmploymentTypeOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// RecruitmentPriorityOption represents a priority option for dropdowns
type RecruitmentPriorityOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// RecruitmentStatusOption represents a status option for dropdowns
type RecruitmentStatusOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
