package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ValuationRunStatus represents the lifecycle status of a valuation run.
type ValuationRunStatus string

const (
	ValuationRunStatusDraft           ValuationRunStatus = "draft"
	ValuationRunStatusPendingApproval ValuationRunStatus = "pending_approval"
	ValuationRunStatusApproved        ValuationRunStatus = "approved"
	ValuationRunStatusPosted          ValuationRunStatus = "posted"
	ValuationRunStatusNoDifference    ValuationRunStatus = "no_difference"
	ValuationRunStatusFailed          ValuationRunStatus = "failed"
)

// ValuationType represents the kind of valuation being run.
type ValuationType string

const (
	ValuationTypeInventory    ValuationType = "inventory"
	ValuationTypeFX           ValuationType = "fx"
	ValuationTypeDepreciation ValuationType = "depreciation"
)

// ValuationRun tracks each valuation run with lifecycle status, totals, and link to generated journal.
type ValuationRun struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	ReferenceID   string             `gorm:"type:varchar(255);uniqueIndex;not null" json:"reference_id"`
	ValuationType ValuationType      `gorm:"type:varchar(50);not null;index" json:"valuation_type"`
	PeriodStart   time.Time          `gorm:"type:date;not null" json:"period_start"`
	PeriodEnd     time.Time          `gorm:"type:date;not null" json:"period_end"`
	Status        ValuationRunStatus `gorm:"type:varchar(30);not null;default:'draft';index" json:"status"`

	TotalDebit  float64 `gorm:"type:decimal(18,2);default:0" json:"total_debit"`
	TotalCredit float64 `gorm:"type:decimal(18,2);default:0" json:"total_credit"`
	TotalDelta  float64 `gorm:"type:decimal(18,2);default:0" json:"total_delta"`

	JournalEntryID *string       `gorm:"type:uuid" json:"journal_entry_id"`
	JournalEntry   *JournalEntry `gorm:"foreignKey:JournalEntryID" json:"journal_entry,omitempty"`

	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"`

	// Period Locking (prevents re-run after posting)
	IsLocked bool       `gorm:"type:boolean;default:false;index" json:"is_locked"`
	LockedAt *time.Time `json:"locked_at,omitempty"`

	// Approval Tracking (audit trail)
	ApprovedBy    *string    `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	ApprovalNotes string     `gorm:"type:text" json:"approval_notes,omitempty"`

	CreatedBy   *string    `gorm:"type:uuid" json:"created_by"`
	CompletedAt *time.Time `json:"completed_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Details []ValuationRunDetail `gorm:"foreignKey:ValuationRunID;constraint:OnDelete:CASCADE" json:"details,omitempty"`
}

func (ValuationRun) TableName() string {
	return "valuation_runs"
}

func (vr *ValuationRun) BeforeCreate(tx *gorm.DB) error {
	if vr.ID == "" {
		vr.ID = uuid.New().String()
	}
	return nil
}
