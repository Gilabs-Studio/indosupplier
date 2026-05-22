package models

import (
	"errors"
	"time"

	organizationModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecruitmentApplicant represents a job applicant for a recruitment request
type RecruitmentApplicant struct {
	ID                   string `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	RecruitmentRequestID string `gorm:"type:uuid;not null;index:idx_applicant_request" json:"recruitment_request_id"`
	StageID              string `gorm:"type:uuid;not null;index:idx_applicant_stage" json:"stage_id"`

	// Personal Info
	FullName    string  `gorm:"type:varchar(255);not null" json:"full_name"`
	Email       string  `gorm:"type:varchar(255);not null" json:"email"`
	Phone       *string `gorm:"type:varchar(20)" json:"phone"`
	ResumeURL   *string `gorm:"type:varchar(500)" json:"resume_url"`
	LinkedinURL *string `gorm:"type:varchar(500)" json:"linkedin_url"`   // LinkedIn profile URL
	Source      string  `gorm:"type:varchar(50);not null" json:"source"` // linkedin, jobstreet, glints, referral, direct, other

	// Tracking
	AppliedAt      time.Time `gorm:"not null" json:"applied_at"`
	LastActivityAt time.Time `gorm:"not null" json:"last_activity_at"`

	// Evaluation
	Rating *int    `gorm:"type:smallint" json:"rating"` // 1-5 stars
	Notes  *string `gorm:"type:text" json:"notes"`

	// Relations
	Stage              ApplicantStage     `gorm:"foreignKey:StageID" json:"stage,omitempty"`
	RecruitmentRequest RecruitmentRequest `gorm:"foreignKey:RecruitmentRequestID" json:"recruitment_request,omitempty"`

	// Link to Employee (set when converted)
	EmployeeID *string                      `gorm:"type:uuid;index;column:employee_id" json:"employee_id"`
	Employee   *organizationModels.Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Audit
	CreatedBy *string `gorm:"type:uuid" json:"created_by"`
	UpdatedBy *string `gorm:"type:uuid" json:"updated_by"`
}

// TableName specifies the table name
func (RecruitmentApplicant) TableName() string {
	return "recruitment_applicants"
}

// BeforeCreate hook to generate UUID and validate
func (a *RecruitmentApplicant) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	if a.FullName == "" {
		return errors.New("full_name is required")
	}

	if a.Email == "" {
		return errors.New("email is required")
	}

	if a.AppliedAt.IsZero() {
		a.AppliedAt = time.Now()
	}

	if a.LastActivityAt.IsZero() {
		a.LastActivityAt = time.Now()
	}

	return nil
}

// BeforeUpdate updates the last activity timestamp
func (a *RecruitmentApplicant) BeforeUpdate(tx *gorm.DB) error {
	a.LastActivityAt = time.Now()
	return nil
}

// IsInTerminalStage returns true if applicant is in Hired or Rejected stage
func (a *RecruitmentApplicant) IsInTerminalStage() bool {
	return a.Stage.IsWon || a.Stage.IsLost
}

// ApplicantSource constants
const (
	ApplicantSourceLinkedIn  = "linkedin"
	ApplicantSourceJobStreet = "jobstreet"
	ApplicantSourceGlints    = "glints"
	ApplicantSourceReferral  = "referral"
	ApplicantSourceDirect    = "direct"
	ApplicantSourceOther     = "other"
)

// ValidApplicantSources returns all valid applicant sources
func ValidApplicantSources() []string {
	return []string{
		ApplicantSourceLinkedIn,
		ApplicantSourceJobStreet,
		ApplicantSourceGlints,
		ApplicantSourceReferral,
		ApplicantSourceDirect,
		ApplicantSourceOther,
	}
}

// IsValidSource checks if a source is valid
func IsValidSource(source string) bool {
	for _, s := range ValidApplicantSources() {
		if s == source {
			return true
		}
	}
	return false
}
