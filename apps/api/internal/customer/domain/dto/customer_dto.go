package dto

import (
	"time"

	coreDTO "github.com/gilabs/gims/api/internal/core/domain/dto"
	supplierDTO "github.com/gilabs/gims/api/internal/supplier/domain/dto"
)

// === Customer DTOs ===

// CreateCustomerRequest for creating a new customer
type CreateCustomerRequest struct {
	Code           string                      `json:"code" binding:"omitempty,max=50"`
	Name           string                      `json:"name" binding:"required,min=2,max=200"`
	CustomerTypeID *string                     `json:"customer_type_id" binding:"omitempty,uuid"`
	Address        string                      `json:"address" binding:"max=500"`
	ProvinceID     *string                     `json:"province_id" binding:"omitempty,uuid"`
	CityID         *string                     `json:"city_id" binding:"omitempty,uuid"`
	DistrictID     *string                     `json:"district_id" binding:"omitempty,uuid"`
	VillageID      *string                     `json:"village_id" binding:"omitempty,uuid"`
	VillageName    *string                     `json:"village_name" binding:"omitempty,max=255"`
	Email          string                      `json:"email" binding:"omitempty,email,max=100"`
	Website        string                      `json:"website" binding:"max=200"`
	NPWP           string                      `json:"npwp" binding:"max=30"`
	ContactPerson  string                      `json:"contact_person" binding:"max=100"`
	Notes          string                      `json:"notes" binding:"max=1000"`
	Latitude       *float64                    `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude      *float64                    `json:"longitude" binding:"omitempty,min=-180,max=180"`
	IsActive       *bool                       `json:"is_active"`
	BankAccounts   []CreateCustomerBankRequest `json:"bank_accounts"`
	// Sales defaults
	DefaultBusinessTypeID *string  `json:"default_business_type_id" binding:"omitempty,uuid"`
	DefaultAreaID         *string  `json:"default_area_id" binding:"omitempty,uuid"`
	DefaultSalesRepID     *string  `json:"default_sales_rep_id" binding:"omitempty,uuid"`
	DefaultPaymentTermsID *string  `json:"default_payment_terms_id" binding:"omitempty,uuid"`
	DefaultTaxRate        *float64 `json:"default_tax_rate" binding:"omitempty,min=0,max=100"`
	CreditLimit           *float64 `json:"credit_limit" binding:"omitempty,min=0"`
	CreditIsActive        *bool    `json:"credit_is_active"`
}

// UpdateCustomerRequest for updating an existing customer
type UpdateCustomerRequest struct {
	Code           *string  `json:"code" binding:"omitempty,min=2,max=50"`
	Name           *string  `json:"name" binding:"omitempty,min=2,max=200"`
	CustomerTypeID *string  `json:"customer_type_id" binding:"omitempty,uuid"`
	Address        *string  `json:"address" binding:"omitempty,max=500"`
	ProvinceID     *string  `json:"province_id" binding:"omitempty,uuid"`
	CityID         *string  `json:"city_id" binding:"omitempty,uuid"`
	DistrictID     *string  `json:"district_id" binding:"omitempty,uuid"`
	VillageID      *string  `json:"village_id" binding:"omitempty,uuid"`
	VillageName    *string  `json:"village_name" binding:"omitempty,max=255"`
	Email          *string  `json:"email" binding:"omitempty,email,max=100"`
	Website        *string  `json:"website" binding:"omitempty,max=200"`
	NPWP           *string  `json:"npwp" binding:"omitempty,max=30"`
	ContactPerson  *string  `json:"contact_person" binding:"omitempty,max=100"`
	Notes          *string  `json:"notes" binding:"omitempty,max=1000"`
	Latitude       *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude      *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`
	IsActive       *bool    `json:"is_active"`
	// Sales defaults
	DefaultBusinessTypeID *string  `json:"default_business_type_id" binding:"omitempty,uuid"`
	DefaultAreaID         *string  `json:"default_area_id" binding:"omitempty,uuid"`
	DefaultSalesRepID     *string  `json:"default_sales_rep_id" binding:"omitempty,uuid"`
	DefaultPaymentTermsID *string  `json:"default_payment_terms_id" binding:"omitempty,uuid"`
	DefaultTaxRate        *float64 `json:"default_tax_rate" binding:"omitempty,min=0,max=100"`
	CreditLimit           *float64 `json:"credit_limit" binding:"omitempty,min=0"`
	CreditIsActive        *bool    `json:"credit_is_active"`
}

// CustomerResponse is the response DTO for a customer
type CustomerResponse struct {
	ID             string                 `json:"id"`
	Code           string                 `json:"code"`
	Name           string                 `json:"name"`
	CustomerTypeID *string                `json:"customer_type_id"`
	CustomerType   *CustomerTypeResponse  `json:"customer_type,omitempty"`
	Address        string                 `json:"address"`
	ProvinceID     *string                `json:"province_id"`
	Province       *ProvinceResponse      `json:"province,omitempty"`
	CityID         *string                `json:"city_id"`
	City           *CityResponse          `json:"city,omitempty"`
	DistrictID     *string                `json:"district_id"`
	District       *DistrictResponse      `json:"district,omitempty"`
	VillageID      *string                `json:"village_id"`
	VillageName    *string                `json:"village_name,omitempty"`
	Village        *VillageResponse       `json:"village,omitempty"`
	Email          string                 `json:"email"`
	Website        string                 `json:"website"`
	NPWP           string                 `json:"npwp"`
	ContactPerson  string                 `json:"contact_person"`
	Notes          string                 `json:"notes"`
	Latitude       *float64               `json:"latitude"`
	Longitude      *float64               `json:"longitude"`
	CreatedBy      *string                `json:"created_by"`
	IsActive       bool                   `json:"is_active"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	BankAccounts   []CustomerBankResponse `json:"bank_accounts,omitempty"`
	// Sales defaults
	DefaultBusinessTypeID *string                  `json:"default_business_type_id"`
	DefaultBusinessType   *SalesDefaultOptionBrief `json:"default_business_type,omitempty"`
	DefaultAreaID         *string                  `json:"default_area_id"`
	DefaultArea           *SalesDefaultOptionBrief `json:"default_area,omitempty"`
	DefaultSalesRepID     *string                  `json:"default_sales_rep_id"`
	DefaultSalesRep       *SalesRepBrief           `json:"default_sales_rep,omitempty"`
	DefaultPaymentTermsID *string                  `json:"default_payment_terms_id"`
	DefaultPaymentTerms   *SalesDefaultOptionBrief `json:"default_payment_terms,omitempty"`
	DefaultTaxRate        *float64                 `json:"default_tax_rate"`
	CreditLimit           float64                  `json:"credit_limit"`
	CreditIsActive        bool                     `json:"credit_is_active"`
	// CRM enrichment
	ContactsCount int64 `json:"contacts_count"`
}

// SalesDefaultOptionBrief is a lightweight reference for business type / area / payment terms
type SalesDefaultOptionBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CustomerAreaFormOption is a lightweight area reference that includes province for auto-mapping.
type CustomerAreaFormOption struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Province string `json:"province,omitempty"`
}

// SalesRepBrief is a lightweight sales rep reference
type SalesRepBrief struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// CustomerFormDataResponse for form dropdown options
type CustomerFormDataResponse struct {
	CustomerTypes []CustomerTypeResponse    `json:"customer_types"`
	BusinessTypes []SalesDefaultOptionBrief `json:"business_types"`
	Areas         []CustomerAreaFormOption  `json:"areas"`
	SalesReps     []SalesRepBrief           `json:"sales_reps"`
	PaymentTerms  []PaymentTermsFormOption  `json:"payment_terms"`
}

// PaymentTermsFormOption is used in form dropdowns for payment terms
type PaymentTermsFormOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Days int    `json:"days"`
}

// === Customer Bank DTOs ===

// CreateCustomerBankRequest for adding a bank account to a customer
type CreateCustomerBankRequest struct {
	BankID        string `json:"bank_id" binding:"required,uuid"`
	CurrencyID    string `json:"currency_id" binding:"required,uuid"`
	AccountNumber string `json:"account_number" binding:"required,max=50"`
	AccountName   string `json:"account_name" binding:"required,max=100"`
	Branch        string `json:"branch" binding:"max=100"`
	IsPrimary     bool   `json:"is_primary"`
}

// UpdateCustomerBankRequest for updating a bank account
type UpdateCustomerBankRequest struct {
	BankID        string `json:"bank_id" binding:"omitempty,uuid"`
	CurrencyID    string `json:"currency_id" binding:"omitempty,uuid"`
	AccountNumber string `json:"account_number" binding:"omitempty,max=50"`
	AccountName   string `json:"account_name" binding:"omitempty,max=100"`
	Branch        string `json:"branch" binding:"max=100"`
	IsPrimary     *bool  `json:"is_primary"`
}

// CustomerBankResponse is the response DTO for a customer bank account
type CustomerBankResponse struct {
	ID            string                    `json:"id"`
	CustomerID    string                    `json:"customer_id"`
	BankID        string                    `json:"bank_id"`
	Bank          *supplierDTO.BankResponse `json:"bank,omitempty"`
	CurrencyID    *string                   `json:"currency_id"`
	Currency      *coreDTO.CurrencyResponse `json:"currency,omitempty"`
	AccountNumber string                    `json:"account_number"`
	AccountName   string                    `json:"account_name"`
	Branch        string                    `json:"branch"`
	IsPrimary     bool                      `json:"is_primary"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

// === Village Response (reuse geographic chain for nested display) ===

// VillageResponse for nested geographic display
type VillageResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	District *DistrictResponse `json:"district,omitempty"`
}

// DistrictResponse for nested geographic display
type DistrictResponse struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	City *CityResponse `json:"city,omitempty"`
}

// CityResponse for nested geographic display
type CityResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Province *ProvinceResponse `json:"province,omitempty"`
}

// ProvinceResponse for nested geographic display
type ProvinceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
