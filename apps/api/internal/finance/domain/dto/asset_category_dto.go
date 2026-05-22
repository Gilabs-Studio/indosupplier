package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type CreateAssetCategoryRequest struct {
	Name string                          `json:"name" binding:"required"`
	Type financeModels.AssetCategoryType `json:"type" binding:"omitempty,oneof=FIXED CURRENT INTANGIBLE OTHER"`

	DepreciationMethod financeModels.DepreciationMethod `json:"depreciation_method" binding:"required,oneof=SL DB NONE"`
	UsefulLifeMonths   int                              `json:"useful_life_months" binding:"omitempty,gte=0"`
	DepreciationRate   float64                          `json:"depreciation_rate" binding:"omitempty,gte=0"`
	IsDepreciable      *bool                            `json:"is_depreciable"`

	AssetAccountID                   string  `json:"asset_account_id" binding:"required,uuid"`
	AccumulatedDepreciationAccountID string  `json:"accumulated_depreciation_account_id" binding:"required,uuid"`
	DepreciationExpenseAccountID     string  `json:"depreciation_expense_account_id" binding:"required,uuid"`
	DisposalGainAccountID            *string `json:"disposal_gain_account_id" binding:"omitempty,uuid"`
	DisposalLossAccountID            *string `json:"disposal_loss_account_id" binding:"omitempty,uuid"`

	IsActive *bool `json:"is_active"`
}

type UpdateAssetCategoryRequest struct {
	Name string                          `json:"name" binding:"required"`
	Type financeModels.AssetCategoryType `json:"type" binding:"omitempty,oneof=FIXED CURRENT INTANGIBLE OTHER"`

	DepreciationMethod financeModels.DepreciationMethod `json:"depreciation_method" binding:"required,oneof=SL DB NONE"`
	UsefulLifeMonths   int                              `json:"useful_life_months" binding:"omitempty,gte=0"`
	DepreciationRate   float64                          `json:"depreciation_rate" binding:"omitempty,gte=0"`
	IsDepreciable      *bool                            `json:"is_depreciable"`

	AssetAccountID                   string  `json:"asset_account_id" binding:"required,uuid"`
	AccumulatedDepreciationAccountID string  `json:"accumulated_depreciation_account_id" binding:"required,uuid"`
	DepreciationExpenseAccountID     string  `json:"depreciation_expense_account_id" binding:"required,uuid"`
	DisposalGainAccountID            *string `json:"disposal_gain_account_id" binding:"omitempty,uuid"`
	DisposalLossAccountID            *string `json:"disposal_loss_account_id" binding:"omitempty,uuid"`

	IsActive *bool `json:"is_active"`
}

type ListAssetCategoriesRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search"`
	SortBy  string `form:"sort_by"`
	SortDir string `form:"sort_dir"`
}

type AssetCategoryResponse struct {
	ID   string                          `json:"id"`
	Name string                          `json:"name"`
	Type financeModels.AssetCategoryType `json:"type"`

	DepreciationMethod financeModels.DepreciationMethod `json:"depreciation_method"`
	UsefulLifeMonths   int                              `json:"useful_life_months"`
	DepreciationRate   float64                          `json:"depreciation_rate"`
	IsDepreciable      bool                             `json:"is_depreciable"`

	AssetAccountID                   string  `json:"asset_account_id"`
	AccumulatedDepreciationAccountID string  `json:"accumulated_depreciation_account_id"`
	DepreciationExpenseAccountID     string  `json:"depreciation_expense_account_id"`
	DisposalGainAccountID            *string `json:"disposal_gain_account_id,omitempty"`
	DisposalLossAccountID            *string `json:"disposal_loss_account_id,omitempty"`

	IsActive bool `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
