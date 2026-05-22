package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/google/uuid"
)

type CreateAssetRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`

	AssetTypeID string `json:"asset_type_id" binding:"required"`
	CategoryID  string `json:"asset_category_id" binding:"required,uuid"`
	LocationID  string `json:"location_id" binding:"required,uuid"`

	AcquisitionDate string   `json:"acquisition_date" binding:"required"`
	PurchasePrice   float64  `json:"purchase_price"`
	AcquisitionCost float64  `json:"acquisition_cost"`
	SalvageValue    float64  `json:"salvage_value" binding:"omitempty,gte=0"`
	UsefulLifeMonths *int    `json:"useful_life_months" binding:"required,gte=1"`
	DepreciationMethod *string `json:"depreciation_method" binding:"required"`
	VendorID        *string `json:"vendor_id"`
	PurchaseInvoiceID *string `json:"purchase_invoice_id"`
	DepartmentID    *string `json:"department_id"`
	CustodianUserID *string `json:"custodian_user_id"`
	CompanyID       *string `json:"company_id" binding:"omitempty,uuid"`
	BusinessUnitID  *string `json:"business_unit_id" binding:"omitempty,uuid"`
	ParentAssetID   *string `json:"parent_asset_id" binding:"omitempty,uuid"`
	AssignedToEmployeeID *string `json:"assigned_to_employee_id" binding:"omitempty,uuid"`

	// Extended fields
	SerialNumber *string `json:"serial_number"`
	Barcode      *string `json:"barcode"`
	AssetTag     *string `json:"asset_tag"`

	// Cost Breakdown
	ShippingCost     float64 `json:"shipping_cost" binding:"omitempty,gte=0"`
	InstallationCost float64 `json:"installation_cost" binding:"omitempty,gte=0"`
	TaxAmount        float64 `json:"tax_amount" binding:"omitempty,gte=0"`
	OtherCosts       float64 `json:"other_costs" binding:"omitempty,gte=0"`
	DepreciationStartDate *string `json:"depreciation_start_date"`

	// Warranty
	WarrantyStart    *string `json:"warranty_start"`
	WarrantyEnd      *string `json:"warranty_end"`
	WarrantyProvider *string `json:"warranty_provider"`
	WarrantyTerms    *string `json:"warranty_terms"`

	// Insurance
	InsurancePolicyNumber *string  `json:"insurance_policy_number"`
	InsuranceProvider     *string  `json:"insurance_provider"`
	InsuranceStart        *string  `json:"insurance_start"`
	InsuranceEnd          *string  `json:"insurance_end"`
	InsuranceValue        *float64 `json:"insurance_value"`
}

type UpdateAssetRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`

	AssetTypeID string `json:"asset_type_id" binding:"required"`
	CategoryID  string `json:"asset_category_id" binding:"required,uuid"`
	LocationID  string `json:"location_id" binding:"required,uuid"`

	AcquisitionDate string                    `json:"acquisition_date" binding:"required"`
	AcquisitionCost float64                   `json:"acquisition_cost" binding:"required,gt=0"`
	SalvageValue    float64                   `json:"salvage_value" binding:"omitempty,gte=0"`
	Status          financeModels.AssetStatus `json:"status" binding:"omitempty,oneof=active inactive sold disposed"`
	UsefulLifeMonths *int                     `json:"useful_life_months" binding:"omitempty,gte=1"`
	DepreciationMethod *string                `json:"depreciation_method"`
	VendorID        *string                   `json:"vendor_id"`
	PurchaseInvoiceID *string                 `json:"purchase_invoice_id"`
	DepartmentID    *string                   `json:"department_id"`
	CustodianUserID *string                   `json:"custodian_user_id"`

	// Extended fields
	SerialNumber *string `json:"serial_number"`
	Barcode      *string `json:"barcode"`
	AssetTag     *string `json:"asset_tag"`

	// Cost Breakdown
	ShippingCost     float64 `json:"shipping_cost" binding:"omitempty,gte=0"`
	InstallationCost float64 `json:"installation_cost" binding:"omitempty,gte=0"`
	TaxAmount        float64 `json:"tax_amount" binding:"omitempty,gte=0"`
	OtherCosts       float64 `json:"other_costs" binding:"omitempty,gte=0"`

	// Warranty
	WarrantyStart    *string `json:"warranty_start"`
	WarrantyEnd      *string `json:"warranty_end"`
	WarrantyProvider *string `json:"warranty_provider"`
	WarrantyTerms    *string `json:"warranty_terms"`

	// Insurance
	InsurancePolicyNumber *string  `json:"insurance_policy_number"`
	InsuranceProvider     *string  `json:"insurance_provider"`
	InsuranceStart        *string  `json:"insurance_start"`
	InsuranceEnd          *string  `json:"insurance_end"`
	InsuranceValue        *float64 `json:"insurance_value"`
}

type ListAssetsRequest struct {
	Page       int                        `form:"page" binding:"omitempty,min=1"`
	PerPage    int                        `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search     string                     `form:"search"`
	Status     *financeModels.AssetStatus `form:"status" binding:"omitempty,oneof=draft pending_capitalization active in_use under_maintenance idle disposed sold written_off transferred inactive"`
	CategoryID *string                    `form:"category_id"`
	TypeID     *string                    `form:"type_id"`
	DeptID     *string                    `form:"dept_id"`
	LocationID *string                    `form:"location_id"`
	StartDate  *string                    `form:"start_date"`
	EndDate    *string                    `form:"end_date"`
	DateFrom   *string                    `form:"date_from"`
	DateTo     *string                    `form:"date_to"`
	WarrantyExpiringDays *int             `form:"warranty_expiring_days" binding:"omitempty,oneof=30 60 90"`
	IsCapitalized *bool                   `form:"is_capitalized"`
	SortBy     string                     `form:"sort_by"`
	SortDir    string                     `form:"sort_dir"`
	Sort       string                     `form:"sort"`
	Order      string                     `form:"order"`
}

type DepreciateAssetRequest struct {
	AsOfDate string `json:"as_of_date" binding:"required"`
}

type BatchDepreciationRequest struct {
	PeriodMonth int `json:"period_month" binding:"required,min=1,max=12"`
	PeriodYear  int `json:"period_year" binding:"required,min=1900,max=3000"`
}

type BatchDepreciationPreviewItem struct {
	AssetID            string  `json:"asset_id"`
	AssetCode          string  `json:"asset_code"`
	AssetName          string  `json:"asset_name"`
	CurrencyCode       string  `json:"currency_code,omitempty"`
	Method             string  `json:"method"`
	OpeningBookValue   float64 `json:"opening_book_value"`
	DepreciationAmount float64 `json:"depreciation_amount"`
	ProjectedBookValue float64 `json:"projected_book_value"`
	Eligible           bool    `json:"eligible"`
	SkipReason         string  `json:"skip_reason,omitempty"`
}

type BatchDepreciationPreviewResponse struct {
	PeriodMonth int                            `json:"period_month"`
	PeriodYear  int                            `json:"period_year"`
	CurrencyCode string                        `json:"currency_code,omitempty"`
	TotalAssets int                            `json:"total_assets"`
	Eligible    int                            `json:"eligible"`
	Skipped     int                            `json:"skipped"`
	Items       []BatchDepreciationPreviewItem `json:"items"`
}

type BatchDepreciationRunItem struct {
	AssetID        string  `json:"asset_id"`
	AssetCode      string  `json:"asset_code"`
	AssetName      string  `json:"asset_name"`
	CurrencyCode   string  `json:"currency_code,omitempty"`
	Amount         float64 `json:"amount"`
	Status         string  `json:"status"` // posted | skipped | failed
	Reason         string  `json:"reason,omitempty"`
	JournalEntryID *string `json:"journal_entry_id,omitempty"`
}

type BatchDepreciationRunResponse struct {
	PeriodMonth int                        `json:"period_month"`
	PeriodYear  int                        `json:"period_year"`
	CurrencyCode string                    `json:"currency_code,omitempty"`
	Processed   int                        `json:"processed"`
	Posted      int                        `json:"posted"`
	Skipped     int                        `json:"skipped"`
	Failed      int                        `json:"failed"`
	Items       []BatchDepreciationRunItem `json:"items"`
}

type RunDepreciationRequest struct {
	Period     string  `json:"period" binding:"required"`
	AssetID   *string `json:"asset_id" binding:"omitempty,uuid"`
	CategoryID *string `json:"category_id" binding:"omitempty,uuid"`
	LocationID *string `json:"location_id" binding:"omitempty,uuid"`
}

type DepreciationScheduleItem struct {
	AssetID                 string  `json:"asset_id"`
	AssetCode               string  `json:"asset_code"`
	AssetName               string  `json:"asset_name"`
	CategoryID              string  `json:"category_id"`
	CategoryName            string  `json:"category_name,omitempty"`
	LocationID              string  `json:"location_id"`
	LocationName            string  `json:"location_name,omitempty"`
	Method                  string  `json:"method"`
	Period                  string  `json:"period"`
	AcquisitionCost         float64 `json:"acquisition_cost"`
	AccumulatedDepreciation float64 `json:"accumulated_depreciation"`
	NetBookValue            float64 `json:"net_book_value"`
	DepreciationAmount      float64 `json:"depreciation_amount"`
	Posted                  bool    `json:"posted"`
	Highlighted             bool    `json:"highlighted"`
	JournalEntryID          *string `json:"journal_entry_id,omitempty"`
	Status                  string  `json:"status"`
	SkipReason              string  `json:"skip_reason,omitempty"`
}

type DepreciationScheduleResponse struct {
	Period      string                   `json:"period"`
	TotalAssets int                      `json:"total_assets"`
	Posted      int                      `json:"posted"`
	Pending     int                      `json:"pending"`
	TotalAmount float64                  `json:"total_amount"`
	Items       []DepreciationScheduleItem `json:"items"`
}

type DepreciationHistoryItem struct {
	DepreciationID string    `json:"depreciation_id"`
	AssetID        string    `json:"asset_id"`
	AssetCode      string    `json:"asset_code"`
	AssetName      string    `json:"asset_name"`
	Period         string    `json:"period"`
	Amount         float64   `json:"amount"`
	Posted         bool      `json:"posted"`
	JournalEntryID *string   `json:"journal_entry_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type DepreciationHistoryResponse struct {
	Period  string                   `json:"period"`
	Total   int                      `json:"total"`
	Posted  int                      `json:"posted"`
	Pending int                      `json:"pending"`
	Items   []DepreciationHistoryItem `json:"items"`
}

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

type ListTransfersRequest struct {
	Status  string `form:"status"`
	AssetID string `form:"asset_id"`
	Page    int    `form:"page,default=1" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page,default=50" binding:"omitempty,min=1,max=500"`
}

type AssetTransferResponse struct {
	ID                  string     `json:"id"`
	AssetID             string     `json:"asset_id"`
	AssetCode           string     `json:"asset_code,omitempty"`
	AssetName           string     `json:"asset_name,omitempty"`
	DivisionsName       string     `json:"divisions_name,omitempty"`
	DivisionsCode       string     `json:"divisions_code,omitempty"`
	FromLocationID      *string    `json:"from_location_id,omitempty"`
	ToLocationID        *string    `json:"to_location_id,omitempty"`
	FromDepartmentID    *string    `json:"from_department_id,omitempty"`
	ToDepartmentID      *string    `json:"to_department_id,omitempty"`
	FromCustodianID     *string    `json:"from_custodian_id,omitempty"`
	ToCustodianID       *string    `json:"to_custodian_id,omitempty"`
	FromCompanyID       *string    `json:"from_company_id,omitempty"`
	ToCompanyID         *string    `json:"to_company_id,omitempty"`
	TransferDate        string     `json:"transfer_date"`
	Notes               *string    `json:"notes,omitempty"`
	Status              string     `json:"status"`
	IsIntercompany      bool       `json:"is_intercompany"`
	CurrentApprovalRole string     `json:"current_approval_role,omitempty"`
	ApprovalStepIndex   int        `json:"approval_step_index"`
	ApprovalStepTotal   int        `json:"approval_step_total"`
	RequestedBy         *string    `json:"requested_by,omitempty"`
	RequestedByName     *string    `json:"requested_by_name,omitempty"`
	RequestedAt         time.Time  `json:"requested_at"`
	DhApprovedBy        *string    `json:"dh_approved_by,omitempty"`
	DhApprovedAt        *time.Time `json:"dh_approved_at,omitempty"`
	FcApprovedBy        *string    `json:"fc_approved_by,omitempty"`
	FcApprovedAt        *time.Time `json:"fc_approved_at,omitempty"`
	RejectedBy          *string    `json:"rejected_by,omitempty"`
	RejectedAt          *time.Time `json:"rejected_at,omitempty"`
	RejectionReason     *string    `json:"rejection_reason,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type DisposeAssetRequest struct {
	DisposalDate   string  `json:"disposal_date" binding:"required"`
	ProceedsAmount float64 `json:"proceeds_amount" binding:"omitempty,gte=0"`
	BankAccountID  *string `json:"bank_account_id" binding:"omitempty,uuid"`
	Description    string  `json:"description"`
}

type PreviewDisposalRequest struct {
	DisposalDate   string  `json:"disposal_date" binding:"required"`
	ProceedsAmount float64 `json:"proceeds_amount" binding:"omitempty,gte=0"`
}

type PreviewDisposalResponse struct {
	AssetID         string  `json:"asset_id"`
	AssetCode       string  `json:"asset_code"`
	AssetName       string  `json:"asset_name"`
	DisposalDate    string  `json:"disposal_date"`
	BookValue       float64 `json:"book_value"`
	ProceedsAmount  float64 `json:"proceeds_amount"`
	GainLossAmount  float64 `json:"gain_loss_amount"`
	GainLossType    string  `json:"gain_loss_type"`
	GainLossAccount *string `json:"gain_loss_account,omitempty"`
}

type SellAssetRequest struct {
	DisposalDate string  `json:"disposal_date" binding:"required"`
	SaleAmount   float64 `json:"sale_amount" binding:"required,gt=0"`
	Description  string  `json:"description"`
}

type RevalueAssetRequest struct {
	RevaluationDate string  `json:"revaluation_date" binding:"required"`
	NewValue        float64 `json:"new_value" binding:"required,gt=0"`
	Description     string  `json:"description"`
}

type AdjustAssetRequest struct {
	AdjustmentDate   string  `json:"adjustment_date" binding:"required"`
	AdjustmentAmount float64 `json:"adjustment_amount" binding:"required"`
	Description      string  `json:"description"`
}

// --- Assignment / Return ---

type AssignAssetRequest struct {
	EmployeeID   string  `json:"employee_id" binding:"required,uuid"`
	DepartmentID *string `json:"department_id" binding:"omitempty,uuid"`
	LocationID   *string `json:"location_id" binding:"omitempty,uuid"`
	Notes        *string `json:"notes"`
}

type ReturnAssetRequest struct {
	ReturnDate   string  `json:"return_date" binding:"required"`
	ReturnReason *string `json:"return_reason"`
}

// --- Attachment ---

type CreateAttachmentRequest struct {
	FileType    string  `json:"file_type" binding:"required"`
	Description *string `json:"description"`
}

// ----- Responses -----

type AssetDepreciationResponse struct {
	ID               string                           `json:"id"`
	AssetID          string                           `json:"asset_id"`
	Period           string                           `json:"period"`
	DepreciationDate time.Time                        `json:"depreciation_date"`
	Method           financeModels.DepreciationMethod `json:"method"`
	Amount           float64                          `json:"amount"`
	Accumulated      float64                          `json:"accumulated"`
	BookValue        float64                          `json:"book_value"`
	JournalEntryID   *string                          `json:"journal_entry_id"`
	CreatedAt        time.Time                        `json:"created_at"`
}

type AssetTransactionResponse struct {
	ID                     string                             `json:"id"`
	AssetID                string                             `json:"asset_id"`
	Type                   financeModels.AssetTransactionType `json:"type"`
	TransactionDate        time.Time                          `json:"transaction_date"`
	Amount                 float64                            `json:"amount"`
	Description            string                             `json:"description"`
	Status                 string                             `json:"status"`
	ReferenceType          *string                            `json:"reference_type"`
	ReferenceID            *string                            `json:"reference_id"`
	ProceedsAmount         float64                            `json:"proceeds_amount"`
	BankAccountID          *string                            `json:"bank_account_id,omitempty"`
	BookValueAtTransaction float64                            `json:"book_value_at_transaction"`
	GainLossAmount         float64                            `json:"gain_loss_amount"`
	GainLossAccountID      *string                            `json:"gain_loss_account_id,omitempty"`
	CreatedAt              time.Time                          `json:"created_at"`
}

type AssetAttachmentResponse struct {
	ID          string    `json:"id"`
	AssetID     string    `json:"asset_id"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	FileURL     string    `json:"file_url"`
	FileType    string    `json:"file_type"`
	FileSize    *int      `json:"file_size,omitempty"`
	MimeType    *string   `json:"mime_type,omitempty"`
	Description *string   `json:"description,omitempty"`
	UploadedBy  *string   `json:"uploaded_by,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditChangeResponse struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

type AssetAuditLogResponse struct {
	ID          string                 `json:"id"`
	AssetID     string                 `json:"asset_id"`
	Action      string                 `json:"action"`
	Changes     []AuditChangeResponse  `json:"changes,omitempty"`
	PerformedBy *string                `json:"performed_by,omitempty"`
	PerformedAt time.Time              `json:"performed_at"`
	IPAddress   *string                `json:"ip_address,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type AssetAssignmentHistoryResponse struct {
	ID           string     `json:"id"`
	AssetID      string     `json:"asset_id"`
	EmployeeID   *string    `json:"employee_id,omitempty"`
	EmployeeName *string    `json:"employee_name,omitempty"`
	DepartmentID *string    `json:"department_id,omitempty"`
	LocationID   *string    `json:"location_id,omitempty"`
	AssignedAt   time.Time  `json:"assigned_at"`
	AssignedBy   *string    `json:"assigned_by,omitempty"`
	ReturnedAt   *time.Time `json:"returned_at,omitempty"`
	ReturnReason *string    `json:"return_reason,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type AssetResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Identity
	SerialNumber *string `json:"serial_number,omitempty"`
	Barcode      *string `json:"barcode,omitempty"`
	QRCode       *string `json:"qr_code,omitempty"`
	AssetTag     *string `json:"asset_tag,omitempty"`

	// Category / Location
	AssetTypeID string                 `json:"asset_type_id"`
	CategoryID string                 `json:"category_id"`
	Category   *AssetCategoryResponse `json:"category,omitempty"`
	CategoryName string                `json:"category_name,omitempty"`
	AssetType    string                `json:"asset_type,omitempty"`
	LocationID string                 `json:"location_id"`
	Location   *AssetLocationResponse `json:"location,omitempty"`
	LocationName string                `json:"location_name,omitempty"`
	DeptName     string                `json:"dept_name,omitempty"`

	// Organization
	CompanyID      *string `json:"company_id,omitempty"`
	BusinessUnitID *string `json:"business_unit_id,omitempty"`
	DepartmentID   *string `json:"department_id,omitempty"`
	CustodianUserID *string `json:"custodian_user_id,omitempty"`
	VendorID       *string `json:"vendor_id,omitempty"`
	PurchaseInvoiceID *string `json:"purchase_invoice_id,omitempty"`

	// Assignment
	AssignedToEmployeeID *string    `json:"assigned_to_employee_id,omitempty"`
	AssignmentDate       *time.Time `json:"assignment_date,omitempty"`

	// Acquisition
	AcquisitionDate time.Time `json:"acquisition_date"`
	AcquisitionCost float64   `json:"acquisition_cost"`
	SalvageValue    float64   `json:"salvage_value"`

	// Cost Breakdown
	ShippingCost     float64 `json:"shipping_cost"`
	InstallationCost float64 `json:"installation_cost"`
	TaxAmount        float64 `json:"tax_amount"`
	OtherCosts       float64 `json:"other_costs"`
	TotalCost        float64 `json:"total_cost"`

	// Depreciation
	AccumulatedDepreciation float64 `json:"accumulated_depreciation"`
	BookValue               float64 `json:"book_value"`

	// Depreciation Config
	DepreciationMethod    *string    `json:"depreciation_method,omitempty"`
	UsefulLifeMonths      *int       `json:"useful_life_months,omitempty"`
	DepreciationStartDate *time.Time `json:"depreciation_start_date,omitempty"`

	// Status / Lifecycle
	Status               financeModels.AssetStatus         `json:"status"`
	LifecycleStage       financeModels.AssetLifecycleStage `json:"lifecycle_stage"`
	DisplayStatus        string                            `json:"display_status"`
	IsCapitalized        bool                              `json:"is_capitalized"`
	IsDepreciable        bool                              `json:"is_depreciable"`
	IsFullyDepreciated   bool                              `json:"is_fully_deprecated"`
	DisposedAt           *time.Time                        `json:"disposed_at,omitempty"`
	DepreciationProgress float64                           `json:"depreciation_progress"`
	AgeInMonths          int                               `json:"age_in_months"`

	// Parent/Child
	ParentAssetID *string `json:"parent_asset_id,omitempty"`
	IsParent      bool    `json:"is_parent"`

	// Warranty
	WarrantyStart         *time.Time `json:"warranty_start,omitempty"`
	WarrantyEnd           *time.Time `json:"warranty_end,omitempty"`
	WarrantyProvider      *string    `json:"warranty_provider,omitempty"`
	WarrantyTerms         *string    `json:"warranty_terms,omitempty"`
	IsUnderWarranty       bool       `json:"is_under_warranty"`
	WarrantyDaysRemaining int        `json:"warranty_days_remaining"`
	WarrantyExpiryDays    *int       `json:"warranty_expiry_days,omitempty"`

	// Insurance
	InsurancePolicyNumber *string    `json:"insurance_policy_number,omitempty"`
	InsuranceProvider     *string    `json:"insurance_provider,omitempty"`
	InsuranceStart        *time.Time `json:"insurance_start,omitempty"`
	InsuranceEnd          *time.Time `json:"insurance_end,omitempty"`
	InsuranceValue        *float64   `json:"insurance_value,omitempty"`
	IsInsured             bool       `json:"is_insured"`

	// Audit
	CreatedBy  *string    `json:"created_by,omitempty"`
	ApprovedBy *string    `json:"approved_by,omitempty"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations (populated with details)
	Depreciations       []AssetDepreciationResponse      `json:"depreciations,omitempty"`
	Transactions        []AssetTransactionResponse       `json:"transactions,omitempty"`
	Attachments         []AssetAttachmentResponse        `json:"attachments,omitempty"`
	AuditLogs           []AssetAuditLogResponse          `json:"audit_logs,omitempty"`
	AssignmentHistories []AssetAssignmentHistoryResponse `json:"assignment_histories,omitempty"`
}

type CreateAssetFromPurchaseRequest struct {
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	AcquisitionDate string  `json:"acquisition_date"`
	AcquisitionCost float64 `json:"acquisition_cost"`
	ReferenceType   string  `json:"reference_type"`
	ReferenceID     string  `json:"reference_id"`
	CategoryID      *string `json:"category_id"`
	LocationID      *string `json:"location_id"`
}

// AssetMiniResponse is a minimal response for asset references
type AssetMiniResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// AvailableAssetCategoryLite for available assets list
type AvailableAssetCategoryLite struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AvailableAssetLocationLite for available assets list
type AvailableAssetLocationLite struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AvailableAssetResponse for employee asset borrowing
type AvailableAssetResponse struct {
	ID         string                      `json:"id"`
	Code       string                      `json:"code"`
	Name       string                      `json:"name"`
	Category   *AvailableAssetCategoryLite `json:"category,omitempty"`
	Location   *AvailableAssetLocationLite `json:"location,omitempty"`
	AssetImage string                      `json:"asset_image,omitempty"`
	Status     string                      `json:"status"`
	BookValue  float64                     `json:"book_value"`
}

// --- Edit Asset with Field Classification ---

// DepreciationImpactPreview shows the effect of depreciation-sensitive changes
type DepreciationImpactPreview struct {
	OldMonthlyAmount     float64 `json:"old_monthly_amount"`
	NewMonthlyAmount     float64 `json:"new_monthly_amount"`
	RemainingMonths      int     `json:"remaining_months"`
	OldTotalRemaining    float64 `json:"old_total_remaining"`
	NewTotalRemaining    float64 `json:"new_total_remaining"`
	ImpactAmount         float64 `json:"impact_amount"`
	NewBookValue         float64 `json:"new_book_value"`
	EntriesToRegenerate  int     `json:"entries_to_regenerate"`
	FirstAffectedPeriod  string  `json:"first_affected_period"` // YYYY-MM
	DepreciationMethod   string  `json:"depreciation_method"`
}

// FieldChangeInfo tracks what changed and its classification
type FieldChangeInfo struct {
	FieldName             string                     `json:"field_name"`
	OldValue              interface{}                `json:"old_value"`
	NewValue              interface{}                `json:"new_value"`
	Group                 string                     `json:"group"` // A, B, C
	DepreciationImpact    *DepreciationImpactPreview `json:"depreciation_impact,omitempty"`
}

// EditAssetRequest is similar to UpdateAssetRequest but tracks field changes
type EditAssetRequest struct {
	Code                  string   `json:"code"`
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	AssetTypeID           string   `json:"asset_type_id"`
	CategoryID            string   `json:"asset_category_id" binding:"omitempty,uuid"`
	LocationID            string   `json:"location_id" binding:"omitempty,uuid"`
	AcquisitionDate       string   `json:"acquisition_date"`
	AcquisitionCost       *float64 `json:"acquisition_cost" binding:"omitempty,gt=0"`
	SalvageValue          *float64 `json:"salvage_value" binding:"omitempty,gte=0"`
	Status                financeModels.AssetStatus `json:"status" binding:"omitempty,oneof=pending_approval active inactive in_use idle under_maintenance disposed sold written_off transferred"`
	UsefulLifeMonths      *int     `json:"useful_life_months" binding:"omitempty,gte=1"`
	DepreciationMethod    *string  `json:"depreciation_method"`
	VendorID              *string  `json:"vendor_id"`
	PurchaseInvoiceID     *string  `json:"purchase_invoice_id"`
	DepartmentID          *string  `json:"department_id"`
	AssignedToEmployeeID   *string  `json:"assigned_to_employee_id"`
	CustodianUserID       *string  `json:"custodian_user_id"`
	SerialNumber          *string  `json:"serial_number"`
	Barcode               *string  `json:"barcode"`
	AssetTag              *string  `json:"asset_tag"`
	ShippingCost          *float64 `json:"shipping_cost" binding:"omitempty,gte=0"`
	InstallationCost      *float64 `json:"installation_cost" binding:"omitempty,gte=0"`
	TaxAmount             *float64 `json:"tax_amount" binding:"omitempty,gte=0"`
	OtherCosts            *float64 `json:"other_costs" binding:"omitempty,gte=0"`
	WarrantyStart         *string  `json:"warranty_start"`
	WarrantyEnd           *string  `json:"warranty_end"`
	WarrantyProvider      *string  `json:"warranty_provider"`
	WarrantyTerms         *string  `json:"warranty_terms"`
	InsurancePolicyNumber *string  `json:"insurance_policy_number"`
	InsuranceProvider     *string  `json:"insurance_provider"`
	InsuranceStart        *string  `json:"insurance_start"`
	InsuranceEnd          *string  `json:"insurance_end"`
	InsuranceValue        *float64 `json:"insurance_value"`
	CascadeToChildren     bool     `json:"cascade_to_children"`
}

// EditAssetResponse includes field changes and depreciation impact
type EditAssetResponse struct {
	*AssetResponse
	ChangedFields []FieldChangeInfo `json:"changed_fields,omitempty"`
}

// UuidPtrToStringPtr converts a uuid.UUID pointer to a string pointer
func UuidPtrToStringPtr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}
