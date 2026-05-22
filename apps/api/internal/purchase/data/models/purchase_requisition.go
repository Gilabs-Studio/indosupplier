package models

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PurchaseRequisitionStatus represents PR workflow status
type PurchaseRequisitionStatus string

const (
	PurchaseRequisitionStatusDraft     PurchaseRequisitionStatus = "DRAFT"
	PurchaseRequisitionStatusSubmitted PurchaseRequisitionStatus = "SUBMITTED"
	PurchaseRequisitionStatusApproved  PurchaseRequisitionStatus = "APPROVED"
	PurchaseRequisitionStatusRejected  PurchaseRequisitionStatus = "REJECTED"
	PurchaseRequisitionStatusConverted PurchaseRequisitionStatus = "CONVERTED"
)

// PurchaseRequisition represents a purchase requisition document
// NOTE: request_date is stored as string (recommended ISO YYYY-MM-DD).
type PurchaseRequisition struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	SupplierID *string                 `gorm:"type:uuid;index" json:"supplier_id"`
	Supplier   *supplierModels.Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`

	SupplierCodeSnapshot string `gorm:"type:varchar(50)" json:"supplier_code_snapshot,omitempty"`
	SupplierNameSnapshot string `gorm:"type:varchar(200)" json:"supplier_name_snapshot,omitempty"`

	CompanyID *string             `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Company   *orgModels.Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	FiscalYearID *string                `gorm:"type:uuid;index" json:"fiscal_year_id,omitempty"`
	FiscalYear   *financeModels.FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`

	PaymentTermsID *string              `gorm:"type:uuid;index" json:"payment_terms_id"`
	PaymentTerms   *coreModels.PaymentTerms `gorm:"foreignKey:PaymentTermsID" json:"payment_terms,omitempty"`

	PaymentTermsNameSnapshot string `gorm:"type:varchar(150)" json:"payment_terms_name_snapshot,omitempty"`
	PaymentTermsDaysSnapshot *int   `gorm:"type:int" json:"payment_terms_days_snapshot,omitempty"`

	BusinessUnitID *string              `gorm:"type:uuid;index" json:"business_unit_id"`
	BusinessUnit   *orgModels.BusinessUnit `gorm:"foreignKey:BusinessUnitID" json:"business_unit,omitempty"`

	BusinessUnitNameSnapshot string `gorm:"type:varchar(150)" json:"business_unit_name_snapshot,omitempty"`

	EmployeeID *string           `gorm:"type:uuid;index" json:"employee_id"`
	Employee   *orgModels.Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

	RequestDate string `gorm:"type:varchar(20);index" json:"request_date"`
	Address     *string `gorm:"type:text" json:"address"`
	Notes       string  `gorm:"type:text" json:"notes"`

	Status PurchaseRequisitionStatus `gorm:"type:varchar(20);default:'DRAFT';index" json:"status"`

	// Workflow timestamps
	SubmittedAt *time.Time `gorm:"index"                json:"submitted_at"`
	ApprovedAt  *time.Time `gorm:"index"                json:"approved_at"`
	RejectedAt  *time.Time `gorm:"index"                json:"rejected_at"`
	ConvertedAt *time.Time `gorm:"index"                json:"converted_at"`

	// Reference to the Purchase Order created from conversion
	ConvertedToPurchaseOrderID *string `gorm:"type:uuid;index" json:"converted_to_purchase_order_id"`

	TaxRate      float64 `gorm:"type:decimal(5,2);default:0" json:"tax_rate"`
	TaxAmount    float64 `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	DeliveryCost float64 `gorm:"type:decimal(15,2);default:0" json:"delivery_cost"`
	OtherCost    float64 `gorm:"type:decimal(15,2);default:0" json:"other_cost"`
	Subtotal     float64 `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	TotalAmount  float64 `gorm:"type:decimal(15,2);default:0" json:"total_amount"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Items []PurchaseRequisitionItem `gorm:"foreignKey:PurchaseRequisitionID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (PurchaseRequisition) TableName() string {
	return "purchase_requisitions"
}

func (pr *PurchaseRequisition) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == "" {
		pr.ID = uuid.New().String()
	}
	return nil
}
