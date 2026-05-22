package dto

import (
	"time"

	coreDTO "github.com/gilabs/gims/api/internal/core/domain/dto"
)

// === Supplier DTOs ===

type CreateSupplierRequest struct {
	Name           string                      `json:"name" binding:"required,min=2,max=200"`
	SupplierTypeID string                      `json:"supplier_type_id" binding:"omitempty,uuid"`
	PaymentTermsID string                      `json:"payment_terms_id" binding:"omitempty,uuid"`
	BusinessUnitID string                      `json:"business_unit_id" binding:"omitempty,uuid"`
	Address        string                      `json:"address" binding:"max=500"`
	ProvinceID     string                      `json:"province_id" binding:"omitempty,uuid"`
	CityID         string                      `json:"city_id" binding:"omitempty,uuid"`
	DistrictID     string                      `json:"district_id" binding:"omitempty,uuid"`
	VillageID      string                      `json:"village_id" binding:"omitempty,uuid"`
	VillageName    string                      `json:"village_name" binding:"omitempty,max=255"`
	Email          string                      `json:"email" binding:"omitempty,email,max=100"`
	Website        string                      `json:"website" binding:"max=200"`
	NPWP           string                      `json:"npwp" binding:"max=30"`
	ContactPerson  string                      `json:"contact_person" binding:"max=100"`
	Notes          string                      `json:"notes" binding:"max=1000"`
	Latitude       *float64                    `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude      *float64                    `json:"longitude" binding:"omitempty,min=-180,max=180"`
	IsActive       *bool                       `json:"is_active"`
	Contacts       []CreateContactRequest      `json:"contacts"`
	BankAccounts   []CreateSupplierBankRequest `json:"bank_accounts"`
}

// UpdateSupplierRequest uses pointer types to distinguish between
// "field not sent" (nil) and "field sent as empty" (pointer to zero value)
// NOTE: For *string pointer fields, "omitempty" only skips validation when
// the pointer is nil. A non-nil pointer to "" does NOT trigger omitempty,
// so format validators like "email" would reject empty strings and block
// the entire request. Email format is validated in the usecase instead.
type UpdateSupplierRequest struct {
	Name           *string  `json:"name" binding:"omitempty,min=2,max=200"`
	SupplierTypeID *string  `json:"supplier_type_id"`
	PaymentTermsID *string  `json:"payment_terms_id"`
	BusinessUnitID *string  `json:"business_unit_id"`
	Address        *string  `json:"address" binding:"omitempty,max=500"`
	ProvinceID     *string  `json:"province_id"`
	CityID         *string  `json:"city_id"`
	DistrictID     *string  `json:"district_id"`
	VillageID      *string  `json:"village_id"`
	VillageName    *string  `json:"village_name,omitempty"`
	Email          *string  `json:"email" binding:"omitempty,max=100"`
	Website        *string  `json:"website" binding:"omitempty,max=200"`
	NPWP           *string  `json:"npwp" binding:"omitempty,max=30"`
	ContactPerson  *string  `json:"contact_person" binding:"omitempty,max=100"`
	Notes          *string  `json:"notes" binding:"omitempty,max=1000"`
	Latitude       *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude      *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`
	IsActive       *bool    `json:"is_active"`
}

type ApproveSupplierRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason" binding:"max=500"`
}

type SupplierResponse struct {
	ID             string                 `json:"id"`
	Code           string                 `json:"code"`
	Name           string                 `json:"name"`
	SupplierTypeID *string                `json:"supplier_type_id"`
	SupplierType   *SupplierTypeResponse  `json:"supplier_type,omitempty"`
	PaymentTermsID *string                `json:"payment_terms_id"`
	PaymentTerms   *PaymentTermsResponse  `json:"payment_terms,omitempty"`
	BusinessUnitID *string                `json:"business_unit_id"`
	BusinessUnit   *BusinessUnitResponse  `json:"business_unit,omitempty"`
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
	Status         string                 `json:"status"`
	IsApproved     bool                   `json:"is_approved"`
	CreatedBy      *string                `json:"created_by"`
	ApprovedBy     *string                `json:"approved_by"`
	ApprovedAt     *time.Time             `json:"approved_at"`
	IsActive       bool                   `json:"is_active"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Contacts       []ContactResponse      `json:"contacts,omitempty"`
	BankAccounts   []SupplierBankResponse `json:"bank_accounts,omitempty"`
}

// === Contact DTOs ===

type CreateContactRequest struct {
	ContactRoleID *string `json:"contact_role_id" binding:"omitempty,uuid"`
	Name          string  `json:"name" binding:"required,max=100"`
	Email         string  `json:"email" binding:"omitempty,email,max=100"`
	Phone         string  `json:"phone" binding:"required,max=30"`
	Notes         string  `json:"notes" binding:"max=1000"`
	IsPrimary     bool    `json:"is_primary"`
	IsActive      *bool   `json:"is_active"`
}

type UpdateContactRequest struct {
	ContactRoleID *string `json:"contact_role_id" binding:"omitempty,uuid"`
	Name          string  `json:"name" binding:"omitempty,max=100"`
	Email         string  `json:"email" binding:"omitempty,email,max=100"`
	Phone         string  `json:"phone" binding:"omitempty,max=30"`
	Notes         string  `json:"notes" binding:"max=1000"`
	IsPrimary     *bool   `json:"is_primary"`
	IsActive      *bool   `json:"is_active"`
}

type ContactRoleInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	BadgeColor string `json:"badge_color"`
}

type ContactResponse struct {
	ID            string           `json:"id"`
	SupplierID    string           `json:"supplier_id"`
	ContactRoleID *string          `json:"contact_role_id"`
	ContactRole   *ContactRoleInfo `json:"contact_role,omitempty"`
	Name          string           `json:"name"`
	Email         string           `json:"email"`
	Phone         string           `json:"phone"`
	Notes         string           `json:"notes"`
	IsPrimary     bool             `json:"is_primary"`
	IsActive      bool             `json:"is_active"`
	CreatedBy     *string          `json:"created_by"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// === Supplier Bank DTOs ===

type CreateSupplierBankRequest struct {
	BankID        string `json:"bank_id" binding:"required,uuid"`
	CurrencyID    string `json:"currency_id" binding:"required,uuid"`
	AccountNumber string `json:"account_number" binding:"required,max=50"`
	AccountName   string `json:"account_name" binding:"required,max=100"`
	Branch        string `json:"branch" binding:"max=100"`
	IsPrimary     bool   `json:"is_primary"`
}

type UpdateSupplierBankRequest struct {
	BankID        string `json:"bank_id" binding:"omitempty,uuid"`
	CurrencyID    string `json:"currency_id" binding:"omitempty,uuid"`
	AccountNumber string `json:"account_number" binding:"omitempty,max=50"`
	AccountName   string `json:"account_name" binding:"omitempty,max=100"`
	Branch        string `json:"branch" binding:"max=100"`
	IsPrimary     *bool  `json:"is_primary"`
}

type SupplierBankResponse struct {
	ID            string                    `json:"id"`
	SupplierID    string                    `json:"supplier_id"`
	BankID        string                    `json:"bank_id"`
	Bank          *BankResponse             `json:"bank,omitempty"`
	CurrencyID    *string                   `json:"currency_id"`
	Currency      *coreDTO.CurrencyResponse `json:"currency,omitempty"`
	AccountNumber string                    `json:"account_number"`
	AccountName   string                    `json:"account_name"`
	Branch        string                    `json:"branch"`
	IsPrimary     bool                      `json:"is_primary"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

// === Village Response (for nested display) ===

type VillageResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	District *DistrictResponse `json:"district,omitempty"`
}

type DistrictResponse struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	City *CityResponse `json:"city,omitempty"`
}

type CityResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Province *ProvinceResponse `json:"province,omitempty"`
}

type ProvinceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PaymentTermsResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Days int    `json:"days"`
}

type BusinessUnitResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
