package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type JournalLineRequest struct {
	ChartOfAccountID string  `json:"chart_of_account_id" binding:"required,uuid"`
	Debit            float64 `json:"debit" binding:"omitempty,gte=0"`
	Credit           float64 `json:"credit" binding:"omitempty,gte=0"`
	Memo             string  `json:"memo"`
}

type CreateJournalEntryRequest struct {
	CompanyID                    string               `json:"company_id" binding:"required,uuid"`
	FiscalYearID                 *string              `json:"fiscal_year_id"`
	EntryDate                    string               `json:"entry_date" binding:"required"`
	Reference                    string               `json:"reference"`
	Description                  string               `json:"description"`
	ReferenceType                *string              `json:"reference_type"`
	ReferenceID                  *string              `json:"reference_id"`
	JournalType                  *string              `json:"journal_type"`
	CurrencyCode                 string               `json:"currency_code"`
	ExchangeRate                 *float64             `json:"exchange_rate"`
	Lines                        []JournalLineRequest `json:"lines" binding:"required,min=2"`
	IsSystemGenerated            bool                 `json:"is_system_generated"`
	SkipControlAccountValidation bool                 `json:"-"`
	SourceDocumentURL            *string              `json:"source_document_url"`
}

type UpdateJournalEntryRequest struct {
	CompanyID     string               `json:"company_id" binding:"required,uuid"`
	FiscalYearID  *string              `json:"fiscal_year_id"`
	EntryDate     string               `json:"entry_date" binding:"required"`
	Reference     string               `json:"reference"`
	Description   string               `json:"description"`
	ReferenceType *string              `json:"reference_type"`
	ReferenceID   *string              `json:"reference_id"`
	JournalType   *string              `json:"journal_type"`
	CurrencyCode  string               `json:"currency_code"`
	ExchangeRate  *float64             `json:"exchange_rate"`
	Lines         []JournalLineRequest `json:"lines" binding:"required,min=2"`
}

// CreateAdjustmentJournalRequest is used for the dedicated adjustment journal endpoint.
// reference_type is always forced to "MANUAL_ADJUSTMENT" on the backend.
// description is required for audit trail.
type CreateAdjustmentJournalRequest struct {
	CompanyID                    string               `json:"company_id" binding:"required,uuid"`
	FiscalYearID                 *string              `json:"fiscal_year_id"`
	EntryDate                    string               `json:"entry_date" binding:"required"`
	Description                  string               `json:"description" binding:"required,min=3"`
	Reference                    string               `json:"reference"`
	CurrencyCode                 string               `json:"currency_code"`
	ExchangeRate                 *float64             `json:"exchange_rate"`
	SourceDocumentURL            *string              `json:"source_document_url"`
	Lines                        []JournalLineRequest `json:"lines" binding:"required,min=2"`
	SkipControlAccountValidation bool                 `json:"-"`
}

// CreateOpeningBalanceJournalRequest is an internal command payload for opening balance journals.
// Opening balances are journal-driven and are not persisted on Chart of Account rows.
type CreateOpeningBalanceJournalRequest struct {
	AccountID   string  `json:"account_id" binding:"required,uuid"`
	Amount      float64 `json:"amount" binding:"required"`
	EntryDate   *string `json:"entry_date"`
	Description *string `json:"description"`
}

type ListJournalEntriesRequest struct {
	Page          int                          `form:"page" binding:"omitempty,min=1"`
	PerPage       int                          `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search        string                       `form:"search"`
	Domain        *string                      `form:"domain" binding:"omitempty,oneof=sales purchase inventory stock cash_bank finance adjustment valuation"`
	Status        *financeModels.JournalStatus `form:"status" binding:"omitempty,oneof=draft posted reversed cancelled"`
	CompanyID     string                       `form:"company_id"`
	FiscalYearID  *string                      `form:"fiscal_year_id"`
	JournalType   *string                      `form:"journal_type"`
	StartDate     *string                      `form:"start_date"`
	EndDate       *string                      `form:"end_date"`
	SortBy        string                       `form:"sort_by"`
	SortDir       string                       `form:"sort_dir"`
	ReferenceType *string                      `form:"reference_type"`
}

type JournalLineResponse struct {
	ID               string                  `json:"id"`
	ChartOfAccountID string                  `json:"chart_of_account_id"`
	ChartOfAccount   *ChartOfAccountResponse `json:"chart_of_account,omitempty"`
	Debit            float64                 `json:"debit"`
	Credit           float64                 `json:"credit"`
	Memo             string                  `json:"memo"`
}

type JournalEntryResponse struct {
	ID            string                      `json:"id"`
	CompanyID     string                      `json:"company_id,omitempty"`
	FiscalYearID  *string                     `json:"fiscal_year_id,omitempty"`
	JournalNumber string                      `json:"journal_number,omitempty"`
	EntryDate     time.Time                   `json:"entry_date"`
	Reference     string                      `json:"reference,omitempty"`
	Description   string                      `json:"description"`
	ReferenceType *string                     `json:"reference_type"`
	ReferenceID   *string                     `json:"reference_id"`
	ReferenceCode *string                     `json:"reference_code"`
	Status        financeModels.JournalStatus `json:"status"`
	JournalType   string                      `json:"journal_type"`
	PostedAt      *time.Time                  `json:"posted_at"`
	PostedBy      *string                     `json:"posted_by"`
	PostedByName  *string                     `json:"posted_by_name,omitempty"`
	PostedByEmail *string                     `json:"posted_by_email,omitempty"`

	CreatedBy      *string `json:"created_by,omitempty"`
	CreatedByName  *string `json:"created_by_name,omitempty"`
	CreatedByEmail *string `json:"created_by_email,omitempty"`

	ReversedAt      *time.Time `json:"reversed_at,omitempty"`
	ReversedBy      *string    `json:"reversed_by,omitempty"`
	ReversedByName  *string    `json:"reversed_by_name,omitempty"`
	ReversedByEmail *string    `json:"reversed_by_email,omitempty"`
	ReversalReason  string     `json:"reversal_reason,omitempty"`

	IsSystemGenerated bool                  `json:"is_system_generated"`
	SourceDocumentURL *string               `json:"source_document_url,omitempty"`
	Lines             []JournalLineResponse `json:"lines"`
	DebitTotal        float64               `json:"debit_total"`
	CreditTotal       float64               `json:"credit_total"`
	CurrencyCode      string                `json:"currency_code"`
	ExchangeRate      float64               `json:"exchange_rate"`
	IsReversal        bool                  `json:"is_reversal"`
	ReversedFrom      *string               `json:"reversed_from,omitempty"`
	IsValuation       bool                  `json:"is_valuation"`
	Source            string                `json:"source"`
	ValuationRunID    *string               `json:"valuation_run_id,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

type TrialBalanceRow struct {
	ChartOfAccountID string                    `json:"chart_of_account_id"`
	Code             string                    `json:"code"`
	Name             string                    `json:"name"`
	Type             financeModels.AccountType `json:"type"`
	OpeningBalance   float64                   `json:"opening_balance"`
	DebitTotal       float64                   `json:"debit_total"`
	CreditTotal      float64                   `json:"credit_total"`
	Balance          float64                   `json:"balance"`
}

type TrialBalanceResponse struct {
	StartDate *time.Time        `json:"start_date"`
	EndDate   *time.Time        `json:"end_date"`
	Rows      []TrialBalanceRow `json:"rows"`
}

// ===== Journal Lines DTOs (sub-ledger list view) =====

// ListJournalLinesRequest for filtering journal lines with pagination
type ListJournalLinesRequest struct {
	Page              int     `form:"page" binding:"omitempty,min=1"`
	PerPage           int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search            string  `form:"search"`
	CashBankJournalID string  `form:"cash_bank_journal_id" binding:"omitempty,uuid"`
	ChartOfAccountID  string  `form:"chart_of_account_id" binding:"omitempty,uuid"`
	AccountType       string  `form:"account_type"`
	ReferenceType     *string `form:"reference_type"`
	JournalStatus     string  `form:"journal_status" binding:"omitempty,oneof=draft posted"`
	StartDate         *string `form:"start_date"`
	EndDate           *string `form:"end_date"`
	SortBy            string  `form:"sort_by"`
	SortDir           string  `form:"sort_dir"`
}

// JournalLineDetailResponse for individual journal line with entry context
type JournalLineDetailResponse struct {
	ID                 string  `json:"id"`
	JournalEntryID     string  `json:"journal_entry_id"`
	EntryDate          string  `json:"entry_date"`
	JournalDescription string  `json:"journal_description"`
	JournalStatus      string  `json:"journal_status"`
	ReferenceType      *string `json:"reference_type"`
	ReferenceID        *string `json:"reference_id"`
	ChartOfAccountID   string  `json:"chart_of_account_id"`
	ChartOfAccountCode string  `json:"chart_of_account_code"`
	ChartOfAccountName string  `json:"chart_of_account_name"`
	ChartOfAccountType string  `json:"chart_of_account_type"`
	Debit              float64 `json:"debit"`
	Credit             float64 `json:"credit"`
	Memo               string  `json:"memo"`
	RunningBalance     float64 `json:"running_balance"`
	CreatedAt          string  `json:"created_at"`
}

// ListJournalLinesResponse wraps lines with summary totals
type ListJournalLinesResponse struct {
	Lines       []JournalLineDetailResponse `json:"lines"`
	TotalDebit  float64                     `json:"total_debit"`
	TotalCredit float64                     `json:"total_credit"`
}

// ===== Form Data DTOs =====

// COAFormOption represents a Chart of Account option for dropdowns.
type COAFormOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// BankAccountFormOption represents a Bank Account option for dropdowns.
type BankAccountFormOption struct {
	ID            string  `json:"id"`
	AccountName   string  `json:"account_name"`
	AccountNumber string  `json:"account_number"`
	BankName      string  `json:"bank_name"`
	Currency      string  `json:"currency"`
	COAId         *string `json:"coa_id,omitempty"`
}

// JournalEntryFormDataResponse for journal entry form options.
type JournalEntryFormDataResponse struct {
	ChartOfAccounts []COAFormOption `json:"chart_of_accounts"`
	JournalTypes    []EnumOption    `json:"journal_types"`
	ReferenceTypes  []EnumOption    `json:"reference_types"`
	Currencies      []EnumOption    `json:"currencies"`
	Statuses        []EnumOption    `json:"statuses"`
}

// ===== Valuation DTOs =====

// RunValuationRequest is the request payload for triggering a valuation run.
type RunValuationRequest struct {
	ValuationType string `json:"valuation_type" binding:"required,oneof=inventory fx depreciation"`
	PeriodStart   string `json:"period_start" binding:"required"` // YYYY-MM-DD
	PeriodEnd     string `json:"period_end" binding:"required"`   // YYYY-MM-DD
	ReferenceID   string `json:"reference_id"`                    // optional, for idempotency
}

// ValuationItemResponse is the per-item valuation breakdown.
type ValuationItemResponse struct {
	ReferenceID string  `json:"reference_id"`
	ProductID   *string `json:"product_id,omitempty"`
	Qty         float64 `json:"qty"`
	BookValue   float64 `json:"book_value"`
	ActualValue float64 `json:"actual_value"`
	Delta       float64 `json:"delta"`
	Direction   string  `json:"direction"`
}

// ValuationPreviewJournalLine is a journal line preview row.
type ValuationPreviewJournalLine struct {
	ChartOfAccountID string  `json:"chart_of_account_id"`
	Debit            float64 `json:"debit"`
	Credit           float64 `json:"credit"`
	Memo             string  `json:"memo"`
}

// ValuationPreviewResponse returns valuation and journal preview before posting.
type ValuationPreviewResponse struct {
	ValuationType string                        `json:"valuation_type"`
	PeriodStart   string                        `json:"period_start"`
	PeriodEnd     string                        `json:"period_end"`
	Items         []ValuationItemResponse       `json:"items"`
	TotalDelta    float64                       `json:"total_delta"`
	TotalGain     float64                       `json:"total_gain"`
	TotalLoss     float64                       `json:"total_loss"`
	JournalLines  []ValuationPreviewJournalLine `json:"journal_lines"`
	IsBalanced    bool                          `json:"is_balanced"`
}

// ApproveValuationRequest is the payload to approve and post a valuation run.
type ApproveValuationRequest struct {
	Notes string `json:"notes"`
}

// UnlockValuationRequest is the payload to unlock a posted (locked) valuation run.
// Only allowed for admin/finance_manager users to enable corrections if errors are discovered.
type UnlockValuationRequest struct {
	UnlockReason string `json:"unlock_reason" binding:"required,min=3"`
}

// BulkApproveValuationRequest is the payload to bulk approve multiple valuation runs.
type BulkApproveValuationRequest struct {
	RunIDs []string `json:"run_ids" binding:"required,min=1,max=100"`
}

// BulkApproveResult is the status of a single run in bulk approve operation.
type BulkApproveResult struct {
	RunID   string `json:"run_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BulkApproveValuationResponse is the response for batch approval operation.
type BulkApproveValuationResponse struct {
	Results        []BulkApproveResult `json:"results"`
	TotalProcessed int                 `json:"total_processed"`
	SuccessCount   int                 `json:"success_count"`
	FailureCount   int                 `json:"failure_count"`
	ProcessedAt    string              `json:"processed_at"`
}

// ValuationRunResponse is the API response for a valuation run record.
type ValuationRunResponse struct {
	ID             string  `json:"id"`
	ReferenceID    string  `json:"reference_id"`
	ValuationType  string  `json:"valuation_type"`
	PeriodStart    string  `json:"period_start"`
	PeriodEnd      string  `json:"period_end"`
	Status         string  `json:"status"`
	TotalDebit     float64 `json:"total_debit"`
	TotalCredit    float64 `json:"total_credit"`
	TotalDelta     float64 `json:"total_delta"`
	JournalEntryID *string `json:"journal_entry_id,omitempty"`
	ErrorMessage   *string `json:"error_message,omitempty"`

	// Approval Tracking (audit trail)
	IsLocked      bool    `json:"is_locked"`
	LockedAt      *string `json:"locked_at,omitempty"`
	ApprovedBy    *string `json:"approved_by,omitempty"`
	ApprovedAt    *string `json:"approved_at,omitempty"`
	ApprovalNotes string  `json:"approval_notes,omitempty"`

	CreatedBy   *string                 `json:"created_by,omitempty"`
	CompletedAt *string                 `json:"completed_at,omitempty"`
	Items       []ValuationItemResponse `json:"items,omitempty"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

// ValuationKPIMeta is additional metadata returned with valuation list endpoints.
type ValuationKPIMeta struct {
	TotalEntries   int64   `json:"total_entries"`
	TotalDebitSum  float64 `json:"total_debit_sum"`
	TotalCreditSum float64 `json:"total_credit_sum"`
	CompletedRuns  int64   `json:"completed_runs"`
	ProcessingRuns int64   `json:"processing_runs"`
	FailedRuns     int64   `json:"failed_runs"`
}

// ListValuationRunsRequest for filtering valuation runs.
type ListValuationRunsRequest struct {
	Page          int     `form:"page" binding:"omitempty,min=1"`
	PerPage       int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	ValuationType *string `form:"valuation_type" binding:"omitempty,oneof=inventory fx depreciation"`
	Status        *string `form:"status" binding:"omitempty,oneof=draft pending_approval approved posted no_difference failed"`
	StartDate     *string `form:"start_date"`
	EndDate       *string `form:"end_date"`
	SortBy        string  `form:"sort_by"`
	SortDir       string  `form:"sort_dir"`
}
