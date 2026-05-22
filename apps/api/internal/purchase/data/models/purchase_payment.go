package models

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchasePaymentStatus string

type PurchasePaymentMethod string

const (
	PurchasePaymentStatusPending   PurchasePaymentStatus = "PENDING"
	PurchasePaymentStatusConfirmed PurchasePaymentStatus = "CONFIRMED"
	PurchasePaymentStatusReversed  PurchasePaymentStatus = "REVERSED"
)

const (
	PurchasePaymentMethodBank PurchasePaymentMethod = "BANK"
	PurchasePaymentMethodCash PurchasePaymentMethod = "CASH"
)

type PurchasePayment struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID string              `gorm:"type:uuid;index;not null" json:"company_id"`
	Company   *orgModels.Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	FiscalYearID *string                `gorm:"type:uuid;index" json:"fiscal_year_id,omitempty"`
	FiscalYear   *financeModels.FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`

	SupplierInvoiceID string           `gorm:"type:uuid;not null;index" json:"invoice_id"`
	SupplierInvoice   *SupplierInvoice `gorm:"foreignKey:SupplierInvoiceID" json:"invoice,omitempty"`

	BankAccountID               *string                 `gorm:"type:uuid;index" json:"bank_account_id,omitempty"`
	BankAccount                 *coreModels.BankAccount `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
	BankAccountNameSnapshot     string                  `gorm:"type:varchar(150)" json:"bank_account_name_snapshot,omitempty"`
	BankAccountNumberSnapshot   string                  `gorm:"type:varchar(50)" json:"bank_account_number_snapshot,omitempty"`
	BankAccountHolderSnapshot   string                  `gorm:"type:varchar(150)" json:"bank_account_holder_snapshot,omitempty"`
	BankAccountCurrencySnapshot string                  `gorm:"type:varchar(10)" json:"bank_account_currency_snapshot,omitempty"`

	PaymentDate time.Time `gorm:"type:date;not null;index" json:"payment_date"`
	Amount      float64   `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`

	Method PurchasePaymentMethod `gorm:"type:varchar(20);not null;index" json:"method"`
	Status PurchasePaymentStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`

	ReferenceNumber       *string `gorm:"type:varchar(100);index" json:"reference_number,omitempty"`
	Notes                 *string `gorm:"type:text" json:"notes,omitempty"`
	CashBankTransactionID *string `gorm:"type:uuid;index" json:"cash_bank_transaction_id,omitempty"`
	SnapshotCOAID         *string `gorm:"type:uuid;index" json:"snapshot_coa_id,omitempty"`
	ResolvedCOAID         *string `gorm:"type:uuid;index" json:"resolved_coa_id,omitempty"`

	CreatedBy string           `gorm:"type:uuid;index;not null" json:"created_by"`
	Creator   *userModels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PurchasePayment) TableName() string {
	return "purchase_payments"
}

func (p *PurchasePayment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
