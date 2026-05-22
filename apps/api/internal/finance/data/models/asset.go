package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetStatus string

const (
	AssetStatusDraft                 AssetStatus = "draft"
	AssetStatusPendingApproval      AssetStatus = "pending_approval"
	AssetStatusPendingCapitalization AssetStatus = "pending_capitalization"
	AssetStatusActive                AssetStatus = "active"
	AssetStatusInUse                 AssetStatus = "in_use"
	AssetStatusUnderMaintenance      AssetStatus = "under_maintenance"
	AssetStatusIdle                  AssetStatus = "idle"
	AssetStatusDisposed              AssetStatus = "disposed"
	AssetStatusSold                  AssetStatus = "sold"
	AssetStatusWrittenOff            AssetStatus = "written_off"
	AssetStatusTransferred           AssetStatus = "transferred"
	AssetStatusTransferRequested     AssetStatus = "transfer_requested"
	AssetStatusInactive              AssetStatus = "inactive" // Deprecated, keep for compatibility
)

type AssetLifecycleStage string

const (
	AssetLifecycleDraft                 AssetLifecycleStage = "draft"
	AssetLifecyclePendingCapitalization AssetLifecycleStage = "pending_capitalization"
	AssetLifecyclePending               AssetLifecycleStage = "pending"
	AssetLifecycleActive                AssetLifecycleStage = "active"
	AssetLifecycleInUse                 AssetLifecycleStage = "in_use"
	AssetLifecycleIdle                  AssetLifecycleStage = "idle"
	AssetLifecycleUnderMaintenance      AssetLifecycleStage = "under_maintenance"
	AssetLifecycleDisposed              AssetLifecycleStage = "disposed"
	AssetLifecycleRetired               AssetLifecycleStage = "retired"
	AssetLifecycleSold                  AssetLifecycleStage = "sold"
	AssetLifecycleWrittenOff            AssetLifecycleStage = "written_off"
)

type Asset struct {
	// Core Identity
	ID          string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string  `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code        string  `gorm:"type:varchar(50);not null;index" json:"code"`
	Name        string  `gorm:"type:varchar(200);not null;index" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	AssetTypeID *string `gorm:"type:varchar(20);index" json:"asset_type_id,omitempty"`

	// NEW: Additional Identifiers
	SerialNumber *string `gorm:"type:varchar(100);index" json:"serial_number,omitempty"`
	Barcode      *string `gorm:"type:varchar(100);index" json:"barcode,omitempty"`
	QRCode       *string `gorm:"type:text" json:"qr_code,omitempty"`
	AssetTag     *string `gorm:"type:varchar(50)" json:"asset_tag,omitempty"`

	// Category & Location
	CategoryID string         `gorm:"type:uuid;not null;index" json:"category_id"`
	Category   *AssetCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`

	LocationID string         `gorm:"type:uuid;not null;index" json:"location_id"`
	Location   *AssetLocation `gorm:"foreignKey:LocationID" json:"location,omitempty"`

	// NEW: Organization Hierarchy
	CompanyID      *string       `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Company        *Company      `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	BusinessUnitID *string       `gorm:"type:uuid;index" json:"business_unit_id,omitempty"`
	BusinessUnit   *BusinessUnit `gorm:"foreignKey:BusinessUnitID" json:"business_unit,omitempty"`
	DepartmentID   *string       `gorm:"type:uuid;index" json:"department_id,omitempty"`
	Department     *Department   `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`

	// NEW: Assignment
	AssignedToEmployeeID *string    `gorm:"type:uuid;index" json:"assigned_to_employee_id,omitempty"`
	AssignedToEmployee   *Employee  `gorm:"foreignKey:AssignedToEmployeeID" json:"assigned_to_employee,omitempty"`
	AssignmentDate       *time.Time `gorm:"type:date" json:"assignment_date,omitempty"`

	// Acquisition Information
	AcquisitionDate time.Time `gorm:"type:date;not null;index" json:"acquisition_date"`
	AcquisitionCost float64   `gorm:"type:numeric(18,2);not null" json:"acquisition_cost"`
	SalvageValue    float64   `gorm:"type:numeric(18,2);default:0" json:"salvage_value"`

	// NEW: Acquisition Details
	SupplierID        *string          `gorm:"type:uuid;index" json:"supplier_id,omitempty"`
	Supplier          *Contact         `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	PurchaseOrderID   *string          `gorm:"type:uuid" json:"purchase_order_id,omitempty"`
	PurchaseOrder     *PurchaseOrder   `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
	SupplierInvoiceID *string          `gorm:"type:uuid" json:"supplier_invoice_id,omitempty"`
	SupplierInvoice   *SupplierInvoice `gorm:"foreignKey:SupplierInvoiceID" json:"supplier_invoice,omitempty"`
	CustodianUserID   *string          `gorm:"type:uuid;index" json:"custodian_user_id,omitempty"`
	CustodianUser     *User            `gorm:"foreignKey:CustodianUserID" json:"custodian_user,omitempty"`

	// NEW: Cost Breakdown
	ShippingCost     float64 `gorm:"type:numeric(18,2);default:0" json:"shipping_cost"`
	InstallationCost float64 `gorm:"type:numeric(18,2);default:0" json:"installation_cost"`
	TaxAmount        float64 `gorm:"type:numeric(18,2);default:0" json:"tax_amount"`
	OtherCosts       float64 `gorm:"type:numeric(18,2);default:0" json:"other_costs"`

	// Depreciation
	AccumulatedDepreciation float64 `gorm:"type:numeric(18,2);default:0" json:"accumulated_depreciation"`
	BookValue               float64 `gorm:"type:numeric(18,2);default:0" json:"book_value"`

	// NEW: Depreciation Configuration (override category defaults)
	DepreciationMethod    *string    `gorm:"type:varchar(10)" json:"depreciation_method,omitempty"` // SL, DB, SYD, UOP, NONE
	UsefulLifeMonths      *int       `gorm:"type:integer" json:"useful_life_months,omitempty"`
	DepreciationStartDate *time.Time `gorm:"type:date" json:"depreciation_start_date,omitempty"`

	// Status & Lifecycle
	Status         AssetStatus         `gorm:"type:varchar(20);default:'active';index" json:"status"`
	LifecycleStage AssetLifecycleStage `gorm:"type:varchar(30);default:'draft';index" json:"lifecycle_stage"`
	DisposedAt     *time.Time          `json:"disposed_at,omitempty"`

	// NEW: Lifecycle Flags
	IsCapitalized      bool `gorm:"type:boolean;default:false;index" json:"is_capitalized"`
	IsDepreciable      bool `gorm:"type:boolean;default:true" json:"is_depreciable"`
	IsFullyDepreciated bool `gorm:"type:boolean;default:false;index" json:"is_fully_depreciated"`

	// NEW: Parent/Child Relationship
	ParentAssetID *string `gorm:"type:uuid;index" json:"parent_asset_id,omitempty"`
	ParentAsset   *Asset  `gorm:"foreignKey:ParentAssetID" json:"parent_asset,omitempty"`
	IsParent      bool    `gorm:"type:boolean;default:false" json:"is_parent"`
	ChildAssets   []Asset `gorm:"foreignKey:ParentAssetID" json:"child_assets,omitempty"`

	// NEW: Warranty Information
	WarrantyStart    *time.Time `gorm:"type:date" json:"warranty_start,omitempty"`
	WarrantyEnd      *time.Time `gorm:"type:date;index" json:"warranty_end,omitempty"`
	WarrantyProvider *string    `gorm:"type:varchar(255)" json:"warranty_provider,omitempty"`
	WarrantyTerms    *string    `gorm:"type:text" json:"warranty_terms,omitempty"`

	// NEW: Insurance Information
	InsurancePolicyNumber *string    `gorm:"type:varchar(100)" json:"insurance_policy_number,omitempty"`
	InsuranceProvider     *string    `gorm:"type:varchar(255)" json:"insurance_provider,omitempty"`
	InsuranceStart        *time.Time `gorm:"type:date" json:"insurance_start,omitempty"`
	InsuranceEnd          *time.Time `gorm:"type:date" json:"insurance_end,omitempty"`
	InsuranceValue        *float64   `gorm:"type:numeric(18,2)" json:"insurance_value,omitempty"`

	// Audit & Metadata
	CreatedBy     *string `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedByUser *User   `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`

	// NEW: Approval Workflow
	ApprovedBy     *string    `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedByUser *User      `gorm:"foreignKey:ApprovedBy" json:"approved_by_user,omitempty"`
	ApprovedAt     *time.Time `gorm:"type:timestamptz" json:"approved_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Depreciations       []AssetDepreciation      `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE" json:"depreciations,omitempty"`
	Transactions        []AssetTransaction       `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE" json:"transactions,omitempty"`
	Attachments         []AssetAttachment        `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE" json:"attachments,omitempty"`
	AuditLogs           []AssetAuditLog          `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE" json:"audit_logs,omitempty"`
	AssignmentHistories []AssetAssignmentHistory `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE" json:"assignment_histories,omitempty"`
}

func (Asset) TableName() string {
	return "fixed_assets"
}

func (a *Asset) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.BookValue == 0 {
		a.BookValue = a.AcquisitionCost
	}
	// Set default lifecycle stage based on status
	if a.LifecycleStage == "" {
		switch a.Status {
		case AssetStatusDraft:
			a.LifecycleStage = AssetLifecycleDraft
		case AssetStatusActive, AssetStatusInUse:
			a.LifecycleStage = AssetLifecycleActive
		case AssetStatusDisposed:
			a.LifecycleStage = AssetLifecycleDisposed
		default:
			a.LifecycleStage = AssetLifecycleActive
		}
	}
	return nil
}

// TotalAcquisitionCost returns the total cost including additional costs
func (a *Asset) TotalAcquisitionCost() float64 {
	return a.ShippingCost + a.InstallationCost + a.TaxAmount + a.OtherCosts
}

// IsUnderWarranty checks if the asset is currently under warranty
func (a *Asset) IsUnderWarranty() bool {
	if a.WarrantyEnd == nil {
		return false
	}
	return time.Now().Before(*a.WarrantyEnd)
}

// WarrantyDaysRemaining returns the number of days remaining in warranty
func (a *Asset) WarrantyDaysRemaining() int {
	if a.WarrantyEnd == nil {
		return 0
	}
	days := int(a.WarrantyEnd.Sub(time.Now()).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// IsInsured checks if the asset has active insurance
func (a *Asset) IsInsured() bool {
	if a.InsuranceEnd == nil {
		return false
	}
	return time.Now().Before(*a.InsuranceEnd)
}

// GetDepreciationMethod returns the depreciation method (asset config overrides category)
func (a *Asset) GetDepreciationMethod() string {
	if a.DepreciationMethod != nil && *a.DepreciationMethod != "" {
		return *a.DepreciationMethod
	}
	if a.Category != nil {
		return string(a.Category.DepreciationMethod)
	}
	return "SL" // Default to straight line
}

// GetUsefulLifeMonths returns useful life in months
func (a *Asset) GetUsefulLifeMonths() int {
	if a.UsefulLifeMonths != nil && *a.UsefulLifeMonths > 0 {
		return *a.UsefulLifeMonths
	}
	if a.Category != nil && a.Category.UsefulLifeMonths > 0 {
		return a.Category.UsefulLifeMonths
	}
	return 60 // Default 5 years
}

// CanDepreciate checks if the asset can be depreciated
func (a *Asset) CanDepreciate() bool {
	return a.IsDepreciable && !a.IsFullyDepreciated && a.BookValue > a.SalvageValue
}

// IsAssigned checks if the asset is currently assigned to an employee
func (a *Asset) IsAssigned() bool {
	return a.AssignedToEmployeeID != nil && *a.AssignedToEmployeeID != ""
}

// AgeInMonths returns the age of the asset in months
func (a *Asset) AgeInMonths() int {
	months := int(time.Since(a.AcquisitionDate).Hours() / 24 / 30)
	if months < 0 {
		return 0
	}
	return months
}

// DepreciationProgress returns the depreciation progress as a percentage
func (a *Asset) DepreciationProgress() float64 {
	if a.AcquisitionCost <= a.SalvageValue {
		return 100
	}
	totalDepreciable := a.AcquisitionCost - a.SalvageValue
	if totalDepreciable <= 0 {
		return 100
	}
	progress := (a.AccumulatedDepreciation / totalDepreciable) * 100
	if progress > 100 {
		return 100
	}
	return progress
}
