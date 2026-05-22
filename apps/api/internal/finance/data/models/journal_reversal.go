package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JournalReversal stores linkage between an original posted journal and its reversal entry.
type JournalReversal struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	OriginalJournalEntryID string        `gorm:"type:uuid;not null;index" json:"original_journal_entry_id"`
	OriginalJournalEntry   *JournalEntry `gorm:"foreignKey:OriginalJournalEntryID" json:"original_journal_entry,omitempty"`

	ReversalJournalEntryID string        `gorm:"type:uuid;not null;uniqueIndex" json:"reversal_journal_entry_id"`
	ReversalJournalEntry   *JournalEntry `gorm:"foreignKey:ReversalJournalEntryID" json:"reversal_journal_entry,omitempty"`

	Reason    string  `gorm:"type:text" json:"reason"`
	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (JournalReversal) TableName() string {
	return "journal_reversals"
}

func (jr *JournalReversal) BeforeCreate(tx *gorm.DB) error {
	if jr.ID == "" {
		jr.ID = uuid.New().String()
	}
	return nil
}
