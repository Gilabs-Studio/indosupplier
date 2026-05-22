package models

import (
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GoodsReceiptItem struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	GoodsReceiptID string       `gorm:"type:uuid;index;not null" json:"goods_receipt_id"`
	GoodsReceipt   *GoodsReceipt `gorm:"foreignKey:GoodsReceiptID" json:"goods_receipt,omitempty"`

	PurchaseOrderItemID string            `gorm:"type:uuid;index;not null" json:"purchase_order_item_id"`
	PurchaseOrderItem   *PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderItemID" json:"purchase_order_item,omitempty"`

	ProductID string               `gorm:"type:uuid;index;not null" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductCodeSnapshot string     `gorm:"type:varchar(50)" json:"product_code_snapshot,omitempty"`
	ProductNameSnapshot string     `gorm:"type:varchar(200)" json:"product_name_snapshot,omitempty"`

	QuantityReceived float64 `gorm:"type:decimal(15,2);default:0" json:"quantity_received"`
	Notes            *string `gorm:"type:text" json:"notes,omitempty"`
}

func (GoodsReceiptItem) TableName() string {
	return "goods_receipt_items"
}

func (it *GoodsReceiptItem) BeforeCreate(tx *gorm.DB) error {
	if it.ID == "" {
		it.ID = uuid.New().String()
	}
	return nil
}
