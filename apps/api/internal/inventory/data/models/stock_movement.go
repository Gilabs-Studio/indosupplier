package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/product/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StockMovementType string

const (
	MovementTypeIn       StockMovementType = "IN"       // GR, Returns
	MovementTypeOut      StockMovementType = "OUT"      // DO, Returns
	MovementTypeAdjust   StockMovementType = "ADJUST"   // Stock Opname
	MovementTypeTransfer StockMovementType = "TRANSFER" // Warehouse Transfer
)

// StockMovement represents the central ledger for all inventory changes
// This table is APPEND-ONLY. No updates or deletes allowed for data integrity.
type StockMovement struct {
	ID           string            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     string            `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Date         time.Time         `gorm:"not null;index;default:CURRENT_TIMESTAMP" json:"date"`
	MovementType StockMovementType `gorm:"column:movement_type;type:varchar(20);not null;index;default:'IN'" json:"type"`
	
	// Reference to source document (polymorphic-ish)
	RefType   string            `gorm:"type:varchar(50);not null;index;default:''" json:"ref_type"` // "GoodsReceipt", "DeliveryOrder", "StockOpname", "Transfer"
	RefID     string            `gorm:"type:uuid;not null;index;default:'00000000-0000-0000-0000-000000000000'" json:"ref_id"`          // ID of the source document
	RefNumber string            `gorm:"type:varchar(100);default:''" json:"ref_number"`             // Human readable ref number (e.g. "GR-2404-001")
	
	// Source/Dest info for UI display
	Source    string            `gorm:"type:varchar(255);default:''" json:"source"`                 // "Supplier A", "Customer B", "Warehouse A -> B"
	
	// Quantities
	QtyIn     float64           `gorm:"type:decimal(15,3);default:0" json:"qty_in"`
	QtyOut    float64           `gorm:"type:decimal(15,3);default:0" json:"qty_out"`
	Balance   float64           `gorm:"type:decimal(15,3);default:0" json:"balance"`     // Running balance after this movement
	
	// Financials
	Cost      float64           `gorm:"type:decimal(15,2);default:0" json:"cost"`        // Unit cost at time of movement
	
	// Relations
	ProductID   string                     `gorm:"type:uuid;not null;index;default:'00000000-0000-0000-0000-000000000000'" json:"product_id"`
	Product     *models.Product            `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WarehouseID string                     `gorm:"type:uuid;not null;index;default:'00000000-0000-0000-0000-000000000000'" json:"warehouse_id"`
	Warehouse   *warehouseModels.Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	
	// Internal Relations
	InventoryBatchID *string         `gorm:"type:uuid;index" json:"inventory_batch_id,omitempty"`
	InventoryBatch   *InventoryBatch `gorm:"foreignKey:InventoryBatchID" json:"inventory_batch,omitempty"`

	// Link to Journal Entry for financial traceability
	JournalEntryID *string `gorm:"type:uuid;index" json:"journal_entry_id,omitempty"`

	// Audit
	CreatedBy   *string          `gorm:"type:uuid;default:null" json:"created_by"`
	Creator     *userModels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

func (StockMovement) TableName() string {
	return "stock_movements"
}

func (sm *StockMovement) BeforeCreate(tx *gorm.DB) error {
	if sm.ID == "" {
		sm.ID = uuid.New().String()
	}

	if sm.TenantID == "" {
		if tid, ok := tx.Statement.Context.Value("tenant_id").(string); ok {
			sm.TenantID = tid
		}
	}
	return nil
}
