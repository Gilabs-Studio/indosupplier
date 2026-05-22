package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type BudgetItemRequest struct {
	ChartOfAccountID string  `json:"chart_of_account_id" binding:"required,uuid"`
	Amount           float64 `json:"amount" binding:"required,gt=0"`
	Memo             string  `json:"memo"`
}

type CreateBudgetRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	StartDate   string              `json:"start_date" binding:"required"`
	EndDate     string              `json:"end_date" binding:"required"`
	Items       []BudgetItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateBudgetRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	StartDate   string              `json:"start_date" binding:"required"`
	EndDate     string              `json:"end_date" binding:"required"`
	Items       []BudgetItemRequest `json:"items" binding:"required,min=1,dive"`
}

type ListBudgetsRequest struct {
	Page      int                         `form:"page" binding:"omitempty,min=1"`
	PerPage   int                         `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search    string                      `form:"search"`
	Status    *financeModels.BudgetStatus `form:"status" binding:"omitempty,oneof=draft approved"`
	StartDate *string                     `form:"start_date"`
	EndDate   *string                     `form:"end_date"`
	SortBy    string                      `form:"sort_by"`
	SortDir   string                      `form:"sort_dir"`
}

type BudgetItemResponse struct {
	ID               string                  `json:"id"`
	ChartOfAccountID string                  `json:"chart_of_account_id"`
	ChartOfAccount   *ChartOfAccountResponse `json:"chart_of_account,omitempty"`
	Amount           float64                 `json:"amount"`
	ActualAmount     float64                 `json:"actual_amount"`
	Memo             string                  `json:"memo"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

type BudgetResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	StartDate   time.Time                  `json:"start_date"`
	EndDate     time.Time                  `json:"end_date"`
	TotalAmount float64                    `json:"total_amount"`
	Status      financeModels.BudgetStatus `json:"status"`
	ApprovedAt  *time.Time                 `json:"approved_at"`
	ApprovedBy  *string                    `json:"approved_by"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
	Items       []BudgetItemResponse       `json:"items,omitempty"`
}
