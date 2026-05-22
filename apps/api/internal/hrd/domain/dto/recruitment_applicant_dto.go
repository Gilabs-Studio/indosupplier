package dto

import "time"

// ---- ApplicantStage DTOs ----

// ApplicantStageResponse represents a stage in the pipeline
type ApplicantStageResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Order    int    `json:"order"`
	IsWon    bool   `json:"is_won"`
	IsLost   bool   `json:"is_lost"`
	IsActive bool   `json:"is_active"`
}

// ---- RecruitmentApplicant DTOs ----

// CreateRecruitmentApplicantDTO represents the request to create an applicant
type CreateRecruitmentApplicantDTO struct {
	RecruitmentRequestID string  `json:"recruitment_request_id" binding:"required,uuid"`
	StageID              string  `json:"stage_id" binding:"required,uuid"`
	FullName             string  `json:"full_name" binding:"required,max=255"`
	Email                string  `json:"email" binding:"required,email,max=255"`
	Phone                *string `json:"phone" binding:"omitempty,max=20"`
	Source               string  `json:"source" binding:"required,oneof=linkedin jobstreet glints referral direct other"`
	Notes                *string `json:"notes" binding:"omitempty,max=5000"`
	ResumeURL            *string `json:"resume_url" binding:"omitempty,max=500"`
	LinkedinURL          *string `json:"linkedin_url" binding:"omitempty,max=500"`
}

// UpdateRecruitmentApplicantDTO represents the request to update an applicant
type UpdateRecruitmentApplicantDTO struct {
	FullName    *string `json:"full_name" binding:"omitempty,max=255"`
	Email       *string `json:"email" binding:"omitempty,email,max=255"`
	Phone       *string `json:"phone" binding:"omitempty,max=20"`
	Source      *string `json:"source" binding:"omitempty,oneof=linkedin jobstreet glints referral direct other"`
	Notes       *string `json:"notes" binding:"omitempty,max=5000"`
	ResumeURL   *string `json:"resume_url" binding:"omitempty,max=500"`
	LinkedinURL *string `json:"linkedin_url" binding:"omitempty,max=500,linkedinurl"`
	Rating      *int    `json:"rating" binding:"omitempty,min=1,max=5"`
}

// MoveApplicantStageDTO represents the request to move an applicant to a different stage
type MoveApplicantStageDTO struct {
	ToStageID string  `json:"to_stage_id" binding:"omitempty,uuid"`
	StageID   string  `json:"stage_id" binding:"omitempty,uuid"`
	Reason    *string `json:"reason" binding:"omitempty,max=1000"`
	Notes     *string `json:"notes" binding:"omitempty,max=2000"`
}

// TargetStageID returns the target stage ID from either supported request field.
func (d MoveApplicantStageDTO) TargetStageID() string {
	if d.ToStageID != "" {
		return d.ToStageID
	}

	return d.StageID
}

// EmployeeSummaryResponse represents a brief employee response for applicant conversion
type EmployeeSummaryResponse struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
	Email        string `json:"email"`
}

// RecruitmentApplicantResponse represents an applicant in API responses
type RecruitmentApplicantResponse struct {
	ID                   string                   `json:"id"`
	RecruitmentRequestID string                   `json:"recruitment_request_id"`
	StageID              string                   `json:"stage_id"`
	FullName             string                   `json:"full_name"`
	Email                string                   `json:"email"`
	Phone                *string                  `json:"phone"`
	ResumeURL            *string                  `json:"resume_url"`
	LinkedinURL          *string                  `json:"linkedin_url"`
	Source               string                   `json:"source"`
	AppliedAt            time.Time                `json:"applied_at"`
	LastActivityAt       time.Time                `json:"last_activity_at"`
	Rating               *int                     `json:"rating"`
	Notes                *string                  `json:"notes"`
	Stage                *ApplicantStageResponse  `json:"stage,omitempty"`
	EmployeeID           *string                  `json:"employee_id,omitempty"`
	Employee             *EmployeeSummaryResponse `json:"employee,omitempty"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
}

// ApplicantsByStageResponse represents applicants grouped by stage for Kanban board
type ApplicantsByStageResponse struct {
	StageID    string                          `json:"stage_id"`
	StageName  string                          `json:"stage_name"`
	StageColor string                          `json:"stage_color"`
	Order      int                             `json:"order"`
	Applicants []*RecruitmentApplicantResponse `json:"applicants"`
	Total      int64                           `json:"total"`
}

// ---- ApplicantActivity DTOs ----

// CreateApplicantActivityDTO represents the request to create an activity
type CreateApplicantActivityDTO struct {
	Type        string         `json:"type" binding:"required,oneof=stage_change note_added interview_scheduled interview_completed offer_sent offer_accepted offer_declined hired rejected created updated resume_uploaded rating_changed"`
	Description string         `json:"description" binding:"required,max=2000"`
	Metadata    map[string]any `json:"metadata" binding:"omitempty"`
}

// ApplicantActivityResponse represents an activity in API responses
type ApplicantActivityResponse struct {
	ID            string         `json:"id"`
	ApplicantID   string         `json:"applicant_id"`
	Type          string         `json:"type"`
	Description   string         `json:"description"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedBy     *string        `json:"created_by"`
	CreatedByName *string        `json:"created_by_name,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// ---- List/Filter Params ----

// ListApplicantsParams represents query parameters for listing applicants
type ListApplicantsParams struct {
	Page                 int    `form:"page" binding:"omitempty,min=1"`
	PerPage              int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search               string `form:"search" binding:"omitempty,max=100"`
	RecruitmentRequestID string `form:"recruitment_request_id" binding:"omitempty,uuid"`
	StageID              string `form:"stage_id" binding:"omitempty,uuid"`
	Source               string `form:"source" binding:"omitempty,oneof=linkedin jobstreet glints referral direct other"`
}

// ListApplicantsByStageParams represents query parameters for listing applicants by stage
type ListApplicantsByStageParams struct {
	StageID              string `form:"stage_id" binding:"omitempty,uuid"`
	RecruitmentRequestID string `form:"recruitment_request_id" binding:"omitempty,uuid"`
	Page                 int    `form:"page" binding:"omitempty,min=1"`
	PerPage              int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search               string `form:"search" binding:"omitempty,max=100"`
}

// ConvertApplicantToEmployeeDTO represents the request to convert applicant to employee
// All fields are optional - if not provided, backend will use applicant data and defaults
type ConvertApplicantToEmployeeDTO struct {
	// Pre-filled from applicant (editable)
	Name          string  `json:"name" binding:"omitempty,max=200"`
	Email         string  `json:"email" binding:"omitempty,email,max=100"`
	Phone         string  `json:"phone" binding:"max=20"`
	DivisionID    *string `json:"division_id"`
	JobPositionID *string `json:"job_position_id"`

	// Additional employee fields (optional, user can fill later in employee module)
	NIK          string  `json:"nik" binding:"omitempty,max=20"`
	DateOfBirth  *string `json:"date_of_birth"` // Format: "2006-01-02"
	PlaceOfBirth string  `json:"place_of_birth" binding:"omitempty,max=100"`
	Gender       string  `json:"gender" binding:"omitempty,oneof=male female"`
	Religion     string  `json:"religion" binding:"omitempty,max=50"`
	Address      string  `json:"address" binding:"omitempty,max=500"`
	VillageID    *string `json:"village_id"`

	// Contract info (optional, can be added later in employee module)
	ContractType   string `json:"contract_type" binding:"omitempty,oneof=PKWTT PKWT Intern"`
	ContractNumber string `json:"contract_number" binding:"omitempty,max=50"`
	StartDate      string `json:"start_date" binding:"omitempty"` // Format: "2006-01-02"
	EndDate        string `json:"end_date"`                       // Optional
}
