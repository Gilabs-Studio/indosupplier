package models

import (
	"time"

	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupplierInvoiceItem struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	SupplierInvoiceID string          `gorm:"type:uuid;index;not null" json:"supplier_invoice_id"`
	SupplierInvoice   *SupplierInvoice `gorm:"foreignKey:SupplierInvoiceID" json:"supplier_invoice,omitempty"`

	PurchaseOrderItemID *string `gorm:"type:uuid;index" json:"purchase_order_item_id"`

	ProductID string               `gorm:"type:uuid;index;not null" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductCodeSnapshot string     `gorm:"type:varchar(50)" json:"product_code_snapshot,omitempty"`
	ProductNameSnapshot string     `gorm:"type:varchar(200)" json:"product_name_snapshot,omitempty"`

	Quantity float64 `gorm:"type:decimal(15,2);default:0" json:"quantity"`
	Price    float64 `gorm:"type:decimal(15,2);default:0" json:"price"`
	Discount float64 `gorm:"type:decimal(5,2);default:0" json:"discount"`
	SubTotal float64 `gorm:"type:decimal(15,2);default:0" json:"sub_total"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierInvoiceItem) TableName() string {
	return "supplier_invoice_items"
}

func (it *SupplierInvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if it.ID == "" {
		it.ID = uuid.New().String()
	}
	return nil
}
