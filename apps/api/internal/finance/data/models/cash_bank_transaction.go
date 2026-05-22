package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CashBankTransactionType string

const (
	CashBankTransactionTypePaymentIn  CashBankTransactionType = "payment_in"
	CashBankTransactionTypePaymentOut CashBankTransactionType = "payment_out"
)

type CashBankTransactionStatus string

const (
	CashBankTransactionStatusDraft    CashBankTransactionStatus = "draft"
	CashBankTransactionStatusPosted   CashBankTransactionStatus = "posted"
	CashBankTransactionStatusReversed CashBankTransactionStatus = "reversed"
)

// CashBankTransaction is the Phase 2 operational cash/bank transaction source.
type CashBankTransaction struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID     string `gorm:"type:uuid;not null;index" json:"company_id"`
	BankAccountID string `gorm:"type:uuid;not null;index" json:"bank_account_id"`

	ReferenceNumber string                  `gorm:"type:varchar(40);not null;index" json:"reference_number"`
	Type            CashBankTransactionType `gorm:"type:varchar(20);not null;index" json:"type"`
	Date            time.Time               `gorm:"type:date;not null;index" json:"date"`
	Amount          float64                 `gorm:"type:numeric(20,4);not null" json:"amount"`

	Reference       string  `gorm:"type:varchar(120)" json:"reference"`
	Description     string  `gorm:"type:text" json:"description"`
	ContraAccountID *string `gorm:"type:uuid;index" json:"contra_account_id"`

	JournalEntryID *string                   `gorm:"type:uuid;index" json:"journal_entry_id"`
	Status         CashBankTransactionStatus `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`

	ReverseOfID  *string `gorm:"type:uuid;index" json:"reverse_of_id"`
	ReversedByID *string `gorm:"type:uuid;index" json:"reversed_by_id"`

	AttachmentURL *string `gorm:"type:text" json:"attachment_url"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`
	UpdatedBy *string `gorm:"type:uuid" json:"updated_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CashBankTransaction) TableName() string {
	return "cash_bank_transactions"
}

func (t *CashBankTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
