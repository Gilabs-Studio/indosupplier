package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdjustmentJournalApprovalAction string

const (
	AdjustmentJournalApprovalActionSubmitted AdjustmentJournalApprovalAction = "submitted"
	AdjustmentJournalApprovalActionApproved  AdjustmentJournalApprovalAction = "approved"
	AdjustmentJournalApprovalActionRejected  AdjustmentJournalApprovalAction = "rejected"
)

type AdjustmentJournalApproval struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	JournalID string                          `gorm:"type:uuid;not null;index" json:"journal_id"`
	Action    AdjustmentJournalApprovalAction `gorm:"type:varchar(20);not null;index" json:"action"`
	ActorID   string                          `gorm:"type:uuid;not null;index" json:"actor_id"`
	Notes     string                          `gorm:"type:text" json:"notes"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AdjustmentJournalApproval) TableName() string {
	return "adjustment_journal_approvals"
}

func (m *AdjustmentJournalApproval) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
