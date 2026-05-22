package dto

import (
	"errors"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

// ===== CREATE ASSET REQUEST =====

type CreateAssetRequest struct {
	// Basic Information
	Code        *string `json:"code" example:"ASSET-001"` // Optional, auto-generated if not provided
	Name        string  `json:"name" binding:"required,max=200" example:"Laptop Dell"`
	Description *string `json:"description"`
	
	// Asset Classification
	AssetType    financeModels.AssetType `json:"asset_type" binding:"required,oneof=TANGIBLE INTANGIBLE NON_DEPRECIABLE CONSTRUCTION"`
	CategoryID   string                  `json:"category_id" binding:"required,uuid"`
	LocationID   string                  `json:"location_id" binding:"required,uuid"`
	DepartmentID *string                 `json:"department_id" binding:"omitempty,uuid"`
	CustodianID  *string                 `json:"custodian_id" binding:"omitempty,uuid"`
	
	// Identity Information
	SerialNumber *string `json:"serial_number" binding:"omitempty,max=100"`
	AssetTag     *string `json:"asset_tag" binding:"omitempty,max=50"`
	BarcodeOrQR  *string `json:"barcode_or_qr" binding:"omitempty,max=200"`
	
	// Financial Information
	AcquisitionDate time.Time `json:"acquisition_date" binding:"required" example:"2024-01-15"`
	AcquisitionCost float64   `json:"acquisition_cost" binding:"required,gt=0"`
	SalvageValue    *float64  `json:"salvage_value" binding:"omitempty,gte=0"`
	CurrencyCode    string    `json:"currency_code" binding:"required,len=3" example:"IDR"`
	ExchangeRate    *float64  `json:"exchange_rate" binding:"omitempty,gt=0" example:"1.00"`
	
	// Depreciation Configuration
	DepreciationMethod *financeModels.DepreciationMethod `json:"depreciation_method" binding:"omitempty,oneof=SL DB SYD UOP NONE"`
	UsefulLifeMonths   *int                              `json:"useful_life_months" binding:"omitempty,gt=0"`
	
	// Supplier/Procurement
	SupplierID        *string `json:"supplier_id" binding:"omitempty,uuid"`
	SupplierName      *string `json:"supplier_name" binding:"omitempty,max=200"`
	PurchaseOrderNumber *string `json:"purchase_order_number" binding:"omitempty,max=100"`
	InvoiceNumber     *string `json:"invoice_number" binding:"omitempty,max=100"`
	InvoiceDate       *time.Time `json:"invoice_date" binding:"omitempty"`
	
	// Warranty & Insurance
	WarrantyStartDate *time.Time `json:"warranty_start_date" binding:"omitempty"`
	WarrantyEndDate   *time.Time `json:"warranty_end_date" binding:"omitempty"`
	InsurancePolicyNumber *string `json:"insurance_policy_number" binding:"omitempty,max=100"`
	InsuranceValue    *float64 `json:"insurance_value" binding:"omitempty,gte=0"`
	
	// Additional Info
	Notes          *string `json:"notes" binding:"omitempty"`
	CostCenterCode *string `json:"cost_center_code" binding:"omitempty,max=50"`
}

// Validate performs custom validations
func (r *CreateAssetRequest) Validate() error {
	// Salvage value must not exceed acquisition cost
	if r.SalvageValue != nil && *r.SalvageValue > r.AcquisitionCost {
		return ErrSalvageValueExceeded
	}

	// Depreciation method validation
	if r.AssetType == financeModels.AssetTypeTangible {
		if r.DepreciationMethod == nil || *r.DepreciationMethod == financeModels.DepreciationMethodNone {
			if r.UsefulLifeMonths != nil && *r.UsefulLifeMonths == 0 {
				return ErrUsefulLifeRequired
			}
		}
	}

	if r.AssetType == financeModels.AssetTypeIntangible || r.AssetType == financeModels.AssetTypeNonDepreciable {
		if r.DepreciationMethod != nil && *r.DepreciationMethod != financeModels.DepreciationMethodNone {
			return ErrDepreciationNotSupported
		}
	}

	return nil
}

// ===== UPDATE ASSET REQUEST =====

type UpdateAssetRequest struct {
	Name            *string `json:"name" binding:"omitempty,max=200"`
	Description     *string `json:"description"`
	
	LocationID      *string `json:"location_id" binding:"omitempty,uuid"`
	DepartmentID    *string `json:"department_id" binding:"omitempty,uuid"`
	CustodianID     *string `json:"custodian_id" binding:"omitempty,uuid"`
	
	SerialNumber    *string `json:"serial_number" binding:"omitempty,max=100"`
	AssetTag        *string `json:"asset_tag" binding:"omitempty,max=50"`
	BarcodeOrQR     *string `json:"barcode_or_qr" binding:"omitempty,max=200"`
	
	SalvageValue    *float64 `json:"salvage_value" binding:"omitempty,gte=0"`
	
	DepreciationMethod *financeModels.DepreciationMethod `json:"depreciation_method" binding:"omitempty,oneof=SL DB SYD UOP NONE"`
	UsefulLifeMonths   *int   `json:"useful_life_months" binding:"omitempty,gt=0"`
	
	WarrantyStartDate  *time.Time `json:"warranty_start_date" binding:"omitempty"`
	WarrantyEndDate    *time.Time `json:"warranty_end_date" binding:"omitempty"`
	InsurancePolicyNumber *string `json:"insurance_policy_number" binding:"omitempty,max=100"`
	InsuranceValue     *float64 `json:"insurance_value" binding:"omitempty,gte=0"`
	
	Notes          *string `json:"notes"`
	CostCenterCode *string `json:"cost_center_code" binding:"omitempty,max=50"`
}

// ===== ACTIVATE ASSET REQUEST =====

type ActivateAssetRequest struct {
	ActivationDate *time.Time `json:"activation_date" binding:"omitempty"`
}

// ===== DEPRECIATION SCHEDULE REQUEST =====

type GenerateDepreciationScheduleRequest struct {
	StartDate *time.Time `json:"start_date" binding:"omitempty"`
	EndDate   *time.Time `json:"end_date" binding:"omitempty"`
	PreviewOnly bool `json:"preview_only" binding:"omitempty"` // If true, don't save
}

// ===== DISPOSE ASSET REQUEST =====

type DisposeAssetRequest struct {
	DisposalType    financeModels.DisposalType `json:"disposal_type" binding:"required,oneof=SOLD SCRAPPED DONATED TRANSFERRED OTHER"`
	DisposalDate    time.Time                  `json:"disposal_date" binding:"required"`
	DisposalValue   *float64                   `json:"disposal_value" binding:"omitempty,gte=0"`
	ReferenceNumber *string                    `json:"reference_number" binding:"omitempty,max=100"`
	Description     *string                    `json:"description"`
}

// ===== TRANSFER ASSET REQUEST =====

type TransferAssetRequest struct {
	ToLocationID   string  `json:"to_location_id" binding:"required,uuid"`
	ToDepartmentID string  `json:"to_department_id" binding:"required,uuid"`
	CustodianUserID *string `json:"custodian_user_id" binding:"omitempty,uuid"`
	IsIntercompany bool    `json:"is_intercompany"`
	EffectiveDate  string  `json:"effective_date" binding:"required"`
	Notes          string  `json:"notes"`
}

type RejectTransferRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ===== MAINTENANCE LOG REQUEST =====

type CreateMaintenanceLogRequest struct {
	MaintenanceType financeModels.MaintenanceType `json:"maintenance_type" binding:"required,oneof=PREVENTIVE CORRECTIVE EMERGENCY"`
	MaintenanceDate time.Time                     `json:"maintenance_date" binding:"required"`
	Description     *string                       `json:"description"`
	Cost            *float64                      `json:"cost" binding:"omitempty,gte=0"`
	ServiceProvider *string                       `json:"service_provider" binding:"omitempty,max=200"`
	PartsReplaced   *string                       `json:"parts_replaced"`
	DurationHours   *float64                      `json:"duration_hours" binding:"omitempty,gt=0"`
	NextMaintenanceDate *time.Time `json:"next_maintenance_date" binding:"omitempty"`
}

// ===== LIST ASSETS REQUEST =====

type ListAssetsRequest struct {
	Page        int    `form:"page,default=1" binding:"min=1"`
	PerPage     int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Status      *string `form:"status"`
	CategoryID  *string `form:"category_id"`
	LocationID  *string `form:"location_id"`
	DepartmentID *string `form:"department_id"`
	CustodianID *string `form:"custodian_id"`
	AssetType   *string `form:"asset_type"`
	Search      *string `form:"search"` // Search name, code, serial number
	SortBy      string `form:"sort_by,default=created_at"`
	Order       string `form:"order,default=desc" binding:"oneof=asc desc"`
}

// ===== RESPONSE TYPES =====

type AssetResponse struct {
	ID              string                           `json:"id"`
	Code            string                           `json:"code"`
	Name            string                           `json:"name"`
	Description     *string                          `json:"description"`
	
	AssetType       financeModels.AssetType          `json:"asset_type"`
	Status          financeModels.AssetStatus        `json:"status"`
	
	CategoryID      string                           `json:"category_id"`
	Category        *AssetCategoryResponse           `json:"category,omitempty"`
	LocationID      string                           `json:"location_id"`
	Location        *AssetLocationResponse           `json:"location,omitempty"`
	DepartmentID    *string                          `json:"department_id"`
	CustodianID     *string                          `json:"custodian_id"`
	
	SerialNumber    *string                          `json:"serial_number"`
	AssetTag        *string                          `json:"asset_tag"`
	BarcodeOrQR     *string                          `json:"barcode_or_qr"`
	
	AcquisitionDate time.Time                        `json:"acquisition_date"`
	AcquisitionCost float64                          `json:"acquisition_cost"`
	SalvageValue    float64                          `json:"salvage_value"`
	CurrencyCode    string                           `json:"currency_code"`
	BookValue       *float64                         `json:"book_value"`
	AccumulatedDepreciation float64                  `json:"accumulated_depreciation"`
	
	DepreciationMethod     *financeModels.DepreciationMethod `json:"depreciation_method"`
	UsefulLifeMonths       *int                               `json:"useful_life_months"`
	DepreciationStartDate  *time.Time                         `json:"depreciation_start_date"`
	
	SupplierName           *string                  `json:"supplier_name"`
	InvoiceNumber          *string                  `json:"invoice_number"`
	
	IsActive               bool                     `json:"is_active"`
	ActivationDate         *time.Time               `json:"activation_date"`
	DisposalDate           *time.Time               `json:"disposal_date"`
	
	WarrantyStartDate      *time.Time               `json:"warranty_start_date"`
	WarrantyEndDate        *time.Time               `json:"warranty_end_date"`
	InsurancePolicyNumber  *string                  `json:"insurance_policy_number"`
	InsuranceValue         *float64                 `json:"insurance_value"`
	
	Notes                  *string                  `json:"notes"`
	
	Attachments            []AssetAttachmentResponse `json:"attachments,omitempty"`
	LatestMaintenanceLog   *MaintenanceLogResponse   `json:"latest_maintenance_log,omitempty"`
	
	CreatedAt              time.Time                 `json:"created_at"`
	UpdatedAt              time.Time                 `json:"updated_at"`
}

type AssetCategoryResponse struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	DefaultUsefulLife   *int      `json:"default_useful_life_months"`
	DefaultMethod       *string   `json:"default_depreciation_method"`
}

type AssetLocationResponse struct {
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	City     *string `json:"city"`
	Country  *string `json:"country"`
}

type DepreciationScheduleResponse struct {
	Period                  string    `json:"period"`
	PeriodMonth             int       `json:"period_month"`
	DepreciationAmount      float64   `json:"depreciation_amount"`
	AccumulatedDepreciation float64   `json:"accumulated_depreciation"`
	BookValue               float64   `json:"book_value"`
	IsPosted                bool      `json:"is_posted"`
}

type AssetAttachmentResponse struct {
	ID              string    `json:"id"`
	FileName        string    `json:"file_name"`
	FileType        *string   `json:"file_type"`
	FileSize        *int64    `json:"file_size"`
	FilePath        string    `json:"file_path"`
	AttachmentType  string    `json:"attachment_type"`
	UploadedAt      time.Time `json:"uploaded_at"`
}

type MaintenanceLogResponse struct {
	ID                  string    `json:"id"`
	MaintenanceType     string    `json:"maintenance_type"`
	MaintenanceDate     time.Time `json:"maintenance_date"`
	Description         *string   `json:"description"`
	Cost                *float64  `json:"cost"`
	NextMaintenanceDate *time.Time `json:"next_maintenance_date"`
	CreatedAt           time.Time `json:"created_at"`
}

type DepreciationPreviewResponse struct {
	Method              financeModels.DepreciationMethod `json:"method"`
	AcquisitionCost     float64                          `json:"acquisition_cost"`
	SalvageValue        float64                          `json:"salvage_value"`
	UsefulLifeMonths    int                              `json:"useful_life_months"`
	TotalDepreciation   float64                          `json:"total_depreciation"`
	MonthlyAmount       float64                          `json:"monthly_amount"`
	AnnualAmount        float64                          `json:"annual_amount"`
	Schedules           []DepreciationScheduleResponse   `json:"schedules"`
}

type DisposalResponse struct {
	ID              string    `json:"id"`
	AssetID         string    `json:"asset_id"`
	DisposalType    string    `json:"disposal_type"`
	DisposalDate    time.Time `json:"disposal_date"`
	DisposalValue   *float64  `json:"disposal_value"`
	GainOrLoss      *float64  `json:"gain_or_loss"`
	Description     *string   `json:"description"`
	ApprovedAt      *time.Time `json:"approved_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// ===== ERROR TYPES =====

var (
	ErrAssetNotFound              = errors.New("asset not found")
	ErrAssetCodeExists            = errors.New("asset code already exists")
	ErrSerialNumberExists         = errors.New("serial number already exists")
	ErrSalvageValueExceeded       = errors.New("salvage value cannot exceed acquisition cost")
	ErrUsefulLifeRequired         = errors.New("useful life is required for depreciable assets")
	ErrDepreciationNotSupported   = errors.New("depreciation not supported for this asset type")
	ErrInvalidStatusTransition    = errors.New("invalid status transition")
	ErrAssetNotDepreciable        = errors.New("asset is not depreciable")
	ErrCannotModifyActive         = errors.New("cannot modify active asset depreciation settings")
	ErrCannotDisposeNonActive     = errors.New("can only dispose active assets")
	ErrInvalidDepreciationMethod  = errors.New("invalid depreciation method")
	ErrMissingRequiredFields      = errors.New("missing required fields")
)
