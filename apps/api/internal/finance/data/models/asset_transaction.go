package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetTransactionType string

const (
	AssetTransactionTypeAcquire    AssetTransactionType = "acquire"
	AssetTransactionTypeUpdate     AssetTransactionType = "update"
	AssetTransactionTypeDepreciate AssetTransactionType = "depreciate"
	AssetTransactionTypeTransfer   AssetTransactionType = "transfer"
	AssetTransactionTypeDispose    AssetTransactionType = "dispose"
	AssetTransactionTypeRevalue    AssetTransactionType = "revalue"
	AssetTransactionTypeAdjust     AssetTransactionType = "adjust"
)

type AssetTransactionStatus string

const (
	AssetTransactionStatusPending   AssetTransactionStatus = "pending"
	AssetTransactionStatusPosted    AssetTransactionStatus = "posted"
	AssetTransactionStatusRejected  AssetTransactionStatus = "rejected"
	AssetTransactionStatusCancelled AssetTransactionStatus = "cancelled"
)

type AssetTransaction struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	AssetID string `gorm:"type:uuid;not null;index" json:"asset_id"`
	Asset   *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`

	Type            AssetTransactionType   `gorm:"type:varchar(20);not null;index" json:"type"`
	TransactionDate time.Time              `gorm:"type:date;not null;index" json:"transaction_date"`
	Amount          float64                `gorm:"type:numeric(18,2);not null;default:0" json:"amount"`
	Description     string                 `gorm:"type:text" json:"description"`
	Status          AssetTransactionStatus `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`

	ReferenceType *string `gorm:"type:varchar(50)" json:"reference_type"`
	ReferenceID   *string `gorm:"type:uuid" json:"reference_id"`

	// Disposal-specific accounting metadata
	ProceedsAmount         float64 `gorm:"type:numeric(18,2);not null;default:0" json:"proceeds_amount"`
	BankAccountID          *string `gorm:"type:uuid;index" json:"bank_account_id,omitempty"`
	BookValueAtTransaction float64 `gorm:"type:numeric(18,2);not null;default:0" json:"book_value_at_transaction"`
	GainLossAmount         float64 `gorm:"type:numeric(18,2);not null;default:0" json:"gain_loss_amount"`
	GainLossAccountID      *string `gorm:"type:uuid;index" json:"gain_loss_account_id,omitempty"`

	CreatedBy *string   `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func (AssetTransaction) TableName() string {
	return "asset_transactions"
}

func (t *AssetTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
