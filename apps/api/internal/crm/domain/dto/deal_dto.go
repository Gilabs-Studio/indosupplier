package dto

import "github.com/google/uuid"

// CreateDealRequest defines the request body for creating a deal
type CreateDealRequest struct {
	Title                string  `json:"title" binding:"required,min=2,max=200"`
	Description          string  `json:"description"`
	PipelineStageID      string  `json:"pipeline_stage_id" binding:"required,uuid"`
	Value                float64 `json:"value"`
	ExpectedCloseDate    *string `json:"expected_close_date"`
	CustomerID           *string `json:"customer_id" binding:"omitempty,uuid"`
	ContactID            *string `json:"contact_id" binding:"omitempty,uuid"`
	AssignedTo           *string `json:"assigned_to" binding:"omitempty,uuid"`
	LeadID               *string `json:"lead_id" binding:"omitempty,uuid"`
	// BANT
	BudgetConfirmed bool                       `json:"budget_confirmed"`
	BudgetAmount    float64                    `json:"budget_amount"`
	AuthConfirmed   bool                       `json:"auth_confirmed"`
	AuthPerson      string                     `json:"auth_person" binding:"max=200"`
	NeedConfirmed   bool                       `json:"need_confirmed"`
	NeedDescription string                     `json:"need_description"`
	TimeConfirmed   bool                       `json:"time_confirmed"`
	Notes           string                     `json:"notes"`
	Items           []CreateDealProductItemDTO `json:"items"`
}

// CreateDealProductItemDTO defines a product item in the deal creation request
type CreateDealProductItemDTO struct {
	ProductID       *string `json:"product_id" binding:"omitempty,uuid"`
	ProductName     string  `json:"product_name" binding:"required,max=200"`
	ProductSKU      string  `json:"product_sku" binding:"max=50"`
	UnitPrice       float64 `json:"unit_price"`
	Quantity        int     `json:"quantity" binding:"min=1"`
	DiscountPercent float64 `json:"discount_percent"`
	DiscountAmount  float64 `json:"discount_amount"`
	Notes           string  `json:"notes"`
}

// UpdateDealRequest defines the request body for updating a deal
type UpdateDealRequest struct {
	Title                *string  `json:"title" binding:"omitempty,min=2,max=200"`
	Description          *string  `json:"description"`
	PipelineStageID      *string  `json:"pipeline_stage_id" binding:"omitempty,uuid"`
	Value                *float64 `json:"value"`
	ExpectedCloseDate    *string  `json:"expected_close_date"`
	CustomerID           *string  `json:"customer_id" binding:"omitempty,uuid"`
	ContactID            *string  `json:"contact_id" binding:"omitempty,uuid"`
	AssignedTo           *string  `json:"assigned_to" binding:"omitempty,uuid"`
	LeadID               *string  `json:"lead_id" binding:"omitempty,uuid"`
	// BANT
	BudgetConfirmed *bool                       `json:"budget_confirmed"`
	BudgetAmount    *float64                    `json:"budget_amount"`
	AuthConfirmed   *bool                       `json:"auth_confirmed"`
	AuthPerson      *string                     `json:"auth_person" binding:"omitempty,max=200"`
	NeedConfirmed   *bool                       `json:"need_confirmed"`
	NeedDescription *string                     `json:"need_description"`
	TimeConfirmed   *bool                       `json:"time_confirmed"`
	Notes           *string                     `json:"notes"`
	Items           *[]CreateDealProductItemDTO `json:"items"`
}

// MoveDealStageRequest defines the request body for moving a deal to a different stage
type MoveDealStageRequest struct {
	ToStageID          string `json:"to_stage_id" binding:"required,uuid"`
	Reason             string `json:"reason" binding:"required,min=2"`
	Notes              string `json:"notes"`
	CloseReason        string `json:"close_reason"`
	ConvertToQuotation bool   `json:"convert_to_quotation"`
}

// MoveDealStageResponse extends DealResponse with optional conversion result
type MoveDealStageResponse struct {
	Deal       DealResponse                `json:"deal"`
	Conversion *ConvertToQuotationResponse `json:"conversion,omitempty"`
}

// DealResponse defines the response body for a deal
type DealResponse struct {
	ID                   string                 `json:"id"`
	Code                 string                 `json:"code"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Status               string                 `json:"status"`
	PipelineStageID      string                 `json:"pipeline_stage_id"`
	PipelineStage        *DealPipelineStageInfo `json:"pipeline_stage,omitempty"`
	Value                float64                `json:"value"`
	Probability          int                    `json:"probability"`
	ExpectedCloseDate    *string                `json:"expected_close_date"`
	ActualCloseDate      *string                `json:"actual_close_date"`
	CloseReason          string                 `json:"close_reason"`
	CustomerID           *string                `json:"customer_id"`
	Customer             *DealCustomerInfo      `json:"customer,omitempty"`
	ContactID            *string                `json:"contact_id"`
	Contact              *DealContactInfo       `json:"contact,omitempty"`
	AssignedTo           *string                `json:"assigned_to"`
	AssignedEmployee     *DealEmployeeInfo      `json:"assigned_employee,omitempty"`
	LeadID               *string                `json:"lead_id"`
	Lead                 *DealLeadInfo          `json:"lead,omitempty"`
	// BANT
	BudgetConfirmed bool    `json:"budget_confirmed"`
	BudgetAmount    float64 `json:"budget_amount"`
	AuthConfirmed   bool    `json:"auth_confirmed"`
	AuthPerson      string  `json:"auth_person"`
	NeedConfirmed   bool    `json:"need_confirmed"`
	NeedDescription string  `json:"need_description"`
	TimeConfirmed   bool    `json:"time_confirmed"`
	// Items
	Items []DealProductItemResponse `json:"items"`
	// Conversion tracking
	ConvertedToQuotationID *string `json:"converted_to_quotation_id"`
	ConvertedAt            *string `json:"converted_at"`
	// Tasks (populated on detail)
	Tasks []TaskSummaryResponse `json:"tasks,omitempty"`
	// Metadata
	Notes     string  `json:"notes"`
	CreatedBy *string `json:"created_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// DealProductItemResponse defines the response for a deal product item
type DealProductItemResponse struct {
	ID              string  `json:"id"`
	DealID          string  `json:"deal_id"`
	ProductID       *string `json:"product_id"`
	ProductName     string  `json:"product_name"`
	ProductSKU      string  `json:"product_sku"`
	UnitPrice       float64 `json:"unit_price"`
	Quantity        int     `json:"quantity"`
	DiscountPercent float64 `json:"discount_percent"`
	DiscountAmount  float64 `json:"discount_amount"`
	Subtotal        float64 `json:"subtotal"`
	Notes           string  `json:"notes"`
	InterestLevel   int     `json:"interest_level"`
	IsDeleted       bool    `json:"is_deleted"`
}

// DealHistoryResponse defines the response for a deal stage transition
type DealHistoryResponse struct {
	ID              string                 `json:"id"`
	DealID          string                 `json:"deal_id"`
	FromStageID     *string                `json:"from_stage_id"`
	FromStage       *DealPipelineStageInfo `json:"from_stage,omitempty"`
	FromStageName   string                 `json:"from_stage_name"`
	ToStageID       string                 `json:"to_stage_id"`
	ToStage         *DealPipelineStageInfo `json:"to_stage,omitempty"`
	ToStageName     string                 `json:"to_stage_name"`
	FromProbability int                    `json:"from_probability"`
	ToProbability   int                    `json:"to_probability"`
	DaysInPrevStage int                    `json:"days_in_prev_stage"`
	ChangedBy       *string                `json:"changed_by"`
	ChangedByUser   *DealEmployeeInfo      `json:"changed_by_user,omitempty"`
	ChangedAt       string                 `json:"changed_at"`
	Reason          string                 `json:"reason"`
	Notes           string                 `json:"notes"`
}

// Compact info types for deal relationships
type DealPipelineStageInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Color       string `json:"color"`
	Order       int    `json:"order"`
	Probability int    `json:"probability"`
	IsWon       bool   `json:"is_won"`
	IsLost      bool   `json:"is_lost"`
}

type DealCustomerInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type DealContactInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type DealEmployeeInfo struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

type DealLeadInfo struct {
	ID          string   `json:"id"`
	Code        string   `json:"code"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	CompanyName string   `json:"company_name"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	Province    string   `json:"province"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

// DealFormDataResponse holds all options required by the deal form
type DealFormDataResponse struct {
	Employees      []DealEmployeeOption      `json:"employees"`
	Customers      []DealCustomerOption      `json:"customers"`
	Contacts       []DealContactOption       `json:"contacts"`
	PipelineStages []DealPipelineStageOption `json:"pipeline_stages"`
	Products       []DealProductOption       `json:"products"`
	Leads          []DealLeadOption          `json:"leads"`
}

type DealEmployeeOption struct {
	ID           uuid.UUID `json:"id"`
	EmployeeCode string    `json:"employee_code"`
	Name         string    `json:"name"`
}

type DealCustomerOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type DealContactOption struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	CustomerID string `json:"customer_id"`
}

type DealPipelineStageOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Color       string `json:"color"`
	Order       int    `json:"order"`
	Probability int    `json:"probability"`
	IsWon       bool   `json:"is_won"`
	IsLost      bool   `json:"is_lost"`
	IsActive    bool   `json:"is_active"`
}

type DealProductOption struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	SellingPrice float64 `json:"selling_price"`
}

type DealLeadOption struct {
	ID                      string `json:"id"`
	Code                    string `json:"code"`
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name"`
	CompanyName             string `json:"company_name"`
	IsConverted             bool   `json:"is_converted"`
	IsQualifiedForConversion bool   `json:"is_qualified_for_conversion"`
}

// DealPipelineSummaryResponse holds pipeline summary for the frontend
type DealPipelineSummaryResponse struct {
	TotalDeals int64                      `json:"total_deals"`
	TotalValue float64                    `json:"total_value"`
	OpenDeals  int64                      `json:"open_deals"`
	OpenValue  float64                    `json:"open_value"`
	WonDeals   int64                      `json:"won_deals"`
	WonValue   float64                    `json:"won_value"`
	LostDeals  int64                      `json:"lost_deals"`
	LostValue  float64                    `json:"lost_value"`
	ByStage    []DealStageSummaryResponse `json:"by_stage"`
}

type DealStageSummaryResponse struct {
	StageID    string  `json:"stage_id"`
	StageName  string  `json:"stage_name"`
	StageColor string  `json:"stage_color"`
	DealCount  int64   `json:"deal_count"`
	TotalValue float64 `json:"total_value"`
}

// DealForecastResponse holds deal forecast for the frontend
type DealForecastResponse struct {
	TotalWeightedValue float64                     `json:"total_weighted_value"`
	TotalDeals         int64                       `json:"total_deals"`
	ByStage            []DealStageForecastResponse `json:"by_stage"`
}

type DealStageForecastResponse struct {
	StageID       string  `json:"stage_id"`
	StageName     string  `json:"stage_name"`
	DealCount     int64   `json:"deal_count"`
	TotalValue    float64 `json:"total_value"`
	Probability   int     `json:"probability"`
	WeightedValue float64 `json:"weighted_value"`
}

// --- Sprint 21: Deal → Sales Quotation Conversion + Stock Check ---

// ConvertToQuotationRequest defines optional overrides when converting a deal to a quotation
type ConvertToQuotationRequest struct {
	PaymentTermsID *string `json:"payment_terms_id" binding:"omitempty,uuid"`
	BusinessUnitID *string `json:"business_unit_id" binding:"omitempty,uuid"`
	BusinessTypeID *string `json:"business_type_id" binding:"omitempty,uuid"`
	Notes          string  `json:"notes"`
}

// ConvertToQuotationResponse is returned after successful conversion
type ConvertToQuotationResponse struct {
	DealID        string `json:"deal_id"`
	QuotationID   string `json:"quotation_id"`
	QuotationCode string `json:"quotation_code"`
}

// StockCheckItemResponse represents stock availability for a single product item
type StockCheckItemResponse struct {
	ProductID         string  `json:"product_id"`
	ProductName       string  `json:"product_name"`
	RequestedQuantity int     `json:"requested_quantity"`
	AvailableStock    float64 `json:"available_stock"`
	ReservedStock     float64 `json:"reserved_stock"`
	IsSufficient      bool    `json:"is_sufficient"`
}

// StockCheckResponse is the full stock availability check response for a deal
type StockCheckResponse struct {
	DealID        string                   `json:"deal_id"`
	Items         []StockCheckItemResponse `json:"items"`
	AllSufficient bool                     `json:"all_sufficient"`
}
