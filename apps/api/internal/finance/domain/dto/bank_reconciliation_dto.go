package dto

import financeModels "github.com/gilabs/gims/api/internal/finance/data/models"

type ImportBankStatementRequest struct {
	CompanyID        string  `json:"company_id" form:"company_id" binding:"required,uuid"`
	BankAccountID    string  `json:"bank_account_id" form:"bank_account_id" binding:"required,uuid"`
	StatementDate    string  `json:"statement_date" form:"statement_date" binding:"required"`
	StatementBalance *float64 `json:"statement_balance" form:"statement_balance" binding:"omitempty"`
	FileFormat       string  `json:"file_format" form:"file_format" binding:"required,oneof=csv mt940 ofx"`
}

type ListBankReconciliationsRequest struct {
	Page          int                                     `form:"page" binding:"omitempty,min=1"`
	PerPage       int                                     `form:"per_page" binding:"omitempty,min=1,max=100"`
	CompanyID     string                                  `form:"company_id"`
	BankAccountID string                                  `form:"bank_account_id"`
	Status        *financeModels.BankReconciliationStatus `form:"status"`
	DateFrom      *string                                 `form:"date_from"`
	DateTo        *string                                 `form:"date_to"`
	SortBy        string                                  `form:"sort_by"`
	SortDir       string                                  `form:"sort_dir"`
}

type MatchBankStatementLineRequest struct {
	CashBankTransactionID *string `json:"cash_bank_transaction_id" binding:"omitempty,uuid"`
}

type ExcludeBankStatementLineRequest struct {
	Reason string `json:"reason"`
}

type BankStatementLineResponse struct {
	ID                       string                                   `json:"id"`
	Date                     string                                   `json:"date"`
	Reference                string                                   `json:"reference"`
	Description              string                                   `json:"description"`
	Amount                   float64                                  `json:"amount"`
	Direction                financeModels.BankStatementLineDirection `json:"direction"`
	Status                   financeModels.BankStatementLineStatus    `json:"status"`
	MatchedWithTransactionID *string                                  `json:"matched_with_transaction_id,omitempty"`
	ExcludeReason            string                                   `json:"exclude_reason,omitempty"`
}

type BankReconciliationResponse struct {
	ID               string                                 `json:"id"`
	CompanyID        string                                 `json:"company_id"`
	BankAccountID    string                                 `json:"bank_account_id"`
	StatementDate    string                                 `json:"statement_date"`
	StatementBalance float64                                `json:"statement_balance"`
	BookBalance      float64                                `json:"book_balance"`
	Difference       float64                                `json:"difference"`
	FileFormat       string                                 `json:"file_format"`
	FileName         string                                 `json:"file_name"`
	Status           financeModels.BankReconciliationStatus `json:"status"`
	ReconciledBy     *string                                `json:"reconciled_by,omitempty"`
	LockedBy         *string                                `json:"locked_by,omitempty"`
	ReconciledAt     *string                                `json:"reconciled_at,omitempty"`
	LockedAt         *string                                `json:"locked_at,omitempty"`
	Lines            []BankStatementLineResponse            `json:"bank_statement_lines,omitempty"`
	CreatedAt        string                                 `json:"created_at"`
	UpdatedAt        string                                 `json:"updated_at"`
}

type AutoMatchBankReconciliationResponse struct {
	ReconciliationID         string                                 `json:"reconciliation_id"`
	Status                   financeModels.BankReconciliationStatus `json:"status"`
	AutoMatchedCount         int                                    `json:"auto_matched_count"`
	ManualMatchRequiredCount int                                    `json:"manual_match_required_count"`
	UnmatchedCount           int                                    `json:"unmatched_count"`
	BankStatementLines       []BankStatementLineResponse            `json:"bank_statement_lines"`
}

type BankReconciliationFormDataResponse struct {
	BankAccounts []BankAccountReconciliationOption `json:"bank_accounts"`
	FileFormats  []ValueLabelOption                `json:"file_formats"`
}

type BankAccountReconciliationOption struct {
	ID                     string  `json:"id"`
	Code                   string  `json:"code"`
	Name                   string  `json:"name"`
	LastReconciliationDate *string `json:"last_reconciliation_date,omitempty"`
}
