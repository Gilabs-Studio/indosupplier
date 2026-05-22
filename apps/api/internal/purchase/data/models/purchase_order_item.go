package models

import (
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseOrderItem struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PurchaseOrderID string        `gorm:"type:uuid;index;not null" json:"purchase_order_id"`
	PurchaseOrder   *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`

	ProductID string              `gorm:"type:uuid;index;not null" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	ProductCodeSnapshot string     `gorm:"type:varchar(50)" json:"product_code_snapshot,omitempty"`
	ProductNameSnapshot string     `gorm:"type:varchar(200)" json:"product_name_snapshot,omitempty"`

	Quantity float64 `gorm:"type:decimal(15,2);default:0" json:"quantity"`
	Price    float64 `gorm:"type:decimal(15,2);default:0" json:"price"`
	Discount float64 `gorm:"type:decimal(5,2);default:0" json:"discount"`
	Subtotal float64 `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	Notes    *string `gorm:"type:text" json:"notes"`
}

func (PurchaseOrderItem) TableName() string {
	return "purchase_order_items"
}

func (it *PurchaseOrderItem) BeforeCreate(tx *gorm.DB) error {
	if it.ID == "" {
		it.ID = uuid.New().String()
	}
	return nil
}
