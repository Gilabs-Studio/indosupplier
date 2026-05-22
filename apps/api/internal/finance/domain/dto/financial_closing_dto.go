package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type CreateFinancialClosingRequest struct {
	PeriodEndDate string `json:"period_end_date" binding:"required"`
	Notes         string `json:"notes"`
}

type ListFinancialClosingsRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	SortBy  string `form:"sort_by"`
	SortDir string `form:"sort_dir"`
}

type FinancialClosingResponse struct {
	ID            string                               `json:"id"`
	PeriodEndDate time.Time                            `json:"period_end_date"`
	Status        financeModels.FinancialClosingStatus `json:"status"`
	Notes         string                               `json:"notes"`
	ApprovedAt    *time.Time                           `json:"approved_at"`
	ApprovedBy    *string                              `json:"approved_by"`
	CreatedAt     time.Time                            `json:"created_at"`
	UpdatedAt     time.Time                            `json:"updated_at"`
}

type FinancialClosingAnalysisRow struct {
	AccountID      string  `json:"account_id"`
	AccountCode    string  `json:"account_code"`
	AccountName    string  `json:"account_name"`
	ClosingBalance float64 `json:"closing_balance"`
	OpeningBalance float64 `json:"opening_balance"`
	Difference     float64 `json:"difference"`
}

type FinancialClosingValidationResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type FinancialClosingSnapshotResponse struct {
	NetProfit               float64 `json:"net_profit"`
	RetainedEarningsBalance float64 `json:"retained_earnings_balance"`
	PeriodEndDate           string  `json:"period_end_date"`
	SnapshotJSON            string  `json:"snapshot_json"`
}

type FinancialClosingAnalysisResponse struct {
	Closing     FinancialClosingResponse        `json:"closing"`
	Rows        []FinancialClosingAnalysisRow  `json:"rows"`
	Validations []FinancialClosingValidationResult `json:"validations,omitempty"`
	Snapshot    *FinancialClosingSnapshotResponse   `json:"snapshot,omitempty"`
}

// YearEndCloseRequest represents the request to perform year-end closing.
type YearEndCloseRequest struct {
	FiscalYear int `json:"fiscal_year" binding:"required,min=2000,max=2100"`
}
