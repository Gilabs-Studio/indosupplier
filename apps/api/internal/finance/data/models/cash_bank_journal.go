package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CashBankStatus string

const (
	CashBankStatusDraft  CashBankStatus = "draft"
	CashBankStatusPosted CashBankStatus = "posted"
)

type CashBankType string

const (
	CashBankTypeCashIn   CashBankType = "cash_in"
	CashBankTypeCashOut  CashBankType = "cash_out"
	CashBankTypeTransfer CashBankType = "transfer"
)

type CashBankJournal struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	TransactionDate time.Time    `gorm:"type:date;not null;index" json:"transaction_date"`
	Type            CashBankType `gorm:"type:varchar(20);not null;index" json:"type"`
	Description     string       `gorm:"type:text" json:"description"`

	BankAccountID               string `gorm:"type:uuid;not null;index" json:"bank_account_id"`
	BankAccountNameSnapshot     string `gorm:"type:varchar(150)" json:"bank_account_name_snapshot,omitempty"`
	BankAccountNumberSnapshot   string `gorm:"type:varchar(50)" json:"bank_account_number_snapshot,omitempty"`
	BankAccountHolderSnapshot   string `gorm:"type:varchar(150)" json:"bank_account_holder_snapshot,omitempty"`
	BankAccountCurrencySnapshot string `gorm:"type:varchar(10)" json:"bank_account_currency_snapshot,omitempty"`

	TotalAmount float64        `gorm:"type:numeric(18,2);not null" json:"total_amount"`
	Status      CashBankStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`

	JournalEntryID *string `gorm:"type:uuid;index" json:"journal_entry_id"`

	PostedAt *time.Time `json:"posted_at"`
	PostedBy *string    `gorm:"type:uuid" json:"posted_by"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Lines []CashBankJournalLine `gorm:"foreignKey:CashBankJournalID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`
}

func (CashBankJournal) TableName() string {
	return "cash_bank_journals"
}

func (cb *CashBankJournal) BeforeCreate(tx *gorm.DB) error {
	if cb.ID == "" {
		cb.ID = uuid.New().String()
	}
	return nil
}
