package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type PaymentAllocationRequest struct {
	ChartOfAccountID string  `json:"chart_of_account_id" binding:"required,uuid"`
	ReferenceType    *string `json:"reference_type" binding:"required"`
	ReferenceID      *string `json:"reference_id" binding:"required,uuid"`
	Amount           float64 `json:"amount" binding:"required,gt=0"`
	Memo             string  `json:"memo"`
}

type CreatePaymentRequest struct {
	PaymentDate   string                    `json:"payment_date" binding:"required"`
	Description   string                    `json:"description"`
	BankAccountID string                    `json:"bank_account_id" binding:"required,uuid"`
	TotalAmount   float64                   `json:"total_amount" binding:"required,gt=0"`
	Allocations   []PaymentAllocationRequest `json:"allocations" binding:"required,min=1,dive"`
}

type UpdatePaymentRequest struct {
	PaymentDate   string                    `json:"payment_date" binding:"required"`
	Description   string                    `json:"description"`
	BankAccountID string                    `json:"bank_account_id" binding:"required,uuid"`
	TotalAmount   float64                   `json:"total_amount" binding:"required,gt=0"`
	Allocations   []PaymentAllocationRequest `json:"allocations" binding:"required,min=1,dive"`
}

type ListPaymentsRequest struct {
	Page     int                         `form:"page" binding:"omitempty,min=1"`
	PerPage  int                         `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search   string                      `form:"search"`
	Status   *financeModels.PaymentStatus `form:"status" binding:"omitempty,oneof=draft posted"`
	StartDate *string                    `form:"start_date"`
	EndDate   *string                    `form:"end_date"`
	SortBy   string                      `form:"sort_by"`
	SortDir  string                      `form:"sort_dir"`
}

type PaymentAllocationResponse struct {
	ID             string     `json:"id"`
	ChartOfAccountID string   `json:"chart_of_account_id"`
	ChartOfAccount *ChartOfAccountResponse `json:"chart_of_account,omitempty"`
	ReferenceType  *string    `json:"reference_type"`
	ReferenceID    *string    `json:"reference_id"`
	Amount         float64    `json:"amount"`
	Memo           string     `json:"memo"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type PaymentResponse struct {
	ID            string                   `json:"id"`
	PaymentDate   time.Time                `json:"payment_date"`
	Description   string                   `json:"description"`
	BankAccountID string                   `json:"bank_account_id"`
	BankAccount   *BankAccountMini         `json:"bank_account,omitempty"`
	TotalAmount   float64                  `json:"total_amount"`
	Status        financeModels.PaymentStatus `json:"status"`
	JournalEntryID *string                 `json:"journal_entry_id"`
	ApprovedAt    *time.Time               `json:"approved_at"`
	ApprovedBy    *string                  `json:"approved_by"`
	PostedAt      *time.Time               `json:"posted_at"`
	PostedBy      *string                  `json:"posted_by"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	Allocations   []PaymentAllocationResponse `json:"allocations,omitempty"`
}
