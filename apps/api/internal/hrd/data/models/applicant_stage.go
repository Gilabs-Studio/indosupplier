package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicantStage represents a stage in the applicant pipeline (Kanban column)
type ApplicantStage struct {
	ID       string `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Color    string `gorm:"type:varchar(7);not null;default:'#3b82f6'" json:"color"`
	Order    int    `gorm:"not null;default:0" json:"order"`
	IsWon    bool   `gorm:"not null;default:false" json:"is_won"`   // Final stage (Hired)
	IsLost   bool   `gorm:"not null;default:false" json:"is_lost"`  // Final stage (Rejected)
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name
func (ApplicantStage) TableName() string {
	return "applicant_stages"
}

// BeforeCreate hook to generate UUID
func (s *ApplicantStage) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// IsTerminal returns true if this is a final stage (won or lost)
func (s *ApplicantStage) IsTerminal() bool {
	return s.IsWon || s.IsLost
}

// DefaultApplicantStages returns the default stages for a new recruitment pipeline
func DefaultApplicantStages() []ApplicantStage {
	return []ApplicantStage{
		{ID: uuid.New().String(), Name: "New", Color: "#6b7280", Order: 0, IsActive: true},
		{ID: uuid.New().String(), Name: "Screening", Color: "#3b82f6", Order: 1, IsActive: true},
		{ID: uuid.New().String(), Name: "Interview", Color: "#f59e0b", Order: 2, IsActive: true},
		{ID: uuid.New().String(), Name: "Offer", Color: "#8b5cf6", Order: 3, IsActive: true},
		{ID: uuid.New().String(), Name: "Hired", Color: "#22c55e", Order: 4, IsWon: true, IsActive: true},
		{ID: uuid.New().String(), Name: "Rejected", Color: "#ef4444", Order: 5, IsLost: true, IsActive: true},
	}
}
