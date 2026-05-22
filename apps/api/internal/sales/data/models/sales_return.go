package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesReturnStatus string

type SalesReturnAction string

const (
	SalesReturnStatusDraft     SalesReturnStatus = "DRAFT"
	SalesReturnStatusSubmitted SalesReturnStatus = "SUBMITTED"
	SalesReturnStatusProcessed SalesReturnStatus = "PROCESSED"
	SalesReturnStatusRejected  SalesReturnStatus = "REJECTED"
)

const (
	SalesReturnActionRefund      SalesReturnAction = "REFUND"
	SalesReturnActionCreditNote  SalesReturnAction = "CREDIT_NOTE"
	SalesReturnActionReplacement SalesReturnAction = "REPLACEMENT"
)

type SalesReturn struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Code       string             `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	InvoiceID  *string            `gorm:"type:uuid;index" json:"invoice_id,omitempty"`
	DeliveryID *string            `gorm:"type:uuid;index" json:"delivery_id,omitempty"`
	WarehouseID string            `gorm:"type:uuid;index;not null" json:"warehouse_id"`
	CustomerID string             `gorm:"type:uuid;index;not null" json:"customer_id"`

	Reason string            `gorm:"type:varchar(50);not null" json:"reason"`
	Action SalesReturnAction `gorm:"type:varchar(20);not null" json:"action"`
	Status SalesReturnStatus `gorm:"type:varchar(20);not null;default:'DRAFT';index" json:"status"`
	Notes  *string           `gorm:"type:text" json:"notes,omitempty"`

	TotalAmount       float64 `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	StockAdjustmentID *string `gorm:"type:uuid;index" json:"stock_adjustment_id,omitempty"`
	CreditNoteID      *string `gorm:"type:uuid;index" json:"credit_note_id,omitempty"`

	CreatedBy string `gorm:"type:uuid;index;not null" json:"created_by"`

	Items []SalesReturnItem `gorm:"foreignKey:SalesReturnID;constraint:OnDelete:CASCADE" json:"items,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalesReturn) TableName() string {
	return "sales_returns"
}

func (m *SalesReturn) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

type SalesReturnItem struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	SalesReturnID string       `gorm:"type:uuid;index;not null" json:"sales_return_id"`
	InvoiceItemID *string      `gorm:"type:uuid;index" json:"invoice_item_id,omitempty"`
	ProductID     string       `gorm:"type:uuid;index;not null" json:"product_id"`
	UOMID         *string      `gorm:"type:uuid;index" json:"uom_id,omitempty"`
	Condition     string       `gorm:"type:varchar(30);not null" json:"condition"`
	Notes         *string      `gorm:"type:text" json:"notes,omitempty"`
	Quantity      float64      `gorm:"type:decimal(15,3);not null" json:"quantity"`
	UnitPrice     float64      `gorm:"type:decimal(15,2);not null" json:"unit_price"`
	Subtotal      float64      `gorm:"type:decimal(15,2);not null" json:"subtotal"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalesReturnItem) TableName() string {
	return "sales_return_items"
}

func (m *SalesReturnItem) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
