package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type CreateFiscalYearRequest struct {
	CompanyID string `json:"company_id" binding:"required,uuid"`
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type UpdateFiscalYearRequest struct {
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type ListFiscalYearsRequest struct {
	CompanyID string                         `form:"company_id" binding:"required,uuid"`
	Status    *financeModels.FiscalYearStatus `form:"status" binding:"omitempty,oneof=draft active locked"`
	Page      int                            `form:"page" binding:"omitempty,min=1"`
	PerPage   int                            `form:"per_page" binding:"omitempty,min=1,max=100"`
}

type FiscalYearResponse struct {
	ID        string                       `json:"id"`
	CompanyID string                       `json:"company_id"`
	Name      string                       `json:"name"`
	StartDate time.Time                    `json:"start_date"`
	EndDate   time.Time                    `json:"end_date"`
	Status    financeModels.FiscalYearStatus `json:"status"`
	CreatedBy *string                      `json:"created_by,omitempty"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
}
