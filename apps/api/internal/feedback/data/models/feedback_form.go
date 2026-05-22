package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FeedbackForm stores the JSON schema for a feedback questionnaire
// scoped to one outlet. One outlet can have one active form at a time.
type FeedbackForm struct {
	ID          string         `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	OutletID    string         `gorm:"type:uuid;not null;index" json:"outlet_id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	// SchemaJSON stores the questions array as a validated JSONB document.
	// Schema: {"questions":[{"id":"q1","type":"rating|multiple_choice|text","label":"...","required":bool,"config":{...}}]}
	SchemaJSON  datatypes.JSON `gorm:"type:jsonb;not null" json:"schema_json"`
	Version     int            `gorm:"default:1" json:"version"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedBy   *string        `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy   *string        `gorm:"type:uuid" json:"updated_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FeedbackForm) TableName() string { return "feedback_forms" }
