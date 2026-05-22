package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusDraft    PaymentStatus = "draft"
	PaymentStatusPosted   PaymentStatus = "posted"
	PaymentStatusReversed PaymentStatus = "reversed"
)

type Payment struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PaymentDate time.Time `gorm:"type:date;not null;index" json:"payment_date"`
	Description string    `gorm:"type:text" json:"description"`

	BankAccountID               string `gorm:"type:uuid;not null;index" json:"bank_account_id"`
	BankAccountNameSnapshot     string `gorm:"type:varchar(150)" json:"bank_account_name_snapshot,omitempty"`
	BankAccountNumberSnapshot   string `gorm:"type:varchar(50)" json:"bank_account_number_snapshot,omitempty"`
	BankAccountHolderSnapshot   string `gorm:"type:varchar(150)" json:"bank_account_holder_snapshot,omitempty"`
	BankAccountCurrencySnapshot string `gorm:"type:varchar(10)" json:"bank_account_currency_snapshot,omitempty"`

	TotalAmount float64       `gorm:"type:numeric(18,2);not null" json:"total_amount"`
	Status      PaymentStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`

	JournalEntryID *string `gorm:"type:uuid;index" json:"journal_entry_id"`

	ApprovedAt *time.Time `json:"approved_at"`
	ApprovedBy *string    `gorm:"type:uuid" json:"approved_by"`

	PostedAt *time.Time `json:"posted_at"`
	PostedBy *string    `gorm:"type:uuid" json:"posted_by"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Allocations []PaymentAllocation `gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE" json:"allocations,omitempty"`
}

func (Payment) TableName() string {
	return "payments"
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
