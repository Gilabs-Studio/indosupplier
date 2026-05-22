package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaxInvoice struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	TaxInvoiceNumber string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"tax_invoice_number"`
	TaxInvoiceDate   time.Time `gorm:"type:date;not null;index" json:"tax_invoice_date"`

	CustomerInvoiceID *string `gorm:"type:uuid;index" json:"customer_invoice_id"`
	SupplierInvoiceID *string `gorm:"type:uuid;index" json:"supplier_invoice_id"`

	DPPAmount   float64 `gorm:"type:numeric(18,2);default:0" json:"dpp_amount"`
	VATAmount   float64 `gorm:"type:numeric(18,2);default:0" json:"vat_amount"`
	TotalAmount float64 `gorm:"type:numeric(18,2);default:0" json:"total_amount"`

	Notes string `gorm:"type:text" json:"notes"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TaxInvoice) TableName() string {
	return "tax_invoices"
}

func (t *TaxInvoice) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
