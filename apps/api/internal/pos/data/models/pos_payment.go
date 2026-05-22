package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POSPaymentMethod represents accepted payment methods at the POS
type POSPaymentMethod string

const (
	POSPaymentMethodCash     POSPaymentMethod = "CASH"
	POSPaymentMethodCard     POSPaymentMethod = "CARD"
	POSPaymentMethodQris     POSPaymentMethod = "QRIS"
	POSPaymentMethodTransfer POSPaymentMethod = "TRANSFER"
	// DIGITAL covers all gateway-routed payments (Xendit QRIS / VA / e-wallet)
	POSPaymentMethodDigital POSPaymentMethod = "DIGITAL"
)

// POSPaymentStatus represents the lifecycle status of a POS payment
type POSPaymentStatus string

const (
	POSPaymentStatusPending   POSPaymentStatus = "PENDING"
	POSPaymentStatusPaid      POSPaymentStatus = "PAID"
	POSPaymentStatusCancelled POSPaymentStatus = "CANCELLED"
	POSPaymentStatusFailed    POSPaymentStatus = "FAILED"
	POSPaymentStatusExpired   POSPaymentStatus = "EXPIRED"
	POSPaymentStatusRefunded  POSPaymentStatus = "REFUNDED"
)

// POSPayment records the monetary transaction for a completed POS order
type POSPayment struct {
	ID              string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID        string           `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	OrderID         string           `gorm:"type:uuid;not null;index" json:"order_id"`
	Method          POSPaymentMethod `gorm:"type:varchar(20);not null;index" json:"method"`
	Status          POSPaymentStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	Amount          float64          `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	TenderAmount    float64          `gorm:"type:decimal(15,2);default:0" json:"tender_amount"`
	ChangeAmount    float64          `gorm:"type:decimal(15,2);default:0" json:"change_amount"`
	ReferenceNumber *string          `gorm:"type:varchar(100)" json:"reference_number"`
	// Digital gateway fields (Xendit)
	ExternalOrderID *string    `gorm:"type:varchar(100);uniqueIndex" json:"external_order_id"`
	XenditInvoiceID *string    `gorm:"type:varchar(100)" json:"xendit_invoice_id"`
	TransactionID   *string    `gorm:"type:varchar(100)" json:"transaction_id"`
	PaymentType     *string    `gorm:"type:varchar(50)" json:"payment_type"`
	VaNumber        *string    `gorm:"type:varchar(100)" json:"va_number"`
	QrCode          *string    `gorm:"type:text" json:"qr_code"`
	PaymentURL      *string    `gorm:"type:text" json:"payment_url"`
	ExpiresAt       *time.Time `json:"expires_at"`
	PaidAt          *time.Time `json:"paid_at"`
	Notes           *string    `gorm:"type:text" json:"notes"`
	CreatedBy       string     `gorm:"type:uuid;not null;index" json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (POSPayment) TableName() string {
	return "pos_payments"
}

func (p *POSPayment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
