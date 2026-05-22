package dto

import "time"

type SupplierInvoiceDownPaymentRegularInvoiceMini struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type SupplierInvoiceDownPaymentListResponse struct {
	ID string `json:"id"`

	CompanyID    *string `json:"company_id"`
	FiscalYearID *string `json:"fiscal_year_id"`

	PurchaseOrder *SupplierInvoicePurchaseOrderMini `json:"purchase_order,omitempty"`

	SupplierID   string `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`

	Code            string  `json:"code"`
	InvoiceNumber   string  `json:"invoice_number"`
	InvoiceDate     string  `json:"invoice_date"`
	DueDate         string  `json:"due_date"`
	Amount          float64 `json:"amount"`
	PaidAmount      float64 `json:"paid_amount"`
	RemainingAmount float64 `json:"remaining_amount"`
	Status          string  `json:"status"`
	Notes           *string `json:"notes"`

	RegularInvoices []SupplierInvoiceDownPaymentRegularInvoiceMini `json:"regular_invoices,omitempty"`

	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	RejectedAt  *time.Time `json:"rejected_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SupplierInvoiceDownPaymentDetailResponse struct {
	ID string `json:"id"`

	CompanyID    *string `json:"company_id"`
	FiscalYearID *string `json:"fiscal_year_id"`

	PurchaseOrder *SupplierInvoicePurchaseOrderMini `json:"purchase_order,omitempty"`

	SupplierID   string `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`

	Code            string  `json:"code"`
	InvoiceNumber   string  `json:"invoice_number"`
	InvoiceDate     string  `json:"invoice_date"`
	DueDate         string  `json:"due_date"`
	Amount          float64 `json:"amount"`
	PaidAmount      float64 `json:"paid_amount"`
	RemainingAmount float64 `json:"remaining_amount"`
	Status          string  `json:"status"`
	Notes           *string `json:"notes"`

	RegularInvoices []SupplierInvoiceDownPaymentRegularInvoiceMini `json:"regular_invoices,omitempty"`

	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	RejectedAt  *time.Time `json:"rejected_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	CreatedBy string `json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSupplierInvoiceDownPaymentRequest struct {
	PurchaseOrderID string  `json:"purchase_order_id" binding:"required,uuid"`
	InvoiceDate     string  `json:"invoice_date" binding:"required"`
	DueDate         string  `json:"due_date" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Notes           *string `json:"notes"`
}

type UpdateSupplierInvoiceDownPaymentRequest struct {
	PurchaseOrderID string  `json:"purchase_order_id" binding:"required,uuid"`
	InvoiceDate     string  `json:"invoice_date" binding:"required"`
	DueDate         string  `json:"due_date" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Notes           *string `json:"notes"`
}

type SupplierInvoiceDownPaymentAddResponse struct {
	PurchaseOrders []SupplierInvoiceAddPurchaseOrder `json:"purchase_orders"`
}
