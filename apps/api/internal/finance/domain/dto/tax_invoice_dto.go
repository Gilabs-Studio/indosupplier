package dto

import "time"

type CreateTaxInvoiceRequest struct {
	TaxInvoiceNumber string `json:"tax_invoice_number" binding:"required"`
	TaxInvoiceDate   string `json:"tax_invoice_date" binding:"required"`

	CustomerInvoiceID *string `json:"customer_invoice_id" binding:"omitempty,uuid"`
	SupplierInvoiceID *string `json:"supplier_invoice_id" binding:"omitempty,uuid"`

	DPPAmount   float64 `json:"dpp_amount" binding:"omitempty,gte=0"`
	VATAmount   float64 `json:"vat_amount" binding:"omitempty,gte=0"`
	TotalAmount float64 `json:"total_amount" binding:"omitempty,gte=0"`

	Notes string `json:"notes"`
}

type UpdateTaxInvoiceRequest struct {
	TaxInvoiceNumber string `json:"tax_invoice_number" binding:"required"`
	TaxInvoiceDate   string `json:"tax_invoice_date" binding:"required"`

	DPPAmount   float64 `json:"dpp_amount" binding:"omitempty,gte=0"`
	VATAmount   float64 `json:"vat_amount" binding:"omitempty,gte=0"`
	TotalAmount float64 `json:"total_amount" binding:"omitempty,gte=0"`

	Notes string `json:"notes"`
}

type ListTaxInvoicesRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search"`
	StartDate *string `form:"start_date"`
	EndDate *string `form:"end_date"`
	SortBy  string `form:"sort_by"`
	SortDir string `form:"sort_dir"`
}

type TaxInvoiceResponse struct {
	ID string `json:"id"`
	TaxInvoiceNumber string `json:"tax_invoice_number"`
	TaxInvoiceDate time.Time `json:"tax_invoice_date"`
	CustomerInvoiceID *string `json:"customer_invoice_id"`
	SupplierInvoiceID *string `json:"supplier_invoice_id"`
	DPPAmount float64 `json:"dpp_amount"`
	VATAmount float64 `json:"vat_amount"`
	TotalAmount float64 `json:"total_amount"`
	Notes string `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
