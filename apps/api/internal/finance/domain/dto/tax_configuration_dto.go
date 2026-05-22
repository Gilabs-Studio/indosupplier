package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type CreateTaxConfigurationRequest struct {
	CompanyID    string               `json:"company_id" binding:"required,uuid"`
	TaxCode      string               `json:"tax_code" binding:"required"`
	TaxName      string               `json:"tax_name" binding:"required"`
	TaxType      financeModels.TaxType `json:"tax_type" binding:"required,oneof=vat income_tax withholding_tax"`
	Rate         float64              `json:"rate" binding:"required,gte=0,lte=100"`
	IsInclusive  bool                 `json:"is_inclusive"`
	AccountID    string               `json:"account_id" binding:"required,uuid"`
	IsActive     *bool                `json:"is_active"`
}

type UpdateTaxConfigurationRequest struct {
	TaxCode      string               `json:"tax_code" binding:"required"`
	TaxName      string               `json:"tax_name" binding:"required"`
	TaxType      financeModels.TaxType `json:"tax_type" binding:"required,oneof=vat income_tax withholding_tax"`
	Rate         float64              `json:"rate" binding:"required,gte=0,lte=100"`
	IsInclusive  bool                 `json:"is_inclusive"`
	AccountID    string               `json:"account_id" binding:"required,uuid"`
}

type ToggleTaxConfigurationStatusRequest struct {
	CompanyID string `json:"company_id" binding:"required,uuid"`
	IsActive  bool   `json:"is_active"`
}

type ListTaxConfigurationsRequest struct {
	CompanyID string                 `form:"company_id" binding:"required,uuid"`
	TaxType   *financeModels.TaxType `form:"type" binding:"omitempty,oneof=vat income_tax withholding_tax"`
	IsActive  *bool                  `form:"is_active"`
	Page      int                    `form:"page" binding:"omitempty,min=1"`
	PerPage   int                    `form:"per_page" binding:"omitempty,min=1,max=100"`
}

type TaxConfigurationResponse struct {
	ID          string               `json:"id"`
	CompanyID   string               `json:"company_id"`
	TaxCode     string               `json:"tax_code"`
	TaxName     string               `json:"tax_name"`
	TaxType     financeModels.TaxType `json:"tax_type"`
	Rate        float64              `json:"rate"`
	IsInclusive bool                 `json:"is_inclusive"`
	AccountID   string               `json:"account_id"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}
