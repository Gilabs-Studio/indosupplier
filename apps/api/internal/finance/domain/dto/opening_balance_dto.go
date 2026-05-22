package dto

import "time"

type OpeningBalanceLineInput struct {
	AccountID      string   `json:"account_id" binding:"required,uuid"`
	DebitAmount    float64  `json:"debit_amount" binding:"gte=0"`
	CreditAmount   float64  `json:"credit_amount" binding:"gte=0"`
	Description    string   `json:"description"`
	ProductID      *string  `json:"product_id" binding:"omitempty,uuid"`
	ProductQty     *float64 `json:"product_qty" binding:"omitempty,gte=0"`
	ProductAvgCost *float64 `json:"product_avg_cost" binding:"omitempty,gte=0"`
}

type UpsertOpeningBalanceRequest struct {
	CompanyID    string                    `json:"company_id" binding:"required,uuid"`
	FiscalYearID string                    `json:"fiscal_year_id" binding:"required,uuid"`
	Lines        []OpeningBalanceLineInput `json:"lines" binding:"required,min=1"`
}

type ValidateOpeningBalanceRequest struct {
	CompanyID    string `json:"company_id" binding:"required,uuid"`
	FiscalYearID string `json:"fiscal_year_id" binding:"required,uuid"`
}

type PostOpeningBalanceRequest struct {
	CompanyID    string `json:"company_id" binding:"required,uuid"`
	FiscalYearID string `json:"fiscal_year_id" binding:"required,uuid"`
}

type OpeningBalanceValidationResponse struct {
	TotalDebit      float64  `json:"total_debit"`
	TotalCredit     float64  `json:"total_credit"`
	Difference      float64  `json:"difference"`
	IsBalanced      bool     `json:"is_balanced"`
	HasExistingTxns bool     `json:"has_existing_txns"`
	Warnings        []string `json:"warnings"`
}

type OpeningBalanceLineResponse struct {
	ID             string    `json:"id"`
	AccountID      string    `json:"account_id"`
	DebitAmount    float64   `json:"debit_amount"`
	CreditAmount   float64   `json:"credit_amount"`
	Description    string    `json:"description,omitempty"`
	ProductID      *string   `json:"product_id,omitempty"`
	ProductQty     *float64  `json:"product_qty,omitempty"`
	ProductAvgCost *float64  `json:"product_avg_cost,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OpeningBalanceSummaryResponse struct {
	CompanyID       string  `json:"company_id"`
	FiscalYearID    string  `json:"fiscal_year_id"`
	TotalDebit      float64 `json:"total_debit"`
	TotalCredit     float64 `json:"total_credit"`
	Difference      float64 `json:"difference"`
	IsBalanced      bool    `json:"is_balanced"`
	IsPosted        bool    `json:"is_posted"`
	PostedJournalID *string `json:"posted_journal_id,omitempty"`
	TotalLines      int     `json:"total_lines"`
	InventoryLines  int     `json:"inventory_lines"`
}

type OpeningBalanceSimulationLineResponse struct {
	AccountID    string  `json:"account_id"`
	AccountCode  string  `json:"account_code"`
	AccountName  string  `json:"account_name"`
	DebitAmount  float64 `json:"debit_amount"`
	CreditAmount float64 `json:"credit_amount"`
	Status       string  `json:"status"`
	Action       string  `json:"action"`
}

type OpeningBalanceSimulationResponse struct {
	CompanyID       string                                 `json:"company_id"`
	FiscalYearID    string                                 `json:"fiscal_year_id"`
	TotalDebit      float64                                `json:"total_debit"`
	TotalCredit     float64                                `json:"total_credit"`
	Difference      float64                                `json:"difference"`
	IsBalanced      bool                                   `json:"is_balanced"`
	ValuationStatus string                                 `json:"valuation_status"`
	Recommendation  string                                 `json:"recommendation"`
	Lines           []OpeningBalanceSimulationLineResponse `json:"lines"`
}

type PostOpeningBalanceResponse struct {
	JournalID     string `json:"journal_id"`
	JournalStatus string `json:"journal_status"`
	JournalType   string `json:"journal_type"`
	PostedAt      string `json:"posted_at"`
}
