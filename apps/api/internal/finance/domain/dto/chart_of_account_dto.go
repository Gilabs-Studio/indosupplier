package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type CreateChartOfAccountRequest struct {
	Code           string                    `json:"code"`
	Name           string                    `json:"name" binding:"required"`
	Type           financeModels.AccountType `json:"type" binding:"required,oneof=ASSET LIABILITY EQUITY REVENUE EXPENSE CASH_BANK CURRENT_ASSET FIXED_ASSET TRADE_PAYABLE COST_OF_GOODS_SOLD SALARY_WAGES OPERATIONAL"`
	ParentID       *string                   `json:"parent_id" binding:"omitempty,uuid"`
	IsActive       *bool                     `json:"is_active"`
	IsPostable     *bool                     `json:"is_postable"`
}

type UpdateChartOfAccountRequest struct {
	Code           string                    `json:"code" binding:"required"`
	Name           string                    `json:"name" binding:"required"`
	Type           financeModels.AccountType `json:"type" binding:"required,oneof=ASSET LIABILITY EQUITY REVENUE EXPENSE CASH_BANK CURRENT_ASSET FIXED_ASSET TRADE_PAYABLE COST_OF_GOODS_SOLD SALARY_WAGES OPERATIONAL"`
	ParentID       *string                   `json:"parent_id" binding:"omitempty,uuid"`
	IsActive       *bool                     `json:"is_active"`
	IsPostable     *bool                     `json:"is_postable"`
}

type ListChartOfAccountsRequest struct {
	Page     int                        `form:"page" binding:"omitempty,min=1"`
	PerPage  int                        `form:"per_page" binding:"omitempty,min=1,max=1000"`
	Search   string                     `form:"search"`
	Type     *financeModels.AccountType `form:"type" binding:"omitempty,oneof=ASSET LIABILITY EQUITY REVENUE EXPENSE CASH_BANK CURRENT_ASSET FIXED_ASSET TRADE_PAYABLE COST_OF_GOODS_SOLD SALARY_WAGES OPERATIONAL"`
	ParentID *string                    `form:"parent_id" binding:"omitempty,uuid"`
	IsActive *bool                      `form:"is_active"`
	SortBy   string                     `form:"sort_by"`
	SortDir  string                     `form:"sort_dir"`
}

type ChartOfAccountResponse struct {
	ID             string                    `json:"id"`
	Code           string                    `json:"code"`
	Name           string                    `json:"name"`
	Type           financeModels.AccountType `json:"type"`
	ParentID       *string                   `json:"parent_id"`
	IsActive       bool                      `json:"is_active"`
	IsPostable     bool                      `json:"is_postable"`
	IsProtected    bool                      `json:"is_protected"`
	Level          int                       `json:"level"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

type ChartOfAccountTreeNode struct {
	ID             string                    `json:"id"`
	Code           string                    `json:"code"`
	Name           string                    `json:"name"`
	Type           financeModels.AccountType `json:"type"`
	ParentID       *string                   `json:"parent_id"`
	IsActive       bool                      `json:"is_active"`
	IsPostable     bool                      `json:"is_postable"`
	IsProtected    bool                      `json:"is_protected"`
	Level          int                       `json:"level"`
	Children       []ChartOfAccountTreeNode  `json:"children"`
}

type AccountBalanceResponse struct {
	ChartOfAccountID string  `json:"chart_of_account_id"`
	DebitTotal       float64 `json:"debit_total"`
	CreditTotal      float64 `json:"credit_total"`
	Balance          float64 `json:"balance"`
}
