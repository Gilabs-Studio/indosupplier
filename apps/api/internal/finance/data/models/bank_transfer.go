package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BankTransferStatus string

const (
	BankTransferStatusPending   BankTransferStatus = "pending"
	BankTransferStatusCompleted BankTransferStatus = "completed"
	BankTransferStatusCancelled BankTransferStatus = "cancelled"
)

type BankTransfer struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID string `gorm:"type:uuid;not null;index" json:"company_id"`

	TransferNumber string `gorm:"type:varchar(40);not null;index" json:"transfer_number"`

	FromBankAccountID string `gorm:"type:uuid;not null;index" json:"from_bank_account_id"`
	ToBankAccountID   string `gorm:"type:uuid;not null;index" json:"to_bank_account_id"`

	Amount      float64   `gorm:"type:numeric(20,4);not null" json:"amount"`
	Date        time.Time `gorm:"type:date;not null;index" json:"date"`
	Reference   string    `gorm:"type:varchar(120)" json:"reference"`
	Description string    `gorm:"type:text" json:"description"`

	Status         BankTransferStatus `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	JournalEntryID *string            `gorm:"type:uuid;index" json:"journal_entry_id,omitempty"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy *string `gorm:"type:uuid" json:"updated_by,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BankTransfer) TableName() string {
	return "bank_transfers"
}

func (t *BankTransfer) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
