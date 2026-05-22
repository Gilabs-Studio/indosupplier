package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PosType represents the type of POS operation
type PosType string

const (
	PosTypeFnbLive   PosType = "FNB_LIVE"   // Restaurant with floor plan + kitchen display
	PosTypeFnbSimple PosType = "FNB_SIMPLE" // Kiosk/counter, no table management
	PosTypeGoods     PosType = "GOODS"      // Goods distributor / souvenir shop
)

// PosOrderStatus represents the lifecycle status of a POS order
type PosOrderStatus string

const (
	PosOrderStatusDraft         PosOrderStatus = "DRAFT"
	PosOrderStatusInProgress    PosOrderStatus = "IN_PROGRESS"
	PosOrderStatusReady         PosOrderStatus = "READY"
	PosOrderStatusPartialServed PosOrderStatus = "PARTIAL_SERVED"
	PosOrderStatusServed        PosOrderStatus = "SERVED"
	PosOrderStatusPaid          PosOrderStatus = "PAID"
	PosOrderStatusCompleted     PosOrderStatus = "COMPLETED"
	PosOrderStatusVoided        PosOrderStatus = "VOIDED"
)

// PosOrderType represents the service type of the order
type PosOrderType string

const (
	PosOrderTypeDineIn   PosOrderType = "DINE_IN"
	PosOrderTypeTakeaway PosOrderType = "TAKEAWAY"
	PosOrderTypeCounter  PosOrderType = "COUNTER"
)

// PosItemStatus represents the preparation status of a single order item
type PosItemStatus string

const (
	PosItemStatusPending   PosItemStatus = "PENDING"
	PosItemStatusPreparing PosItemStatus = "PREPARING"
	PosItemStatusDone      PosItemStatus = "DONE"
	PosItemStatusServed    PosItemStatus = "SERVED"
)

// PosOrder represents a customer order within a POS session
type PosOrder struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	OrderNumber       string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_number"`
	SessionID         *string        `gorm:"type:uuid;index" json:"session_id,omitempty"`
	OutletID          string         `gorm:"type:uuid;not null;index" json:"outlet_id"`
	PosType           PosType        `gorm:"type:varchar(20);not null;default:'GOODS'" json:"pos_type"`
	TableID           *string        `gorm:"type:varchar(100);index" json:"table_id"`
	TableLabel        *string        `gorm:"type:varchar(50)" json:"table_label"`
	OrderType         PosOrderType   `gorm:"type:varchar(20);not null;default:'COUNTER'" json:"order_type"`
	CustomerID        *string        `gorm:"type:uuid;index" json:"customer_id"`
	CustomerName      *string        `gorm:"type:varchar(255)" json:"customer_name"`
	CustomerPhone     *string        `gorm:"type:varchar(50)" json:"customer_phone"`
	CustomerEmail     *string        `gorm:"type:varchar(255)" json:"customer_email"`
	GuestCount        int            `gorm:"default:1" json:"guest_count"`
	Subtotal          float64        `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	DiscountAmount    float64        `gorm:"type:decimal(15,2);default:0" json:"discount_amount"`
	TaxAmount         float64        `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	ServiceCharge     float64        `gorm:"type:decimal(15,2);default:0" json:"service_charge"`
	TotalAmount       float64        `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	Status            PosOrderStatus `gorm:"type:varchar(20);not null;default:'DRAFT';index" json:"status"`
	Notes             *string        `gorm:"type:text" json:"notes"`
	VoidReason        *string        `gorm:"type:text" json:"void_reason"`
	SalesOrderID      *string        `gorm:"type:uuid;index" json:"sales_order_id"`
	CustomerInvoiceID *string        `gorm:"type:uuid;index" json:"customer_invoice_id"`
	// LoyaltyMemberID and LoyaltyRewardID link an applied loyalty reward to the order.
	LoyaltyMemberID *string `gorm:"type:uuid;index" json:"loyalty_member_id,omitempty"`
	LoyaltyRewardID *string `gorm:"type:varchar(100)" json:"loyalty_reward_id,omitempty"`
	// OrderSource distinguishes self-service customer orders from staff-created orders.
	OrderSource string `gorm:"type:varchar(20);not null;default:'STAFF'" json:"order_source"`
	CreatedBy   string `gorm:"type:uuid;not null;index" json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	Items []PosOrderItem `gorm:"foreignKey:PosOrderID" json:"items,omitempty"`
}

func (PosOrder) TableName() string {
	return "pos_orders"
}

func (o *PosOrder) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// PosOrderItem represents a single product line within a POS order
type PosOrderItem struct {
	ID         string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID   string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	PosOrderID string `gorm:"type:uuid;not null;index" json:"pos_order_id"`
	ProductID   string        `gorm:"type:uuid;not null;index" json:"product_id"`
	ProductName string        `gorm:"type:varchar(255);not null" json:"product_name"`
	ProductCode string        `gorm:"type:varchar(100)" json:"product_code"`
	Quantity    float64       `gorm:"type:decimal(15,3);not null;default:1" json:"quantity"`
	UnitPrice   float64       `gorm:"type:decimal(15,2);not null;default:0" json:"unit_price"`
	Discount    float64       `gorm:"type:decimal(15,2);default:0" json:"discount"`
	Subtotal    float64       `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	Notes       *string       `gorm:"type:text" json:"notes"`
	Status      PosItemStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	KitchenNote *string       `gorm:"type:text" json:"kitchen_note"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (PosOrderItem) TableName() string {
	return "pos_order_items"
}

func (i *PosOrderItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}
