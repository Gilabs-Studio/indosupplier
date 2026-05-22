package dto

import "time"

type CreateNonTradePayableRequest struct {
	TransactionDate  string  `json:"transaction_date" binding:"required"`
	Description      string  `json:"description"`
	ChartOfAccountID string  `json:"chart_of_account_id" binding:"required,uuid"`
	Amount           float64 `json:"amount" binding:"required,gt=0"`
	VendorName       string  `json:"vendor_name"`
	DueDate          *string `json:"due_date"`
	ReferenceNo      string  `json:"reference_no"`
}

type UpdateNonTradePayableRequest struct {
	TransactionDate  string  `json:"transaction_date" binding:"required"`
	Description      string  `json:"description"`
	ChartOfAccountID string  `json:"chart_of_account_id" binding:"required,uuid"`
	Amount           float64 `json:"amount" binding:"required,gt=0"`
	VendorName       string  `json:"vendor_name"`
	DueDate          *string `json:"due_date"`
	ReferenceNo      string  `json:"reference_no"`
}

type ListNonTradePayablesRequest struct {
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PerPage   int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search    string  `form:"search"`
	StartDate *string `form:"start_date"`
	EndDate   *string `form:"end_date"`
	SortBy    string  `form:"sort_by"`
	SortDir   string  `form:"sort_dir"`
	Status    *string `form:"status"`
}

type NonTradePayableResponse struct {
	ID               string                  `json:"id"`
	TransactionDate  time.Time               `json:"transaction_date"`
	Code             string                  `json:"code"`
	Description      string                  `json:"description"`
	ChartOfAccountID string                  `json:"chart_of_account_id"`
	ChartOfAccount   *ChartOfAccountResponse `json:"chart_of_account,omitempty"`
	Amount           float64                 `json:"amount"`
	VendorName       string                  `json:"vendor_name"`
	DueDate          *time.Time              `json:"due_date"`
	ReferenceNo      string                  `json:"reference_no"`
	Status           string                  `json:"status"`
	PaidAmount       float64                 `json:"paid_amount,omitempty"`
	RemainingAmount  float64                 `json:"remaining_amount,omitempty"`
	JournalID        *string                 `json:"journal_id,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

type PayNonTradePayableRequest struct {
	PaymentDate      string  `json:"payment_date" binding:"required"`
	BankReference    string  `json:"bank_reference"`
	ChartOfAccountID string  `json:"chart_of_account_id" binding:"omitempty,uuid"` // Bank/Cash Account
	BankAccountID    string  `json:"bank_account_id" binding:"omitempty,uuid"`
	Amount           float64 `json:"amount" binding:"required,gt=0"`
}
