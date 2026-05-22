package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/product/data/models"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InventoryBatch represents a batch of inventory for a specific product and warehouse
type InventoryBatch struct {
	ID                 string                     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID           string                     `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	ProductID          string                     `gorm:"type:uuid;not null;index" json:"product_id"`
	Product            *models.Product            `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WarehouseID        string                     `gorm:"type:uuid;not null;index" json:"warehouse_id"`
	Warehouse          *warehouseModels.Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	
	// Batch info
	BatchNumber        string                     `gorm:"type:varchar(100);index" json:"batch_number"`
	ExpiryDate         *time.Time                 `gorm:"index" json:"expiry_date"`
	ProductionDate     *time.Time                 `json:"production_date"`
	
	// Quantities
	InitialQuantity    float64                    `gorm:"type:decimal(15,3);default:0" json:"initial_quantity"`
	CurrentQuantity    float64                    `gorm:"type:decimal(15,3);default:0" json:"current_quantity"`
	ReservedQuantity   float64                    `gorm:"type:decimal(15,3);default:0" json:"reserved_quantity"`
	
	// Costing
	CostPrice          float64                    `gorm:"type:decimal(15,2);default:0" json:"cost_price"` // Individual batch cost
	
	// Status
	IsActive           bool                       `gorm:"default:true" json:"is_active"`
	
	// Timestamps
	CreatedAt          time.Time                  `json:"created_at"`
	UpdatedAt          time.Time                  `json:"updated_at"`
	DeletedAt          gorm.DeletedAt             `gorm:"index" json:"-"`
}

func (InventoryBatch) TableName() string {
	return "inventory_batches"
}

func (ib *InventoryBatch) BeforeCreate(tx *gorm.DB) error {
	if ib.ID == "" {
		ib.ID = uuid.New().String()
	}

	if ib.TenantID == "" {
		if tid, ok := tx.Statement.Context.Value("tenant_id").(string); ok {
			ib.TenantID = tid
		}
	}
	return nil
}

func (ib *InventoryBatch) AvailableQuantity() float64 {
	return ib.CurrentQuantity - ib.ReservedQuantity
}
