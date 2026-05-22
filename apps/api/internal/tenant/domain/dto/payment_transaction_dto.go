package dto

import (
	"time"
)

// PaymentTransactionResponse represents a payment transaction returned to clients.
type PaymentTransactionResponse struct {
	ID               string    `json:"id"`
	Provider         string    `json:"provider"`
	Status           string    `json:"status"`
	PaymentMethod    string    `json:"payment_method,omitempty"`
	AmountIDR        int64     `json:"amount_idr"`
	ProviderInvoiceID string   `json:"provider_invoice_id,omitempty"`
	ReceiptURL       string    `json:"receipt_url,omitempty"`
	InvoiceURL       string    `json:"invoice_url,omitempty"`
	Description      string    `json:"description,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// PaymentHistoryListParams query parameters for listing payment transactions.
type PaymentHistoryListParams struct {
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
	Status   string `form:"status"`
	Provider string `form:"provider"`
}

// PaymentHistoryListResponse represents a paginated list of payment transactions.
type PaymentHistoryListResponse struct {
	Data       []*PaymentTransactionResponse `json:"data"`
	Pagination interface{}                   `json:"pagination,omitempty"`
}

// CreatePaymentTransactionRequest is used to create a new payment transaction (internal).
type CreatePaymentTransactionRequest struct {
	TenantID           string `json:"tenant_id" binding:"required"`
	SubscriptionID     string `json:"subscription_id,omitempty"`
	Provider           string `json:"provider" binding:"required,oneof=xendit midtrans internal"`
	Status             string `json:"status" binding:"required,oneof=pending paid failed expired canceled"`
	PaymentMethod      string `json:"payment_method,omitempty"`
	AmountIDR          int64  `json:"amount_idr" binding:"required,min=1"`
	ProviderInvoiceID  string `json:"provider_invoice_id,omitempty"`
	ProviderPaymentID  string `json:"provider_payment_id,omitempty"`
	ReceiptURL         string `json:"receipt_url,omitempty"`
	InvoiceURL         string `json:"invoice_url,omitempty"`
	Description        string `json:"description,omitempty"`
	Metadata           string `json:"metadata,omitempty"`
	Notes              string `json:"notes,omitempty"`
}

// UpdatePaymentTransactionRequest is used to update payment transaction status.
type UpdatePaymentTransactionRequest struct {
	Status            string `json:"status" binding:"omitempty,oneof=pending paid failed expired canceled"`
	PaymentMethod     string `json:"payment_method,omitempty"`
	ProviderPaymentID string `json:"provider_payment_id,omitempty"`
	ReceiptURL        string `json:"receipt_url,omitempty"`
	Notes             string `json:"notes,omitempty"`
}
