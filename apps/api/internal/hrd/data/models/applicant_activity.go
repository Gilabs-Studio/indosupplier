package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ApplicantActivity represents an activity or event in an applicant's history
type ApplicantActivity struct {
	ID          string          `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	ApplicantID string          `gorm:"type:uuid;not null;index:idx_activity_applicant" json:"applicant_id"`
	Type        string          `gorm:"type:varchar(50);not null" json:"type"` // stage_change, note_added, interview_scheduled, etc.
	Description string          `gorm:"type:text;not null" json:"description"`
	Metadata    *datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional data like from_stage, to_stage, etc.
	CreatedBy   *string         `gorm:"type:uuid" json:"created_by"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (ApplicantActivity) TableName() string {
	return "applicant_activities"
}

// BeforeCreate hook to generate UUID
func (a *ApplicantActivity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// ActivityType constants
const (
	ActivityTypeStageChange        = "stage_change"
	ActivityTypeNoteAdded          = "note_added"
	ActivityTypeInterviewScheduled = "interview_scheduled"
	ActivityTypeInterviewCompleted = "interview_completed"
	ActivityTypeOfferSent          = "offer_sent"
	ActivityTypeOfferAccepted      = "offer_accepted"
	ActivityTypeOfferDeclined      = "offer_declined"
	ActivityTypeHired              = "hired"
	ActivityTypeRejected           = "rejected"
	ActivityTypeCreated            = "created"
	ActivityTypeUpdated            = "updated"
	ActivityTypeResumeUploaded     = "resume_uploaded"
	ActivityTypeRatingChanged      = "rating_changed"
	ActivityTypeConverted          = "converted"
)

// ValidActivityTypes returns all valid activity types
func ValidActivityTypes() []string {
	return []string{
		ActivityTypeStageChange,
		ActivityTypeNoteAdded,
		ActivityTypeInterviewScheduled,
		ActivityTypeInterviewCompleted,
		ActivityTypeOfferSent,
		ActivityTypeOfferAccepted,
		ActivityTypeOfferDeclined,
		ActivityTypeHired,
		ActivityTypeRejected,
		ActivityTypeCreated,
		ActivityTypeUpdated,
		ActivityTypeResumeUploaded,
		ActivityTypeRatingChanged,
		ActivityTypeConverted,
	}
}
