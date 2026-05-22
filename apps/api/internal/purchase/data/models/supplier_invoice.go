package models

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupplierInvoiceType string

type SupplierInvoiceStatus string

const (
	SupplierInvoiceTypeNormal      SupplierInvoiceType = "NORMAL"
	SupplierInvoiceTypeDownPayment SupplierInvoiceType = "DOWN_PAYMENT"
)

const (
	SupplierInvoiceStatusDraft          SupplierInvoiceStatus = "DRAFT"
	SupplierInvoiceStatusSubmitted      SupplierInvoiceStatus = "SUBMITTED"
	SupplierInvoiceStatusApproved       SupplierInvoiceStatus = "APPROVED"
	SupplierInvoiceStatusRejected       SupplierInvoiceStatus = "REJECTED"
	SupplierInvoiceStatusCancelled      SupplierInvoiceStatus = "CANCELLED"
	SupplierInvoiceStatusUnpaid         SupplierInvoiceStatus = "UNPAID"
	SupplierInvoiceStatusWaitingPayment SupplierInvoiceStatus = "WAITING_PAYMENT"
	SupplierInvoiceStatusPartial        SupplierInvoiceStatus = "PARTIAL"
	SupplierInvoiceStatusPaid           SupplierInvoiceStatus = "PAID"
	SupplierInvoiceStatusReversed       SupplierInvoiceStatus = "REVERSED"
)

type SupplierInvoice struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Type SupplierInvoiceType `gorm:"type:varchar(20);not null;index" json:"type"`

	PurchaseOrderID string         `gorm:"type:uuid;index;not null" json:"purchase_order_id"`
	PurchaseOrder   *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`

	CompanyID    *string                   `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Company      *orgModels.Company        `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	FiscalYearID *string                   `gorm:"type:uuid;index" json:"fiscal_year_id,omitempty"`
	FiscalYear   *financeModels.FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`

	SupplierID           string                   `gorm:"type:uuid;index;not null" json:"supplier_id"`
	Supplier             *supplierModels.Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	SupplierCodeSnapshot string                   `gorm:"type:varchar(50)" json:"supplier_code_snapshot,omitempty"`
	SupplierNameSnapshot string                   `gorm:"type:varchar(200)" json:"supplier_name_snapshot,omitempty"`

	PaymentTermsID           *string                  `gorm:"type:uuid;index" json:"payment_terms_id"`
	PaymentTerms             *coreModels.PaymentTerms `gorm:"foreignKey:PaymentTermsID" json:"payment_terms,omitempty"`
	PaymentTermsNameSnapshot string                   `gorm:"type:varchar(150)" json:"payment_terms_name_snapshot,omitempty"`
	PaymentTermsDaysSnapshot *int                     `gorm:"type:int" json:"payment_terms_days_snapshot,omitempty"`

	Code             string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	InvoiceNumber    string    `gorm:"type:varchar(100);index;not null" json:"invoice_number"`
	InvoiceDate      time.Time `gorm:"type:date;index;not null" json:"invoice_date"`
	DueDate          time.Time `gorm:"type:date;index;not null" json:"due_date"`
	TaxInvoiceNumber string    `gorm:"type:varchar(100);index" json:"tax_invoice_number,omitempty"`

	TaxRate           float64 `gorm:"type:decimal(5,2);default:0" json:"tax_rate"`
	TaxAmount         float64 `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	DeliveryCost      float64 `gorm:"type:decimal(15,2);default:0" json:"delivery_cost"`
	OtherCost         float64 `gorm:"type:decimal(15,2);default:0" json:"other_cost"`
	DownPaymentAmount float64 `gorm:"type:decimal(15,2);default:0" json:"down_payment_amount"`
	SubTotal          float64 `gorm:"type:decimal(15,2);default:0" json:"sub_total"`
	Amount            float64 `gorm:"type:decimal(15,2);default:0" json:"amount"`
	PaidAmount        float64 `gorm:"type:decimal(15,2);default:0" json:"paid_amount"`
	RemainingAmount   float64 `gorm:"type:decimal(15,2);default:0" json:"remaining_amount"`

	GoodsReceiptID *string       `gorm:"type:uuid;index" json:"goods_receipt_id,omitempty"`
	GoodsReceipt   *GoodsReceipt `gorm:"foreignKey:GoodsReceiptID" json:"goods_receipt,omitempty"`

	DownPaymentInvoiceID *string          `gorm:"type:uuid;index" json:"down_payment_invoice_id,omitempty"`
	DownPaymentInvoice   *SupplierInvoice `gorm:"foreignKey:DownPaymentInvoiceID" json:"down_payment_invoice,omitempty"`

	Status            SupplierInvoiceStatus `gorm:"type:varchar(20);default:'DRAFT';index" json:"status"`
	IsPosted          bool                  `gorm:"default:false;index" json:"is_posted"`
	JournalEntryID    *string               `gorm:"type:uuid;index" json:"journal_entry_id,omitempty"`
	GRIRVarianceTotal float64               `gorm:"-" json:"gr_ir_variance_total,omitempty"`
	PaymentAt         *time.Time            `gorm:"type:timestamp" json:"payment_at,omitempty"`
	SubmittedAt       *time.Time            `gorm:"type:timestamp" json:"submitted_at,omitempty"`
	ApprovedAt        *time.Time            `gorm:"type:timestamp" json:"approved_at,omitempty"`
	RejectedAt        *time.Time            `gorm:"type:timestamp" json:"rejected_at,omitempty"`
	CancelledAt       *time.Time            `gorm:"type:timestamp" json:"cancelled_at,omitempty"`
	Notes             *string               `gorm:"type:text" json:"notes,omitempty"`

	CreatedBy string           `gorm:"type:uuid;index;not null" json:"created_by"`
	Creator   *userModels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	Items           []SupplierInvoiceItem `gorm:"foreignKey:SupplierInvoiceID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	RegularInvoices []SupplierInvoice     `gorm:"foreignKey:DownPaymentInvoiceID;references:ID" json:"regular_invoices,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierInvoice) TableName() string {
	return "supplier_invoices"
}

func (si *SupplierInvoice) BeforeCreate(tx *gorm.DB) error {
	if si.ID == "" {
		si.ID = uuid.New().String()
	}
	return nil
}
