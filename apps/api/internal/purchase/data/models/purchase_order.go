package models

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseOrderStatus string

const (
	PurchaseOrderStatusDraft     PurchaseOrderStatus = "DRAFT"
	PurchaseOrderStatusSubmitted PurchaseOrderStatus = "SUBMITTED"
	PurchaseOrderStatusApproved  PurchaseOrderStatus = "APPROVED"
	PurchaseOrderStatusRejected  PurchaseOrderStatus = "REJECTED"
	PurchaseOrderStatusClosed    PurchaseOrderStatus = "CLOSED"
)

type PurchaseOrder struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	SupplierID *string                 `gorm:"type:uuid;index" json:"supplier_id"`
	Supplier   *supplierModels.Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`

	SupplierCodeSnapshot string            `gorm:"type:varchar(50)" json:"supplier_code_snapshot,omitempty"`
	SupplierNameSnapshot string            `gorm:"type:varchar(200)" json:"supplier_name_snapshot,omitempty"`

	CompanyID *string               `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Company   *orgModels.Company    `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	FiscalYearID *string                 `gorm:"type:uuid;index" json:"fiscal_year_id,omitempty"`
	FiscalYear   *financeModels.FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`

	PaymentTermsID *string               `gorm:"type:uuid;index" json:"payment_terms_id"`
	PaymentTerms   *coreModels.PaymentTerms `gorm:"foreignKey:PaymentTermsID" json:"payment_terms,omitempty"`

	PaymentTermsNameSnapshot string        `gorm:"type:varchar(150)" json:"payment_terms_name_snapshot,omitempty"`
	PaymentTermsDaysSnapshot *int          `gorm:"type:int" json:"payment_terms_days_snapshot,omitempty"`

	BusinessUnitID *string               `gorm:"type:uuid;index" json:"business_unit_id"`
	BusinessUnit   *orgModels.BusinessUnit `gorm:"foreignKey:BusinessUnitID" json:"business_unit,omitempty"`

	BusinessUnitNameSnapshot string        `gorm:"type:varchar(150)" json:"business_unit_name_snapshot,omitempty"`
	CreatedBy string          `gorm:"type:uuid;index;not null" json:"created_by"`
	Creator   *userModels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	PurchaseRequisitionID *string                `gorm:"type:uuid;index" json:"purchase_requisitions_id"`
	PurchaseRequisition   *PurchaseRequisition     `gorm:"foreignKey:PurchaseRequisitionID" json:"purchase_requisition,omitempty"`

	SalesOrderID *string           `gorm:"type:uuid;index" json:"sales_order_id"`
	SalesOrder   *salesModels.SalesOrder `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`

	OrderDate string  `gorm:"type:varchar(20);index" json:"order_date"`
	DueDate   *string `gorm:"type:varchar(20);index" json:"due_date"`

	RevisionComment *string `gorm:"type:text" json:"revision_comment"`
	Notes           string  `gorm:"type:text" json:"notes"`

	Status PurchaseOrderStatus `gorm:"type:varchar(20);default:'DRAFT';index" json:"status"`

	SubmittedAt *time.Time `gorm:"index" json:"submitted_at"`
	ApprovedAt  *time.Time `gorm:"index" json:"approved_at"`
	ClosedAt    *time.Time `gorm:"index" json:"closed_at"`

	TaxRate      float64 `gorm:"type:decimal(5,2);default:0" json:"tax_rate"`
	TaxAmount    float64 `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	DeliveryCost float64 `gorm:"type:decimal(15,2);default:0" json:"delivery_cost"`
	OtherCost    float64 `gorm:"type:decimal(15,2);default:0" json:"other_cost"`
	SubTotal     float64 `gorm:"type:decimal(15,2);default:0" json:"sub_total"`
	TotalAmount  float64 `gorm:"type:decimal(15,2);default:0" json:"total_amount"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Items []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	GoodsReceipts    []GoodsReceipt    `gorm:"foreignKey:PurchaseOrderID" json:"goods_receipts,omitempty"`
	SupplierInvoices []SupplierInvoice `gorm:"foreignKey:PurchaseOrderID" json:"supplier_invoices,omitempty"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

func (po *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	if po.ID == "" {
		po.ID = uuid.New().String()
	}
	return nil
}
