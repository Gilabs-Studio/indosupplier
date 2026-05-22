package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"

	inventoryModels "github.com/gilabs/gims/api/internal/inventory/data/models"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
)

// DeliveryOrderStatus represents the status of a delivery order
type DeliveryOrderStatus string

const (
	DeliveryOrderStatusDraft     DeliveryOrderStatus = "draft"
	DeliveryOrderStatusSent      DeliveryOrderStatus = "sent"
	DeliveryOrderStatusApproved  DeliveryOrderStatus = "approved"
	DeliveryOrderStatusRejected  DeliveryOrderStatus = "rejected"
	DeliveryOrderStatusPrepared  DeliveryOrderStatus = "prepared"
	DeliveryOrderStatusShipped   DeliveryOrderStatus = "shipped"
	DeliveryOrderStatusDelivered DeliveryOrderStatus = "delivered"
	DeliveryOrderStatusCancelled DeliveryOrderStatus = "cancelled"
)

// DeliveryOrder represents a delivery order document
type DeliveryOrder struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	DeliveryDate time.Time `gorm:"type:date;not null;index" json:"delivery_date"`

	// Relations
	SalesOrderID string      `gorm:"type:uuid;not null;index" json:"sales_order_id"`
	SalesOrder   *SalesOrder `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`

	WarehouseID *string                    `gorm:"type:uuid;index" json:"warehouse_id"`
	Warehouse   *warehouseModels.Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`

	DeliveredByID *string             `gorm:"type:uuid;index" json:"delivered_by_id"`
	DeliveredBy   *orgModels.Employee `gorm:"foreignKey:DeliveredByID" json:"delivered_by,omitempty"`

	CourierAgencyID *string               `gorm:"type:uuid;index" json:"courier_agency_id"`
	CourierAgency   *models.CourierAgency `gorm:"foreignKey:CourierAgencyID" json:"courier_agency,omitempty"`

	TrackingNumber string `gorm:"type:varchar(100);index" json:"tracking_number"`

	// Delivery information
	ReceiverName      string `gorm:"type:varchar(100)" json:"receiver_name"`
	ReceiverPhone     string `gorm:"type:varchar(20)" json:"receiver_phone"`
	DeliveryAddress   string `gorm:"type:text" json:"delivery_address"`
	ReceiverSignature string `gorm:"type:text" json:"receiver_signature"` // Base64 encoded signature image

	// Status and workflow
	Status DeliveryOrderStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	Notes  string              `gorm:"type:text" json:"notes"`
	IsPosted bool              `gorm:"default:false;index" json:"is_posted"`

	// Partial delivery tracking
	IsPartialDelivery bool `gorm:"default:false" json:"is_partial_delivery"`

	// Audit fields
	CreatedBy          *string    `gorm:"type:uuid" json:"created_by"`
	ShippedBy          *string    `gorm:"type:uuid" json:"shipped_by"`
	ShippedAt          *time.Time `json:"shipped_at"`
	DeliveredAt        *time.Time `json:"delivered_at"`
	CancelledBy        *string    `gorm:"type:uuid" json:"cancelled_by"`
	CancelledAt        *time.Time `json:"cancelled_at"`
	CancellationReason *string    `gorm:"type:text" json:"cancellation_reason"`

	JournalEntryID *string `gorm:"type:uuid;index" json:"journal_entry_id"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Items []DeliveryOrderItem `gorm:"foreignKey:DeliveryOrderID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

// TableName specifies the table name for DeliveryOrder
func (DeliveryOrder) TableName() string {
	return "delivery_orders"
}

// BeforeCreate hook to generate UUID
func (do *DeliveryOrder) BeforeCreate(tx *gorm.DB) error {
	if do.ID == "" {
		do.ID = uuid.New().String()
	}
	return nil
}

// DeliveryOrderItem represents an item in a delivery order
type DeliveryOrderItem struct {
	ID              string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID        string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	DeliveryOrderID string         `gorm:"type:uuid;not null;index" json:"delivery_order_id"`
	DeliveryOrder   *DeliveryOrder `gorm:"foreignKey:DeliveryOrderID" json:"delivery_order,omitempty"`

	WarehouseID *string           `gorm:"type:uuid;index" json:"warehouse_id,omitempty"`
	Warehouse   *warehouseModels.Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`

	SalesOrderItemID *string         `gorm:"type:uuid;index" json:"sales_order_item_id"`
	SalesOrderItem   *SalesOrderItem `gorm:"foreignKey:SalesOrderItemID" json:"sales_order_item,omitempty"`

	ProductID string                 `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	// Batch selection (FIFO/FEFO)
	InventoryBatchID *string                         `gorm:"type:uuid;index" json:"inventory_batch_id"`
	InventoryBatch   *inventoryModels.InventoryBatch `gorm:"foreignKey:InventoryBatchID" json:"inventory_batch,omitempty"`

	Quantity        float64 `gorm:"type:decimal(15,3);not null" json:"quantity"`
	Price           float64 `gorm:"type:decimal(15,2);not null" json:"price"`
	Subtotal        float64 `gorm:"type:decimal(15,2);not null" json:"subtotal"`
	AvgCostSnapshot float64 `gorm:"type:decimal(18,6);default:0" json:"avg_cost_snapshot"`
	COGSAmount      float64 `gorm:"type:decimal(18,2);default:0" json:"cogs_amount"`

	// Equipment-specific fields (for installation/function test)
	IsEquipment        bool       `gorm:"default:false" json:"is_equipment"`
	InstallationStatus *string    `gorm:"type:varchar(50)" json:"installation_status"`  // pending, completed, failed
	FunctionTestStatus *string    `gorm:"type:varchar(50)" json:"function_test_status"` // pending, passed, failed
	InstallationDate   *time.Time `json:"installation_date"`
	FunctionTestDate   *time.Time `json:"function_test_date"`
	InstallationNotes  string     `gorm:"type:text" json:"installation_notes"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for DeliveryOrderItem
func (DeliveryOrderItem) TableName() string {
	return "delivery_order_items"
}

// BeforeCreate hook to generate UUID
func (doi *DeliveryOrderItem) BeforeCreate(tx *gorm.DB) error {
	if doi.ID == "" {
		doi.ID = uuid.New().String()
	}
	return nil
}

// CalculateSubtotal calculates the subtotal for the item
func (doi *DeliveryOrderItem) CalculateSubtotal() {
	doi.Subtotal = doi.Price * doi.Quantity
}
