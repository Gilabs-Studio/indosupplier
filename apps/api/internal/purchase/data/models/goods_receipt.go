package models

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GoodsReceiptStatus string

const (
	GoodsReceiptStatusDraft     GoodsReceiptStatus = "DRAFT"
	GoodsReceiptStatusSubmitted GoodsReceiptStatus = "SUBMITTED"
	GoodsReceiptStatusApproved  GoodsReceiptStatus = "APPROVED"
	GoodsReceiptStatusPartial   GoodsReceiptStatus = "PARTIAL"
	GoodsReceiptStatusClosed    GoodsReceiptStatus = "CLOSED"
	GoodsReceiptStatusRejected  GoodsReceiptStatus = "REJECTED"
	// Kept for backward compatibility with existing data.
	GoodsReceiptStatusConfirmed GoodsReceiptStatus = "CONFIRMED"
)

type GoodsReceipt struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	PurchaseOrderID string                     `gorm:"type:uuid;index;not null" json:"purchase_order_id"`
	PurchaseOrder   *PurchaseOrder             `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`

	CompanyID *string              `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Company   *orgModels.Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	FiscalYearID *string                 `gorm:"type:uuid;index" json:"fiscal_year_id,omitempty"`
	FiscalYear   *financeModels.FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
	WarehouseID     *string                    `gorm:"type:uuid;index" json:"warehouse_id,omitempty"`
	Warehouse       *warehouseModels.Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`

	SupplierID           string                   `gorm:"type:uuid;index;not null" json:"supplier_id"`
	Supplier             *supplierModels.Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	SupplierCodeSnapshot string                   `gorm:"type:varchar(50)" json:"supplier_code_snapshot,omitempty"`
	SupplierNameSnapshot string                   `gorm:"type:varchar(200)" json:"supplier_name_snapshot,omitempty"`

	ReceiptDate   *time.Time         `gorm:"index" json:"receipt_date,omitempty"`
	Notes         *string            `gorm:"type:text" json:"notes,omitempty"`
	ProofImageURL *string            `gorm:"type:varchar(500)" json:"proof_image_url,omitempty"`
	Status        GoodsReceiptStatus `gorm:"type:varchar(20);default:'DRAFT';index" json:"status"`

	CreatedBy string           `gorm:"type:uuid;index;not null" json:"created_by"`
	Creator   *userModels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	SubmittedAt *time.Time `gorm:"index" json:"submitted_at,omitempty"`
	ApprovedAt  *time.Time `gorm:"index" json:"approved_at,omitempty"`
	ClosedAt    *time.Time `gorm:"index" json:"closed_at,omitempty"`
	RejectedAt  *time.Time `gorm:"index" json:"rejected_at,omitempty"`

	ConvertedAt                  *time.Time `gorm:"index" json:"converted_at,omitempty"`
	ConvertedToSupplierInvoiceID *string    `gorm:"type:uuid;index" json:"converted_to_supplier_invoice_id,omitempty"`
	JournalEntryID               *string    `gorm:"type:uuid;index" json:"journal_entry_id,omitempty"`

	Items []GoodsReceiptItem `gorm:"foreignKey:GoodsReceiptID;constraint:OnDelete:CASCADE" json:"items,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (GoodsReceipt) TableName() string {
	return "goods_receipts"
}

func (gr *GoodsReceipt) BeforeCreate(tx *gorm.DB) error {
	if gr.ID == "" {
		gr.ID = uuid.New().String()
	}
	return nil
}
