package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InventoryValuationMethod string

const (
	InventoryValuationMethodAverageCost            InventoryValuationMethod = "average_cost"
	InventoryValuationMethodFIFO                   InventoryValuationMethod = "fifo"
	InventoryValuationMethodSpecificIdentification InventoryValuationMethod = "specific_identification"
)

type InventorySettings struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID string `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`

	ValuationMethod InventoryValuationMethod `gorm:"type:varchar(40);not null;default:'average_cost'" json:"valuation_method"`
	IsLocked        bool                     `gorm:"not null;default:false" json:"is_locked"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InventorySettings) TableName() string {
	return "inventory_settings"
}

func (i *InventorySettings) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

type InventoryAverageCost struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	CompanyID string `gorm:"type:uuid;not null;uniqueIndex:idx_inventory_avg_cost_company_product,priority:1" json:"company_id"`
	ProductID string `gorm:"type:uuid;not null;uniqueIndex:idx_inventory_avg_cost_company_product,priority:2" json:"product_id"`

	AverageCost   float64   `gorm:"type:decimal(18,6);not null;default:0" json:"average_cost"`
	TotalQuantity float64   `gorm:"type:decimal(18,4);not null;default:0" json:"total_quantity"`
	TotalValue    float64   `gorm:"type:decimal(18,2);not null;default:0" json:"total_value"`
	LastUpdated   time.Time `gorm:"not null" json:"last_updated"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InventoryAverageCost) TableName() string {
	return "inventory_average_costs"
}

func (i *InventoryAverageCost) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}
