package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentTransactionStatus represents the state of a payment transaction.
type PaymentTransactionStatus string

const (
	PaymentStatusPending  PaymentTransactionStatus = "pending"
	PaymentStatusPaid     PaymentTransactionStatus = "paid"
	PaymentStatusFailed   PaymentTransactionStatus = "failed"
	PaymentStatusExpired  PaymentTransactionStatus = "expired"
	PaymentStatusCanceled PaymentTransactionStatus = "canceled"
)

// PaymentMethod represents the payment method used.
type PaymentMethod string

const (
	PaymentMethodCard        PaymentMethod = "card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodEWallet      PaymentMethod = "ewallet"
	PaymentMethodDANA         PaymentMethod = "dana"
	PaymentMethodOVO          PaymentMethod = "ovo"
	PaymentMethodGCash        PaymentMethod = "gcash"
	PaymentMethodCoupon       PaymentMethod = "coupon"
)

// PaymentProvider represents the payment gateway provider.
type PaymentProvider string

const (
	PaymentProviderXendit   PaymentProvider = "xendit"
	PaymentProviderMidtrans PaymentProvider = "midtrans"
	PaymentProviderInternal PaymentProvider = "internal"
)

// PaymentTransaction represents a payment transaction for a subscription or order.
type PaymentTransaction struct {
	ID                 string                   `gorm:"type:uuid;primaryKey"                 json:"id"`
	TenantID           string                   `gorm:"type:uuid;index"                      json:"tenant_id"`
	SubscriptionID     string                   `gorm:"type:uuid;index"                      json:"subscription_id,omitempty"`
	Provider           PaymentProvider          `gorm:"type:varchar(32);not null;default:'xendit'" json:"provider"`
	Status             PaymentTransactionStatus `gorm:"type:varchar(32);not null;default:'pending'" json:"status"`
	PaymentMethod      PaymentMethod            `gorm:"type:varchar(32)"                     json:"payment_method,omitempty"`
	AmountIDR          int64                    `gorm:"type:bigint;not null"                 json:"amount_idr"`
	ProviderInvoiceID  string                   `gorm:"type:varchar(255);index"              json:"provider_invoice_id,omitempty"` // Xendit or Midtrans invoice ID
	ProviderPaymentID  string                   `gorm:"type:varchar(255);index"              json:"provider_payment_id,omitempty"` // Xendit or Midtrans payment ID
	ReceiptURL         string                   `gorm:"type:text"                            json:"receipt_url,omitempty"`
	InvoiceURL         string                   `gorm:"type:text"                            json:"invoice_url,omitempty"`
	Description        string                   `gorm:"type:text"                            json:"description,omitempty"`
	PaidAt             *time.Time               `gorm:"type:timestamptz"                     json:"paid_at,omitempty"`
	ExpiresAt          *time.Time               `gorm:"type:timestamptz"                     json:"expires_at,omitempty"`
	Metadata           string                   `gorm:"type:jsonb"                           json:"metadata,omitempty"` // Additional provider-specific data
	Notes              string                   `gorm:"type:text"                            json:"notes,omitempty"`
	CreatedAt          time.Time                `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time                `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt           `gorm:"index"                                json:"deleted_at,omitempty"`
}

// TableName specifies the table name for PaymentTransaction.
func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

// BeforeCreate hook to generate UUID.
func (p *PaymentTransaction) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// IsPaid returns true if the payment is completed.
func (p *PaymentTransaction) IsPaid() bool {
	return p.Status == PaymentStatusPaid
}

// IsActive returns true if the payment is still pending or active.
func (p *PaymentTransaction) IsActive() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusPaid
}
