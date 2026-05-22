package dto

import financeModels "github.com/gilabs/gims/api/internal/finance/data/models"

type CreateCashBankTransactionRequest struct {
	CompanyID       string                                `json:"company_id" binding:"required,uuid"`
	BankAccountID   string                                `json:"bank_account_id" binding:"required,uuid"`
	Type            financeModels.CashBankTransactionType `json:"type" binding:"required,oneof=payment_in payment_out"`
	Date            string                                `json:"date" binding:"required"`
	Amount          float64                               `json:"amount" binding:"required,gt=0"`
	Reference       string                                `json:"reference"`
	Description     string                                `json:"description"`
	ContraAccountID *string                               `json:"contra_account_id" binding:"omitempty,uuid"`
	AttachmentURL   *string                               `json:"attachment_url"`
	AutoPost        *bool                                 `json:"auto_post"`
}

type ListCashBankTransactionsRequest struct {
	Page          int                                      `form:"page" binding:"omitempty,min=1"`
	PerPage       int                                      `form:"per_page" binding:"omitempty,min=1,max=100"`
	CompanyID     string                                   `form:"company_id"`
	BankAccountID string                                   `form:"bank_account_id"`
	Type          *financeModels.CashBankTransactionType   `form:"type"`
	Status        *financeModels.CashBankTransactionStatus `form:"status"`
	DateFrom      *string                                  `form:"date_from"`
	DateTo        *string                                  `form:"date_to"`
	Search        string                                   `form:"search"`
	SortBy        string                                   `form:"sort_by"`
	SortDir       string                                   `form:"sort_dir"`
}

type ReverseCashBankTransactionRequest struct {
	Reason string `json:"reason"`
}

type CashBankTransactionResponse struct {
	ID                 string                                  `json:"id"`
	CompanyID          string                                  `json:"company_id"`
	ReferenceNumber    string                                  `json:"reference_number"`
	BankAccountID      string                                  `json:"bank_account_id"`
	Type               financeModels.CashBankTransactionType   `json:"type"`
	Date               string                                  `json:"date"`
	Amount             float64                                 `json:"amount"`
	Reference          string                                  `json:"reference"`
	Description        string                                  `json:"description"`
	ContraAccountID    *string                                 `json:"contra_account_id,omitempty"`
	JournalEntryID     *string                                 `json:"journal_entry_id,omitempty"`
	JournalEntryNumber string                                  `json:"journal_entry_number,omitempty"`
	Status             financeModels.CashBankTransactionStatus `json:"status"`
	AttachmentURL      *string                                 `json:"attachment_url,omitempty"`
	ReverseOfID        *string                                 `json:"reverse_of_id,omitempty"`
	ReversedByID       *string                                 `json:"reversed_by_id,omitempty"`
	CreatedBy          *string                                 `json:"created_by,omitempty"`
	UpdatedBy          *string                                 `json:"updated_by,omitempty"`
	CreatedAt          string                                  `json:"created_at"`
	UpdatedAt          string                                  `json:"updated_at"`
}

type CashBankTransactionFormDataResponse struct {
	BankAccounts     []CashBankAccountFormOption `json:"bank_accounts"`
	TransactionTypes []ValueLabelOption          `json:"transaction_types"`
	ContraAccounts   []CashBankCOAFormOption     `json:"contra_accounts"`
}

type CashBankAccountFormOption struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
}

type CashBankCOAFormOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ValueLabelOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
