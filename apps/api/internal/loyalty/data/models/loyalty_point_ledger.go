package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LedgerEntryType string

const (
	LedgerEntryTypeEarned   LedgerEntryType = "earned"
	LedgerEntryTypeRedeemed LedgerEntryType = "redeemed"
	LedgerEntryTypeExpired  LedgerEntryType = "expired"
	LedgerEntryTypeAdjusted LedgerEntryType = "adjusted"
)

// LoyaltyPointLedger is an append-only audit log of every point mutation.
// It captures a balance snapshot (BalanceAfter) after each entry for full auditability.
type LoyaltyPointLedger struct {
	ID    string          `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	MemberID  string          `gorm:"type:uuid;not null;index" json:"member_id"`
	EntryType LedgerEntryType `gorm:"type:varchar(20);not null;index" json:"entry_type"`
	// TransactionType identifies the originating entity (pos_order / sales_order / manual).
	TransactionType *string `gorm:"type:varchar(50)" json:"transaction_type,omitempty"`
	TransactionID   *string `gorm:"type:uuid;index" json:"transaction_id,omitempty"`
	// Points is positive for earned entries and negative for redeemed/expired/adjusted entries.
	Points         int64    `gorm:"not null" json:"points"`
	BalanceAfter   int64    `gorm:"not null" json:"balance_after"`
	LifetimeAfter  int64    `gorm:"not null" json:"lifetime_after"`
	Multiplier     *float64 `gorm:"type:numeric(5,2)" json:"multiplier,omitempty"`
	RewardID       *string  `gorm:"type:varchar(100)" json:"reward_id,omitempty"`
	RewardName     *string  `gorm:"type:varchar(255)" json:"reward_name,omitempty"`
	Notes          *string  `gorm:"type:text" json:"notes,omitempty"`
	ExpiresAt      *time.Time `gorm:"type:timestamptz;index" json:"expires_at,omitempty"`
	ProcessedBy    *string  `gorm:"type:uuid" json:"processed_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

func (LoyaltyPointLedger) TableName() string { return "loyalty_point_ledgers" }

func (l *LoyaltyPointLedger) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
