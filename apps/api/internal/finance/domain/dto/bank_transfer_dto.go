package dto

import financeModels "github.com/gilabs/gims/api/internal/finance/data/models"

type CreateBankTransferRequest struct {
	CompanyID         string  `json:"company_id" binding:"required,uuid"`
	FromBankAccountID string  `json:"from_bank_account_id" binding:"required,uuid"`
	ToBankAccountID   string  `json:"to_bank_account_id" binding:"required,uuid"`
	Amount            float64 `json:"amount" binding:"required,gt=0"`
	Date              string  `json:"date" binding:"required"`
	Reference         string  `json:"reference"`
	Description       string  `json:"description"`
}

type ListBankTransfersRequest struct {
	Page              int                               `form:"page" binding:"omitempty,min=1"`
	PerPage           int                               `form:"per_page" binding:"omitempty,min=1,max=100"`
	CompanyID         string                            `form:"company_id"`
	FromBankAccountID string                            `form:"from_bank_account_id"`
	ToBankAccountID   string                            `form:"to_bank_account_id"`
	Status            *financeModels.BankTransferStatus `form:"status"`
	DateFrom          *string                           `form:"date_from"`
	DateTo            *string                           `form:"date_to"`
	Search            string                            `form:"search"`
	SortBy            string                            `form:"sort_by"`
	SortDir           string                            `form:"sort_dir"`
}

type CancelBankTransferRequest struct {
	Reason string `json:"reason"`
}

type BankTransferResponse struct {
	ID                string                           `json:"id"`
	CompanyID         string                           `json:"company_id"`
	TransferNumber    string                           `json:"transfer_number"`
	FromBankAccountID string                           `json:"from_bank_account_id"`
	ToBankAccountID   string                           `json:"to_bank_account_id"`
	Amount            float64                          `json:"amount"`
	Date              string                           `json:"date"`
	Reference         string                           `json:"reference"`
	Description       string                           `json:"description"`
	Status            financeModels.BankTransferStatus `json:"status"`
	JournalEntryID    *string                          `json:"journal_entry_id,omitempty"`
	CreatedBy         *string                          `json:"created_by,omitempty"`
	UpdatedBy         *string                          `json:"updated_by,omitempty"`
	CreatedAt         string                           `json:"created_at"`
	UpdatedAt         string                           `json:"updated_at"`
}
