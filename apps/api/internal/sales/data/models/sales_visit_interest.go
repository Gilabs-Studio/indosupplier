package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesVisitInterestQuestion represents a survey question for scoring interest
type SalesVisitInterestQuestion struct {
	ID           string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	QuestionText string `gorm:"type:text;not null" json:"question_text"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`
	Sequence     int    `gorm:"default:0" json:"sequence"`

	// Relations
	Options []SalesVisitInterestOption `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalesVisitInterestQuestion) TableName() string {
	return "sales_visit_interest_questions"
}

func (q *SalesVisitInterestQuestion) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	return nil
}

// SalesVisitInterestOption represents a possible answer for a question with a score
type SalesVisitInterestOption struct {
	ID         string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	QuestionID string `gorm:"type:uuid;not null;index" json:"question_id"`
	OptionText string `gorm:"type:varchar(255);not null" json:"option_text"`
	Score      int    `gorm:"default:0" json:"score"` // e.g., Yes=1, No=0

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalesVisitInterestOption) TableName() string {
	return "sales_visit_interest_options"
}

func (o *SalesVisitInterestOption) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// SalesVisitInterestAnswer represents the selected answer for a specific visit detail
type SalesVisitInterestAnswer struct {
	ID               string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesVisitDetailID string `gorm:"type:uuid;not null;index" json:"sales_visit_detail_id"`
	QuestionID       string `gorm:"type:uuid;not null;index" json:"question_id"`
	OptionID         string `gorm:"type:uuid;not null;index" json:"option_id"`
	
	// Denormalized for easier querying/display
	Score int `gorm:"default:0" json:"score"` 

	// Relations
	Question *SalesVisitInterestQuestion `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Option   *SalesVisitInterestOption   `gorm:"foreignKey:OptionID" json:"option,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
}

func (SalesVisitInterestAnswer) TableName() string {
	return "sales_visit_interest_answers"
}

func (a *SalesVisitInterestAnswer) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}
