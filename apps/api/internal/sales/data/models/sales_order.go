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

// SalesOrderStatus represents the status of a sales order
type SalesOrderStatus string

const (
	SalesOrderStatusDraft     SalesOrderStatus = "draft"
	SalesOrderStatusSubmitted SalesOrderStatus = "submitted"
	SalesOrderStatusApproved  SalesOrderStatus = "approved"
	SalesOrderStatusClosed    SalesOrderStatus = "closed"
	SalesOrderStatusRejected  SalesOrderStatus = "rejected"
	SalesOrderStatusCancelled SalesOrderStatus = "cancelled"
)

// SalesOrder represents a sales order document
type SalesOrder struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	OrderDate time.Time `gorm:"type:date;not null;index" json:"order_date"`

	// Relations
	SalesQuotationID *string         `gorm:"type:uuid;index" json:"sales_quotation_id"`
	SalesQuotation   *SalesQuotation `gorm:"foreignKey:SalesQuotationID" json:"sales_quotation,omitempty"`

	PaymentTermsID *string              `gorm:"type:uuid;index" json:"payment_terms_id"`
	PaymentTerms   *models.PaymentTerms `gorm:"foreignKey:PaymentTermsID" json:"payment_terms,omitempty"`

	SalesRepID *string             `gorm:"type:uuid;index" json:"sales_rep_id"`
	SalesRep   *orgModels.Employee `gorm:"foreignKey:SalesRepID" json:"sales_rep,omitempty"`

	BusinessUnitID *string                 `gorm:"type:uuid;index" json:"business_unit_id"`
	BusinessUnit   *orgModels.BusinessUnit `gorm:"foreignKey:BusinessUnitID" json:"business_unit,omitempty"`

	BusinessTypeID *string                 `gorm:"type:uuid;index" json:"business_type_id"`
	BusinessType   *orgModels.BusinessType `gorm:"foreignKey:BusinessTypeID" json:"business_type,omitempty"`

	DeliveryAreaID *string         `gorm:"type:uuid;index" json:"delivery_area_id"`
	DeliveryArea   *orgModels.Area `gorm:"foreignKey:DeliveryAreaID" json:"delivery_area,omitempty"`

	// Customer reference (FK to master data customer)
	CustomerID        *string                  `gorm:"type:uuid;index" json:"customer_id"`
	Customer          *customerModels.Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	CustomerContactID *string                  `gorm:"type:uuid;index" json:"customer_contact_id"`

	// Customer snapshot (stored at order creation for historical record)
	CustomerName    string `gorm:"type:varchar(255)" json:"customer_name"`
	CustomerContact string `gorm:"type:varchar(255)" json:"customer_contact"`
	CustomerPhone   string `gorm:"type:varchar(50)" json:"customer_phone"`
	CustomerEmail   string `gorm:"type:varchar(255)" json:"customer_email"`

	// Financial calculations
	Subtotal       float64 `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	DiscountAmount float64 `gorm:"type:decimal(15,2);default:0" json:"discount_amount"`
	TaxRate        float64 `gorm:"type:decimal(5,2);default:11.00" json:"tax_rate"` // Default 11% PPN
	TaxAmount      float64 `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	DeliveryCost   float64 `gorm:"type:decimal(15,2);default:0" json:"delivery_cost"`
	OtherCost      float64 `gorm:"type:decimal(15,2);default:0" json:"other_cost"`
	TotalAmount    float64 `gorm:"type:decimal(15,2);default:0" json:"total_amount"`

	// Stock reservation
	ReservedStock bool `gorm:"default:false;index" json:"reserved_stock"`

	// Status and workflow
	Status SalesOrderStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	Notes  string           `gorm:"type:text" json:"notes"`

	// Source tracking — identifies orders created from other modules (e.g. POS)
	SourceType       string  `gorm:"type:varchar(20);default:'';index" json:"source_type"`
	SourcePOSOrderID *string `gorm:"type:uuid;index" json:"source_pos_order_id"`

	// Audit fields
	CreatedBy          *string    `gorm:"type:uuid" json:"created_by"`
	ConfirmedBy        *string    `gorm:"type:uuid" json:"confirmed_by"`
	ConfirmedAt        *time.Time `json:"confirmed_at"`
	CancelledBy        *string    `gorm:"type:uuid" json:"cancelled_by"`
	CancelledAt        *time.Time `json:"cancelled_at"`
	CancellationReason *string    `gorm:"type:text" json:"cancellation_reason"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Items            []SalesOrderItem  `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	DeliveryOrders   []DeliveryOrder   `gorm:"foreignKey:SalesOrderID" json:"delivery_orders,omitempty"`
	CustomerInvoices []CustomerInvoice `gorm:"foreignKey:SalesOrderID" json:"customer_invoices,omitempty"`
}

// TableName specifies the table name for SalesOrder
func (SalesOrder) TableName() string {
	return "sales_orders"
}

// BeforeCreate hook to generate UUID
func (so *SalesOrder) BeforeCreate(tx *gorm.DB) error {
	if so.ID == "" {
		so.ID = uuid.New().String()
	}
	return nil
}

// SalesOrderItem represents an item in a sales order
type SalesOrderItem struct {
	ID           string      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     string      `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SalesOrderID string      `gorm:"type:uuid;not null;index" json:"sales_order_id"`
	SalesOrder   *SalesOrder `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`

	ProductID string                 `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	Quantity float64 `gorm:"type:decimal(15,3);not null" json:"quantity"`
	Price    float64 `gorm:"type:decimal(15,2);not null" json:"price"`
	Discount float64 `gorm:"type:decimal(15,2);default:0" json:"discount"` // Discount amount, not percentage
	Subtotal float64 `gorm:"type:decimal(15,2);not null" json:"subtotal"`

	// Snapshot fields
	ProductCode string `gorm:"type:varchar(50)" json:"product_code"`
	ProductName string `gorm:"type:varchar(255)" json:"product_name"`

	// Stock reservation tracking
	ReservedQuantity  float64 `gorm:"type:decimal(15,3);default:0" json:"reserved_quantity"`
	DeliveredQuantity float64 `gorm:"type:decimal(15,3);default:0" json:"delivered_quantity"`
	InvoicedQuantity  float64 `gorm:"type:decimal(15,3);default:0" json:"invoiced_quantity"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SalesOrderItem
func (SalesOrderItem) TableName() string {
	return "sales_order_items"
}

// BeforeCreate hook to generate UUID
func (soi *SalesOrderItem) BeforeCreate(tx *gorm.DB) error {
	if soi.ID == "" {
		soi.ID = uuid.New().String()
	}
	return nil
}

// CalculateSubtotal calculates the subtotal for the item
func (soi *SalesOrderItem) CalculateSubtotal() {
	soi.Subtotal = (soi.Price * soi.Quantity) - soi.Discount
}
