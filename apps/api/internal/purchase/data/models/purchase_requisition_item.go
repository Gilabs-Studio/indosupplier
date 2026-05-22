package models

import (
	"time"

	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PurchaseRequisitionItem represents an item inside a PR
type PurchaseRequisitionItem struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PurchaseRequisitionID string               `gorm:"type:uuid;not null;index" json:"purchase_requisition_id"`
	PurchaseRequisition   *PurchaseRequisition `gorm:"foreignKey:PurchaseRequisitionID" json:"purchase_requisition,omitempty"`

	ProductID string               `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	ProductCodeSnapshot string `gorm:"type:varchar(50)" json:"product_code_snapshot,omitempty"`
	ProductNameSnapshot string `gorm:"type:varchar(200)" json:"product_name_snapshot,omitempty"`

	Quantity      float64 `gorm:"type:decimal(15,3);not null" json:"quantity"`
	PurchasePrice float64 `gorm:"type:decimal(15,2);not null" json:"purchase_price"`
	Discount      float64 `gorm:"type:decimal(5,2);default:0" json:"discount"` // percentage 0..100
	Subtotal      float64 `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	Notes         *string `gorm:"type:text" json:"notes"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PurchaseRequisitionItem) TableName() string {
	return "purchase_requisition_items"
}

func (pri *PurchaseRequisitionItem) BeforeCreate(tx *gorm.DB) error {
	if pri.ID == "" {
		pri.ID = uuid.New().String()
	}
	return nil
}
