package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/data/models"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesQuotationStatus represents the status of a sales quotation
type SalesQuotationStatus string

const (
	SalesQuotationStatusDraft     SalesQuotationStatus = "draft"
	SalesQuotationStatusSent      SalesQuotationStatus = "sent"
	SalesQuotationStatusApproved  SalesQuotationStatus = "approved"
	SalesQuotationStatusRejected  SalesQuotationStatus = "rejected"
	SalesQuotationStatusConverted SalesQuotationStatus = "converted"
)

// SalesQuotation represents a sales quotation document
type SalesQuotation struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	QuotationDate time.Time `gorm:"type:date;not null;index" json:"quotation_date"`
	ValidUntil    *time.Time `gorm:"type:date" json:"valid_until"`
	
	// Relations
	PaymentTermsID *string              `gorm:"type:uuid;index" json:"payment_terms_id"`
	PaymentTerms   *models.PaymentTerms `gorm:"foreignKey:PaymentTermsID" json:"payment_terms,omitempty"`
	
	SalesRepID     *string              `gorm:"type:uuid;index" json:"sales_rep_id"`
	SalesRep       *orgModels.Employee  `gorm:"foreignKey:SalesRepID" json:"sales_rep,omitempty"`
	
	BusinessUnitID *string              `gorm:"type:uuid;index" json:"business_unit_id"`
	BusinessUnit   *orgModels.BusinessUnit `gorm:"foreignKey:BusinessUnitID" json:"business_unit,omitempty"`
	
	BusinessTypeID *string              `gorm:"type:uuid;index" json:"business_type_id"`
	BusinessType   *orgModels.BusinessType `gorm:"foreignKey:BusinessTypeID" json:"business_type,omitempty"`
	
	// Customer reference (FK to master data customer)
	CustomerID      *string                    `gorm:"type:uuid;index" json:"customer_id"`
	Customer        *customerModels.Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	CustomerContactID *string                  `gorm:"type:uuid;index" json:"customer_contact_id"`

	// Customer snapshot (stored at quotation creation for historical record)
	CustomerName    string `gorm:"type:varchar(255)" json:"customer_name"`
	CustomerContact string `gorm:"type:varchar(255)" json:"customer_contact"`
	CustomerPhone   string `gorm:"type:varchar(50)" json:"customer_phone"`
	CustomerEmail   string `gorm:"type:varchar(255)" json:"customer_email"`
	
	// Financial calculations
	Subtotal       float64   `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	DiscountAmount float64   `gorm:"type:decimal(15,2);default:0" json:"discount_amount"`
	TaxRate        float64   `gorm:"type:decimal(5,2);default:11.00" json:"tax_rate"` // Default 11% PPN
	TaxAmount      float64   `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	DeliveryCost   float64   `gorm:"type:decimal(15,2);default:0" json:"delivery_cost"`
	OtherCost      float64   `gorm:"type:decimal(15,2);default:0" json:"other_cost"`
	TotalAmount    float64   `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	
	// Status and workflow
	Status         SalesQuotationStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	Notes          string               `gorm:"type:text" json:"notes"`
	
	// Audit fields
	CreatedBy      *string    `gorm:"type:uuid" json:"created_by"`
	ApprovedBy     *string    `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt     *time.Time `json:"approved_at"`
	RejectedBy     *string    `gorm:"type:uuid" json:"rejected_by"`
	RejectedAt     *time.Time `json:"rejected_at"`
	RejectionReason *string   `gorm:"type:text" json:"rejection_reason"`
	
	// Source tracking (backlink from CRM Deal conversion)
	SourceDealID *string `gorm:"type:uuid;index" json:"source_deal_id"`

	// Conversion tracking
	ConvertedToSalesOrderID *string    `gorm:"type:uuid;index" json:"converted_to_sales_order_id"`
	ConvertedAt             *time.Time `json:"converted_at"`
	
	// Timestamps
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Items          []SalesQuotationItem `gorm:"foreignKey:SalesQuotationID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

// TableName specifies the table name for SalesQuotation
func (SalesQuotation) TableName() string {
	return "sales_quotations"
}

// BeforeCreate hook to generate UUID
func (sq *SalesQuotation) BeforeCreate(tx *gorm.DB) error {
	if sq.ID == "" {
		sq.ID = uuid.New().String()
	}
	return nil
}

// SalesQuotationItem represents an item in a sales quotation
type SalesQuotationItem struct {
	ID               string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID         string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SalesQuotationID string    `gorm:"type:uuid;not null;index" json:"sales_quotation_id"`
	SalesQuotation   *SalesQuotation `gorm:"foreignKey:SalesQuotationID" json:"sales_quotation,omitempty"`
	
	ProductID        string    `gorm:"type:uuid;not null;index" json:"product_id"`
	Product          *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	
	Quantity         float64   `gorm:"type:decimal(15,3);not null" json:"quantity"`
	Price            float64   `gorm:"type:decimal(15,2);not null" json:"price"`
	Discount         float64   `gorm:"type:decimal(15,2);default:0" json:"discount"` // Discount amount, not percentage
	Subtotal         float64   `gorm:"type:decimal(15,2);not null" json:"subtotal"`
	
	// Timestamps
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SalesQuotationItem
func (SalesQuotationItem) TableName() string {
	return "sales_quotation_items"
}

// BeforeCreate hook to generate UUID
func (sqi *SalesQuotationItem) BeforeCreate(tx *gorm.DB) error {
	if sqi.ID == "" {
		sqi.ID = uuid.New().String()
	}
	return nil
}

// CalculateSubtotal calculates the subtotal for the item
func (sqi *SalesQuotationItem) CalculateSubtotal() {
	sqi.Subtotal = (sqi.Price * sqi.Quantity) - sqi.Discount
}
