package dto

import "github.com/google/uuid"

// CreateLeadRequest defines the request body for creating a lead
type CreateLeadRequest struct {
	FirstName            string   `json:"first_name" binding:"required,min=1,max=100"`
	LastName             string   `json:"last_name" binding:"max=100"`
	CompanyName          string   `json:"company_name" binding:"max=200"`
	Email                string   `json:"email" binding:"omitempty,email,max=100"`
	Phone                string   `json:"phone" binding:"omitempty,max=20"`
	ContactRoleID        *string  `json:"contact_role_id" binding:"omitempty,uuid"`
	JobTitle             string   `json:"job_title" binding:"max=100"`
	Address              string   `json:"address"`
	City                 string   `json:"city" binding:"max=100"`
	Province             string   `json:"province" binding:"max=100"`
	ProvinceID           *string  `json:"province_id" binding:"omitempty,uuid"`
	CityID               *string  `json:"city_id" binding:"omitempty,uuid"`
	DistrictID           *string  `json:"district_id" binding:"omitempty,uuid"`
	VillageName          string   `json:"village_name" binding:"max=200"`
	LeadSourceID         *string  `json:"lead_source_id" binding:"omitempty,uuid"`
	LeadStatusID         *string  `json:"lead_status_id" binding:"omitempty,uuid"`
	EstimatedValue       float64  `json:"estimated_value"`
	Probability          int      `json:"probability" binding:"min=0,max=100"`
	Website              string   `json:"website" binding:"omitempty,max=255"`
	BankAccountID        *string  `json:"bank_account_id" binding:"omitempty,uuid"`
	BankAccountReference string   `json:"bank_account_reference" binding:"omitempty,max=255"`
	Latitude             *float64 `json:"latitude"`
	Longitude            *float64 `json:"longitude"`
	// BANT
	BudgetConfirmed bool    `json:"budget_confirmed"`
	BudgetAmount    float64 `json:"budget_amount"`
	AuthConfirmed   bool    `json:"auth_confirmed"`
	AuthPerson      string  `json:"auth_person" binding:"max=200"`
	NeedConfirmed   bool    `json:"need_confirmed"`
	NeedDescription string  `json:"need_description"`
	TimeConfirmed   bool    `json:"time_confirmed"`
	TimeExpected    *string `json:"time_expected"`
	// Assignment
	AssignedTo *string `json:"assigned_to" binding:"omitempty,uuid"`
	Notes      string  `json:"notes"`
	// Sales defaults for customer conversion
	BusinessTypeID *string `json:"business_type_id" binding:"omitempty,uuid"`
	AreaID         *string `json:"area_id" binding:"omitempty,uuid"`
	PaymentTermsID *string `json:"payment_terms_id" binding:"omitempty,uuid"`
}

// UpdateLeadRequest defines the request body for updating a lead
type UpdateLeadRequest struct {
	FirstName            *string  `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName             *string  `json:"last_name" binding:"omitempty,max=100"`
	CompanyName          *string  `json:"company_name" binding:"omitempty,max=200"`
	Email                *string  `json:"email" binding:"omitempty,email,max=100"`
	Phone                *string  `json:"phone" binding:"omitempty,max=20"`
	ContactRoleID        *string  `json:"contact_role_id" binding:"omitempty,uuid"`
	JobTitle             *string  `json:"job_title" binding:"omitempty,max=100"`
	Address              *string  `json:"address"`
	City                 *string  `json:"city" binding:"omitempty,max=100"`
	Province             *string  `json:"province" binding:"omitempty,max=100"`
	ProvinceID           *string  `json:"province_id" binding:"omitempty,uuid"`
	CityID               *string  `json:"city_id" binding:"omitempty,uuid"`
	DistrictID           *string  `json:"district_id" binding:"omitempty,uuid"`
	VillageName          *string  `json:"village_name" binding:"omitempty,max=200"`
	LeadSourceID         *string  `json:"lead_source_id" binding:"omitempty,uuid"`
	LeadStatusID         *string  `json:"lead_status_id" binding:"omitempty,uuid"`
	EstimatedValue       *float64 `json:"estimated_value"`
	Probability          *int     `json:"probability" binding:"omitempty,min=0,max=100"`
	Website              *string  `json:"website" binding:"omitempty,max=255"`
	BankAccountID        *string  `json:"bank_account_id" binding:"omitempty,uuid"`
	BankAccountReference *string  `json:"bank_account_reference" binding:"omitempty,max=255"`
	// BANT
	BudgetConfirmed *bool    `json:"budget_confirmed"`
	BudgetAmount    *float64 `json:"budget_amount"`
	AuthConfirmed   *bool    `json:"auth_confirmed"`
	AuthPerson      *string  `json:"auth_person" binding:"omitempty,max=200"`
	NeedConfirmed   *bool    `json:"need_confirmed"`
	NeedDescription *string  `json:"need_description"`
	TimeConfirmed   *bool    `json:"time_confirmed"`
	TimeExpected    *string  `json:"time_expected"`
	// Assignment
	AssignedTo *string `json:"assigned_to" binding:"omitempty,uuid"`
	Notes      *string `json:"notes"`
	// Additional fields for conversion completeness
	NPWP      *string  `json:"npwp" binding:"omitempty,max=30"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	// Sales defaults for customer conversion
	BusinessTypeID *string `json:"business_type_id" binding:"omitempty,uuid"`
	AreaID         *string `json:"area_id" binding:"omitempty,uuid"`
	PaymentTermsID *string `json:"payment_terms_id" binding:"omitempty,uuid"`
}

// ConvertLeadRequest defines the request body for converting a lead to a deal
type ConvertLeadRequest struct {
	PipelineStageID *string  `json:"pipeline_stage_id" binding:"omitempty,uuid"`
	DealTitle       string   `json:"deal_title" binding:"max=200"`
	DealValue       *float64 `json:"deal_value"`
	Notes           string   `json:"notes"`
}

// BulkUpsertLeadRequest defines the request body for bulk upserting leads from automation
type BulkUpsertLeadRequest struct {
	Leads []UpsertLeadItem `json:"leads" binding:"required,min=1,max=100,dive"`
}

// UpsertLeadItem represents a single lead in a bulk upsert operation.
// Uses email as the deduplication key: existing leads are updated, new ones are created.
type UpsertLeadItem struct {
	FirstName      string  `json:"first_name" binding:"required,min=1,max=100"`
	LastName       string  `json:"last_name" binding:"max=100"`
	CompanyName    string  `json:"company_name" binding:"max=200"`
	Email          string  `json:"email" binding:"omitempty,email,max=100"`
	Phone          string  `json:"phone" binding:"omitempty,max=20"`
	JobTitle       string  `json:"job_title" binding:"max=100"`
	Address        string  `json:"address"`
	City           string  `json:"city" binding:"max=100"`
	Province       string  `json:"province" binding:"max=100"`
	ProvinceID     *string `json:"province_id" binding:"omitempty,uuid"`
	CityID         *string `json:"city_id" binding:"omitempty,uuid"`
	DistrictID     *string `json:"district_id" binding:"omitempty,uuid"`
	VillageName    string  `json:"village_name" binding:"max=200"`
	LeadSourceID   *string `json:"lead_source_id" binding:"omitempty,uuid"`
	EstimatedValue float64 `json:"estimated_value"`
	Website        string  `json:"website" binding:"omitempty,max=255"`
	Notes          string  `json:"notes"`
	// External Scraping Fields
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Rating       *float64 `json:"rating"`
	RatingCount  *int     `json:"ratingCount"`
	Types        any      `json:"types"`
	OpeningHours any      `json:"openingHours"`
	ThumbnailURL string   `json:"thumbnailUrl"`
	CID          string   `json:"cid"`
	PlaceID      string   `json:"placeId"`
}

// BulkUpsertLeadResponse describes the outcome of a bulk upsert operation
type BulkUpsertLeadResponse struct {
	Created int            `json:"created"`
	Updated int            `json:"updated"`
	Errors  int            `json:"errors"`
	Items   []LeadResponse `json:"items"`
}

// LeadResponse defines the response body for a lead
type LeadResponse struct {
	ID             string               `json:"id"`
	Code           string               `json:"code"`
	FirstName      string               `json:"first_name"`
	LastName       string               `json:"last_name"`
	CompanyName    string               `json:"company_name"`
	Email          string               `json:"email"`
	Phone          string               `json:"phone"`
	ContactRoleID  *string              `json:"contact_role_id"`
	ContactRole    *LeadContactRoleInfo `json:"contact_role,omitempty"`
	JobTitle       string               `json:"job_title"`
	Address        string               `json:"address"`
	City           string               `json:"city"`
	Province       string               `json:"province"`
	ProvinceID     *string              `json:"province_id"`
	CityID         *string              `json:"city_id"`
	DistrictID     *string              `json:"district_id"`
	VillageName    string               `json:"village_name"`
	LeadSourceID   *string              `json:"lead_source_id"`
	LeadSource     *LeadSourceInfo      `json:"lead_source,omitempty"`
	LeadStatusID   *string              `json:"lead_status_id"`
	LeadStatus     *LeadStatusInfo      `json:"lead_status,omitempty"`
	LeadScore      int                  `json:"lead_score"`
	Probability    int                  `json:"probability"`
	EstimatedValue float64              `json:"estimated_value"`
	// BANT
	BudgetConfirmed bool    `json:"budget_confirmed"`
	BudgetAmount    float64 `json:"budget_amount"`
	AuthConfirmed   bool    `json:"auth_confirmed"`
	AuthPerson      string  `json:"auth_person"`
	NeedConfirmed   bool    `json:"need_confirmed"`
	NeedDescription string  `json:"need_description"`
	TimeConfirmed   bool    `json:"time_confirmed"`
	TimeExpected    *string `json:"time_expected"`
	// Assignment
	AssignedTo       *string           `json:"assigned_to"`
	AssignedEmployee *LeadEmployeeInfo `json:"assigned_employee,omitempty"`
	// Conversion
	CustomerID  *string           `json:"customer_id"`
	Customer    *LeadCustomerInfo `json:"customer,omitempty"`
	ContactID   *string           `json:"contact_id"`
	DealID      *string           `json:"deal_id"`
	Deal        *LeadDealInfo     `json:"deal,omitempty"`
	ConvertedAt *string           `json:"converted_at"`
	ConvertedBy *string           `json:"converted_by"`
	// Metadata
	Notes     string  `json:"notes"`
	NPWP      string  `json:"npwp"`
	CreatedBy *string `json:"created_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`

	// External Scraping Fields
	Latitude             *float64 `json:"latitude"`
	Longitude            *float64 `json:"longitude"`
	Rating               *float64 `json:"rating"`
	RatingCount          *int     `json:"rating_count"`
	Types                string   `json:"types"`
	OpeningHours         string   `json:"opening_hours"`
	ThumbnailURL         string   `json:"thumbnail_url"`
	CID                  string   `json:"cid"`
	PlaceID              string   `json:"place_id"`
	Website              string   `json:"website"`
	BankAccountID        *string  `json:"bank_account_id"`
	BankAccountReference string   `json:"bank_account_reference"`
	// Sales defaults
	BusinessTypeID *string               `json:"business_type_id"`
	BusinessType   *LeadBusinessTypeInfo `json:"business_type,omitempty"`
	AreaID         *string               `json:"area_id"`
	Area           *LeadAreaInfo         `json:"area,omitempty"`
	PaymentTermsID *string               `json:"payment_terms_id"`
	// Activities log (populated on detail)
	Activities []ActivityResponse `json:"activities,omitempty"`
	// Tasks (populated on detail)
	Tasks []TaskSummaryResponse `json:"tasks,omitempty"`
	// Product items (populated on detail)
	ProductItems []LeadProductItemResponse `json:"product_items,omitempty"`
}

// LeadProductItemResponse is the DTO for a lead product item
type LeadProductItemResponse struct {
	ID                  string  `json:"id"`
	LeadID              string  `json:"lead_id"`
	ProductID           *string `json:"product_id"`
	ProductName         string  `json:"product_name"`
	ProductSKU          string  `json:"product_sku"`
	InterestLevel       int     `json:"interest_level"`
	Quantity            int     `json:"quantity"`
	UnitPrice           float64 `json:"unit_price"`
	Notes               string  `json:"notes"`
	SourceVisitReportID *string `json:"source_visit_report_id"`
	LastSurveyAnswers   *string `json:"last_survey_answers"`
	IsDeleted           bool    `json:"is_deleted"`
	CreatedAt           string  `json:"created_at"`
}

// LeadSourceInfo is a compact representation of lead source
type LeadSourceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// LeadStatusInfo is a compact representation of lead status
type LeadStatusInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Color       string `json:"color"`
	Score       int    `json:"score"`
	IsConverted bool   `json:"is_converted"`
}

// LeadEmployeeInfo is a compact representation of an employee
type LeadEmployeeInfo struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// LeadContactRoleInfo is a compact representation of contact role
type LeadContactRoleInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	BadgeColor string `json:"badge_color"`
}

// LeadCustomerInfo is a compact representation of a customer
type LeadCustomerInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// LeadDealInfo is a compact representation of a deal for lead response
type LeadDealInfo struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Stage  string `json:"stage"`
}

// LeadBusinessTypeInfo is a compact representation of a business type
type LeadBusinessTypeInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LeadAreaInfo is a compact representation of an area
type LeadAreaInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LeadFormDataResponse holds all options required by the lead form
type LeadFormDataResponse struct {
	Employees      []LeadEmployeeOption      `json:"employees"`
	LeadSources    []LeadSourceOption        `json:"lead_sources"`
	LeadStatuses   []LeadStatusOption        `json:"lead_statuses"`
	PipelineStages []LeadPipelineStageOption `json:"pipeline_stages"`
	BusinessTypes  []LeadBusinessTypeOption  `json:"business_types"`
	Areas          []LeadAreaOption          `json:"areas"`
	PaymentTerms   []LeadPaymentTermsOption  `json:"payment_terms"`
}

// LeadEmployeeOption for employee dropdown
type LeadEmployeeOption struct {
	ID           uuid.UUID `json:"id"`
	EmployeeCode string    `json:"employee_code"`
	Name         string    `json:"name"`
}

// LeadSourceOption for lead source dropdown
type LeadSourceOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// LeadStatusOption for lead status dropdown
type LeadStatusOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Color       string `json:"color"`
	IsDefault   bool   `json:"is_default"`
	IsConverted bool   `json:"is_converted"`
}

// LeadPipelineStageOption for pipeline stage dropdown in conversion
type LeadPipelineStageOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Order       int    `json:"order"`
	Probability int    `json:"probability"`
}

// LeadBusinessTypeOption for business type dropdown
type LeadBusinessTypeOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LeadAreaOption for area dropdown
type LeadAreaOption struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Province string `json:"province,omitempty"`
}

// LeadPaymentTermsOption for payment terms dropdown
type LeadPaymentTermsOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Days int    `json:"days"`
}
