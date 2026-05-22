package models

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseReturnStatus string

type PurchaseReturnAction string

const (
	PurchaseReturnStatusDraft     PurchaseReturnStatus = "DRAFT"
	PurchaseReturnStatusSubmitted PurchaseReturnStatus = "SUBMITTED"
	PurchaseReturnStatusApproved  PurchaseReturnStatus = "APPROVED"
	PurchaseReturnStatusRejected  PurchaseReturnStatus = "REJECTED"
)

const (
	PurchaseReturnActionSupplierCredit PurchaseReturnAction = "SUPPLIER_CREDIT"
	PurchaseReturnActionRefund         PurchaseReturnAction = "REFUND"
	PurchaseReturnActionReplacement    PurchaseReturnAction = "REPLACEMENT"
)

type PurchaseReturn struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Code           string               `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	CompanyID      *string              `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Company        *orgModels.Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	FiscalYearID   *string              `gorm:"type:uuid;index" json:"fiscal_year_id,omitempty"`
	FiscalYear     *financeModels.FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
	GoodsReceiptID string               `gorm:"type:uuid;index;not null" json:"goods_receipt_id"`
	PurchaseOrderID *string             `gorm:"type:uuid;index" json:"purchase_order_id,omitempty"`
	SupplierID     string               `gorm:"type:uuid;index;not null" json:"supplier_id"`
	WarehouseID    string               `gorm:"type:uuid;index;not null" json:"warehouse_id"`
	Reason         string               `gorm:"type:varchar(50);not null" json:"reason"`
	Action         PurchaseReturnAction `gorm:"type:varchar(20);not null" json:"action"`
	Status         PurchaseReturnStatus `gorm:"type:varchar(20);not null;default:'DRAFT';index" json:"status"`
	Notes          *string              `gorm:"type:text" json:"notes,omitempty"`

	TotalAmount       float64 `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	StockAdjustmentID *string `gorm:"type:uuid;index" json:"stock_adjustment_id,omitempty"`
	DebitNoteID       *string `gorm:"type:uuid;index" json:"debit_note_id,omitempty"`

	CreatedBy string `gorm:"type:uuid;index;not null" json:"created_by"`

	Items []PurchaseReturnItem `gorm:"foreignKey:PurchaseReturnID;constraint:OnDelete:CASCADE" json:"items,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PurchaseReturn) TableName() string {
	return "purchase_returns"
}

func (m *PurchaseReturn) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

type PurchaseReturnItem struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	PurchaseReturnID   string  `gorm:"type:uuid;index;not null" json:"purchase_return_id"`
	GoodsReceiptItemID *string `gorm:"type:uuid;index" json:"goods_receipt_item_id,omitempty"`
	ProductID          string  `gorm:"type:uuid;index;not null" json:"product_id"`
	UOMID              *string `gorm:"type:uuid;index" json:"uom_id,omitempty"`
	Condition          string  `gorm:"type:varchar(30);not null" json:"condition"`
	Notes              *string `gorm:"type:text" json:"notes,omitempty"`
	Quantity           float64 `gorm:"type:decimal(15,3);not null" json:"quantity"`
	UnitCost           float64 `gorm:"type:decimal(15,2);not null" json:"unit_cost"`
	Subtotal           float64 `gorm:"type:decimal(15,2);not null" json:"subtotal"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PurchaseReturnItem) TableName() string {
	return "purchase_return_items"
}

func (m *PurchaseReturnItem) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
