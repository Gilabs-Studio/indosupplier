package models

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesPaymentStatus string

type SalesPaymentMethod string

const (
	SalesPaymentStatusPending   SalesPaymentStatus = "PENDING"
	SalesPaymentStatusConfirmed SalesPaymentStatus = "CONFIRMED"
	SalesPaymentStatusReversed  SalesPaymentStatus = "REVERSED"
)

const (
	SalesPaymentMethodBank SalesPaymentMethod = "BANK"
	SalesPaymentMethodCash SalesPaymentMethod = "CASH"
)

type SalesPayment struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CustomerInvoiceID string           `gorm:"type:uuid;not null;index" json:"invoice_id"`
	CustomerInvoice   *CustomerInvoice `gorm:"foreignKey:CustomerInvoiceID" json:"invoice,omitempty"`

	// BankAccountID is always stored for operational tracking (cash drawer or bank account).
	// Journal COA resolution happens at confirm time and does not read this field directly.
	BankAccountID               string                  `gorm:"type:uuid;not null;index" json:"bank_account_id"`
	BankAccount                 *coreModels.BankAccount `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
	BankAccountNameSnapshot     string                  `gorm:"type:varchar(150)" json:"bank_account_name_snapshot,omitempty"`
	BankAccountNumberSnapshot   string                  `gorm:"type:varchar(50)" json:"bank_account_number_snapshot,omitempty"`
	BankAccountHolderSnapshot   string                  `gorm:"type:varchar(150)" json:"bank_account_holder_snapshot,omitempty"`
	BankAccountCurrencySnapshot string                  `gorm:"type:varchar(10)" json:"bank_account_currency_snapshot,omitempty"`

	PaymentDate string  `gorm:"type:varchar(20);not null;index" json:"payment_date"`
	Amount      float64 `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	TenderAmount float64 `gorm:"type:decimal(15,2);not null;default:0" json:"tender_amount"`
	ChangeAmount float64 `gorm:"type:decimal(15,2);not null;default:0" json:"change_amount"`

	Method SalesPaymentMethod `gorm:"type:varchar(20);not null;index" json:"method"`
	Status SalesPaymentStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`

	ReferenceNumber *string `gorm:"type:varchar(100);index" json:"reference_number,omitempty"`
	Notes           *string `gorm:"type:text" json:"notes,omitempty"`
	SnapshotCOAID   *string `gorm:"type:uuid;index" json:"snapshot_coa_id,omitempty"`
	ResolvedCOAID   *string `gorm:"type:uuid;index" json:"resolved_coa_id,omitempty"`

	CreatedBy string           `gorm:"type:uuid;index;not null" json:"created_by"`
	Creator   *userModels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalesPayment) TableName() string {
	return "sales_payments"
}

func (p *SalesPayment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
