package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerInvoiceStatus represents the status of a customer invoice
type CustomerInvoiceStatus string

const (
	CustomerInvoiceStatusDraft          CustomerInvoiceStatus = "DRAFT"
	CustomerInvoiceStatusSubmitted      CustomerInvoiceStatus = "SUBMITTED"
	CustomerInvoiceStatusApproved       CustomerInvoiceStatus = "APPROVED"
	CustomerInvoiceStatusRejected       CustomerInvoiceStatus = "REJECTED"
	CustomerInvoiceStatusUnpaid         CustomerInvoiceStatus = "UNPAID"
	CustomerInvoiceStatusOverdue        CustomerInvoiceStatus = "OVERDUE"
	CustomerInvoiceStatusWaitingApproval CustomerInvoiceStatus = "WAITING_APPROVAL"
	CustomerInvoiceStatusWaitingPayment CustomerInvoiceStatus = "WAITING_PAYMENT"
	CustomerInvoiceStatusPartial        CustomerInvoiceStatus = "PARTIAL"
	CustomerInvoiceStatusPaid           CustomerInvoiceStatus = "PAID"
	CustomerInvoiceStatusCancelled      CustomerInvoiceStatus = "CANCELLED"
	CustomerInvoiceStatusReversed       CustomerInvoiceStatus = "REVERSED"
)

// CustomerInvoiceType represents the type of invoice
type CustomerInvoiceType string

const (
	CustomerInvoiceTypeRegular     CustomerInvoiceType = "regular"
	CustomerInvoiceTypeProforma    CustomerInvoiceType = "proforma"
	CustomerInvoiceTypeDownPayment CustomerInvoiceType = "down_payment"
)

// CustomerInvoice represents a customer invoice document
type CustomerInvoice struct {
	ID            string              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string              `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code          string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	InvoiceNumber *string             `gorm:"type:varchar(100);uniqueIndex" json:"invoice_number"`
	Type          CustomerInvoiceType `gorm:"type:varchar(20);default:'regular'" json:"type"`
	InvoiceDate   time.Time           `gorm:"type:date;not null;index" json:"invoice_date"`
	DueDate       *time.Time          `gorm:"type:date;index" json:"due_date"`

	// Relations
	SalesOrderID *string     `gorm:"type:uuid;index" json:"sales_order_id"`
	SalesOrder   *SalesOrder `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`

	// Optional link to Delivery Order (invoice-on-delivery pattern)
	DeliveryOrderID *string        `gorm:"type:uuid;index" json:"delivery_order_id"`
	DeliveryOrder   *DeliveryOrder `gorm:"foreignKey:DeliveryOrderID" json:"delivery_order,omitempty"`

	PaymentTermsID *string              `gorm:"type:uuid;index" json:"payment_terms_id"`
	PaymentTerms   *models.PaymentTerms `gorm:"foreignKey:PaymentTermsID" json:"payment_terms,omitempty"`

	// Financial calculations
	Subtotal          float64 `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	TaxRate           float64 `gorm:"type:decimal(5,2);default:11.00" json:"tax_rate"` // Default 11% PPN
	TaxAmount         float64 `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	DeliveryCost      float64 `gorm:"type:decimal(15,2);default:0" json:"delivery_cost"`
	OtherCost         float64 `gorm:"type:decimal(15,2);default:0" json:"other_cost"`
	DownPaymentAmount float64 `gorm:"type:decimal(15,2);default:0" json:"down_payment_amount"`
	Amount            float64 `gorm:"type:decimal(15,2);default:0" json:"amount"` // Total invoice amount

	PaidAmount      float64 `gorm:"type:decimal(15,2);default:0" json:"paid_amount"`
	RemainingAmount float64 `gorm:"type:decimal(15,2);default:0" json:"remaining_amount"`

	// Track linked DP invoice
	DownPaymentInvoiceID *string          `gorm:"type:uuid;index" json:"down_payment_invoice_id"`
	DownPaymentInvoice   *CustomerInvoice `gorm:"foreignKey:DownPaymentInvoiceID" json:"down_payment_invoice,omitempty"`

	// Status and workflow
	Status        CustomerInvoiceStatus `gorm:"type:varchar(20);default:'DRAFT';index" json:"status"`
	AttachmentURL *string               `gorm:"type:text" json:"attachment_url,omitempty"`
	Notes         string                `gorm:"type:text" json:"notes"`
	IsPosted      bool                  `gorm:"default:false;index" json:"is_posted"`

	// Accounting linkage
	JournalEntryID *string `gorm:"type:uuid;index" json:"journal_entry_id,omitempty"`

	// Payment timestamp
	PaymentAt *time.Time `json:"payment_at"`

	// Workflow timestamps
	SubmittedAt *time.Time `gorm:"type:timestamp" json:"submitted_at,omitempty"`
	ApprovedAt  *time.Time `gorm:"type:timestamp" json:"approved_at,omitempty"`
	RejectedAt  *time.Time `gorm:"type:timestamp" json:"rejected_at,omitempty"`

	// Tax Invoice relation (optional)
	TaxInvoiceID *string `gorm:"type:uuid;index" json:"tax_invoice_id"`

	// Audit fields
	CreatedBy   *string    `gorm:"type:uuid" json:"created_by"`
	CancelledBy *string    `gorm:"type:uuid" json:"cancelled_by"`
	CancelledAt *time.Time `json:"cancelled_at"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Items []CustomerInvoiceItem `gorm:"foreignKey:CustomerInvoiceID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

// TableName specifies the table name for CustomerInvoice
func (CustomerInvoice) TableName() string {
	return "customer_invoices"
}

// BeforeCreate hook to generate UUID
func (ci *CustomerInvoice) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == "" {
		ci.ID = uuid.New().String()
	}
	return nil
}

// CustomerInvoiceItem represents an item in a customer invoice
type CustomerInvoiceItem struct {
	ID                string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID          string           `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CustomerInvoiceID string           `gorm:"type:uuid;not null;index" json:"customer_invoice_id"`
	CustomerInvoice   *CustomerInvoice `gorm:"foreignKey:CustomerInvoiceID" json:"customer_invoice,omitempty"`

	ProductID string                 `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	// Link to SO/DO items for partial invoicing tracking
	SalesOrderItemID    *string            `gorm:"type:uuid;index" json:"sales_order_item_id"`
	SalesOrderItem      *SalesOrderItem    `gorm:"foreignKey:SalesOrderItemID" json:"sales_order_item,omitempty"`
	DeliveryOrderItemID *string            `gorm:"type:uuid;index" json:"delivery_order_item_id"`
	DeliveryOrderItem   *DeliveryOrderItem `gorm:"foreignKey:DeliveryOrderItemID" json:"delivery_order_item,omitempty"`

	Quantity  float64 `gorm:"type:decimal(15,3);not null" json:"quantity"`
	Price     float64 `gorm:"type:decimal(15,2);not null" json:"price"`
	Discount  float64 `gorm:"type:decimal(15,2);default:0" json:"discount"` // Discount amount
	Subtotal  float64 `gorm:"type:decimal(15,2);not null" json:"subtotal"`
	HPPAmount float64 `gorm:"type:decimal(15,2);default:0" json:"hpp_amount"` // Cost of goods sold per unit

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for CustomerInvoiceItem
func (CustomerInvoiceItem) TableName() string {
	return "customer_invoice_items"
}

// BeforeCreate hook to generate UUID
func (cii *CustomerInvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if cii.ID == "" {
		cii.ID = uuid.New().String()
	}
	return nil
}

// CalculateSubtotal calculates the subtotal for the item
func (cii *CustomerInvoiceItem) CalculateSubtotal() {
	cii.Subtotal = (cii.Price * cii.Quantity) - cii.Discount
}
