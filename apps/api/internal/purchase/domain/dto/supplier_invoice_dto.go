package dto

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	purchaseModels "github.com/gilabs/gims/api/internal/purchase/data/models"
)

type SupplierInvoiceAuditTrailEntry = PurchaseRequisitionAuditTrailEntry

type SupplierInvoicePurchaseOrderMini struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type SupplierInvoiceGoodsReceiptMini struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type SupplierInvoicePaymentTermsMini struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Days *int   `json:"days,omitempty"`
}

type SupplierInvoiceListResponse struct {
	ID string `json:"id"`

	CompanyID    *string `json:"company_id"`
	FiscalYearID *string `json:"fiscal_year_id"`

	PurchaseOrder *SupplierInvoicePurchaseOrderMini `json:"purchase_order,omitempty"`
	GoodsReceipt  *SupplierInvoiceGoodsReceiptMini  `json:"goods_receipt,omitempty"`
	PaymentTerms  *SupplierInvoicePaymentTermsMini  `json:"payment_terms,omitempty"`

	Type          string `json:"type"`
	Code          string `json:"code"`
	InvoiceNumber string `json:"invoice_number"`
	InvoiceDate   string `json:"invoice_date"`
	DueDate       string `json:"due_date"`

	SupplierID   string `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`

	TaxRate           float64 `json:"tax_rate"`
	TaxAmount         float64 `json:"tax_amount"`
	DeliveryCost      float64 `json:"delivery_cost"`
	OtherCost         float64 `json:"other_cost"`
	SubTotal          float64 `json:"sub_total"`
	Amount            float64 `json:"amount"`
	PaidAmount        float64 `json:"paid_amount"`
	RemainingAmount   float64 `json:"remaining_amount"`
	DownPaymentAmount float64 `json:"down_payment_amount"`

	DownPaymentInvoice *SupplierInvoiceAddDownPaymentMini `json:"down_payment_invoice,omitempty"`
	IsPosted           bool                               `json:"is_posted"`
	JournalEntryID     *string                            `json:"journal_entry_id,omitempty"`

	Status string  `json:"status"`
	Notes  *string `json:"notes"`

	CreatedBy   string     `json:"created_by"`
	SubmittedAt *time.Time `json:"submitted_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
	RejectedAt  *time.Time `json:"rejected_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
}

type SupplierInvoiceItemResponse struct {
	ID                  string      `json:"id"`
	SupplierInvoiceID   string      `json:"supplier_invoice_id"`
	ProductID           string      `json:"product_id"`
	Product             interface{} `json:"product,omitempty"`
	Quantity            float64     `json:"quantity"`
	Price               float64     `json:"price"`
	Discount            float64     `json:"discount"`
	SubTotal            float64     `json:"sub_total"`
	PurchaseOrderItemID *string     `json:"purchase_order_item_id,omitempty"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}

type SupplierInvoiceDetailResponse struct {
	ID string `json:"id"`

	CompanyID    *string `json:"company_id"`
	FiscalYearID *string `json:"fiscal_year_id"`

	PurchaseOrder *SupplierInvoicePurchaseOrderMini `json:"purchase_order,omitempty"`
	GoodsReceipt  *SupplierInvoiceGoodsReceiptMini  `json:"goods_receipt,omitempty"`
	PaymentTerms  *SupplierInvoicePaymentTermsMini  `json:"payment_terms,omitempty"`

	Type          string `json:"type"`
	Code          string `json:"code"`
	InvoiceNumber string `json:"invoice_number"`
	InvoiceDate   string `json:"invoice_date"`
	DueDate       string `json:"due_date"`

	SupplierID   string `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`

	TaxRate           float64 `json:"tax_rate"`
	TaxAmount         float64 `json:"tax_amount"`
	DeliveryCost      float64 `json:"delivery_cost"`
	OtherCost         float64 `json:"other_cost"`
	SubTotal          float64 `json:"sub_total"`
	Amount            float64 `json:"amount"`
	PaidAmount        float64 `json:"paid_amount"`
	RemainingAmount   float64 `json:"remaining_amount"`
	DownPaymentAmount float64 `json:"down_payment_amount"`

	DownPaymentInvoice *SupplierInvoiceAddDownPaymentMini `json:"down_payment_invoice,omitempty"`
	IsPosted           bool                               `json:"is_posted"`
	JournalEntryID     *string                            `json:"journal_entry_id,omitempty"`

	Status string  `json:"status"`
	Notes  *string `json:"notes"`

	Items []SupplierInvoiceItemResponse `json:"items"`

	CreatedBy   string     `json:"created_by"`
	SubmittedAt *time.Time `json:"submitted_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
	RejectedAt  *time.Time `json:"rejected_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
}

type CreateSupplierInvoiceItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Discount  float64 `json:"discount" binding:"omitempty,gte=0,lte=100"`
}

type CreateSupplierInvoiceRequest struct {
	GoodsReceiptID string                             `json:"goods_receipt_id" binding:"required,uuid"`
	PaymentTermsID string                             `json:"payment_terms_id" binding:"required,uuid"`
	InvoiceNumber  string                             `json:"invoice_number" binding:"required"`
	InvoiceDate    string                             `json:"invoice_date" binding:"required"`
	DueDate        string                             `json:"due_date" binding:"required"`
	TaxRate        float64                            `json:"tax_rate" binding:"omitempty,gte=0,lte=100"`
	DeliveryCost   float64                            `json:"delivery_cost" binding:"omitempty,gte=0"`
	OtherCost      float64                            `json:"other_cost" binding:"omitempty,gte=0"`
	Notes          *string                            `json:"notes"`
	Items          []CreateSupplierInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateSupplierInvoiceRequest struct {
	GoodsReceiptID string                             `json:"goods_receipt_id" binding:"required,uuid"`
	PaymentTermsID string                             `json:"payment_terms_id" binding:"required,uuid"`
	InvoiceNumber  string                             `json:"invoice_number" binding:"required"`
	InvoiceDate    string                             `json:"invoice_date" binding:"required"`
	DueDate        string                             `json:"due_date" binding:"required"`
	TaxRate        float64                            `json:"tax_rate" binding:"omitempty,gte=0,lte=100"`
	DeliveryCost   float64                            `json:"delivery_cost" binding:"omitempty,gte=0"`
	OtherCost      float64                            `json:"other_cost" binding:"omitempty,gte=0"`
	Notes          *string                            `json:"notes"`
	Items          []CreateSupplierInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
}

// SupplierInvoiceJournalPreviewRequest reuses create payload for simulation.
type SupplierInvoiceJournalPreviewRequest = CreateSupplierInvoiceRequest

type SupplierInvoiceThreeWayMatchingLine struct {
	ProductID             string  `json:"product_id"`
	ProductCode           string  `json:"product_code,omitempty"`
	ProductName           string  `json:"product_name,omitempty"`
	QuantityPO            float64 `json:"quantity_po"`
	QuantityGR            float64 `json:"quantity_gr"`
	QuantityAlreadyBilled float64 `json:"quantity_already_billed"`
	QuantityBill          float64 `json:"quantity_bill"`
	QuantityRemaining     float64 `json:"quantity_remaining"`
	IsValid               bool    `json:"is_valid"`
}

type SupplierInvoiceJournalPreviewLine struct {
	ChartOfAccountID   string  `json:"chart_of_account_id"`
	ChartOfAccountCode string  `json:"chart_of_account_code,omitempty"`
	ChartOfAccountName string  `json:"chart_of_account_name,omitempty"`
	Debit              float64 `json:"debit"`
	Credit             float64 `json:"credit"`
	Memo               string  `json:"memo"`
}

type SupplierInvoiceJournalPreviewResponse struct {
	ReferenceType      string                               `json:"reference_type"`
	ReferenceID        string                               `json:"reference_id"`
	InvoiceDate        string                               `json:"invoice_date"`
	InvoiceNumber      string                               `json:"invoice_number,omitempty"`
	Subtotal           float64                              `json:"subtotal"`
	TaxAmount          float64                              `json:"tax_amount"`
	DownPayment        float64                              `json:"down_payment"`
	TotalAmount        float64                              `json:"total_amount"`
	IsBalanced         bool                                 `json:"is_balanced"`
	ThreeWayMatching   []SupplierInvoiceThreeWayMatchingLine `json:"three_way_matching"`
	Lines              []SupplierInvoiceJournalPreviewLine  `json:"lines"`
}

// Add data DTOs

type SupplierInvoiceAddPaymentTerms struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SupplierInvoiceAddProductMini struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Code     string  `json:"code"`
	ImageURL *string `json:"image_url"`
}

type SupplierInvoiceAddGoodsReceiptItem struct {
	ID                  string                         `json:"id"`
	PurchaseOrderItemID string                         `json:"purchase_order_item_id"`
	Product             *SupplierInvoiceAddProductMini `json:"product,omitempty"`
	QuantityPO          float64                        `json:"quantity_po"`
	QuantityReceived    float64                        `json:"quantity_received"`
	QuantityInvoiced    float64                        `json:"quantity_invoiced"`
	QuantityRemaining   float64                        `json:"quantity_remaining"`
	Price               float64                        `json:"price"`
	SubTotal            float64                        `json:"sub_total"`
}

type SupplierInvoiceAddSupplierMini struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SupplierInvoiceAddDownPaymentMini struct {
	ID            string                            `json:"id"`
	PurchaseOrder *SupplierInvoicePurchaseOrderMini `json:"purchase_order,omitempty"`
	Code          string                            `json:"code"`
	InvoiceNumber string                            `json:"invoice_number"`
	InvoiceDate   string                            `json:"invoice_date"`
	DueDate       string                            `json:"due_date"`
	Amount        float64                           `json:"amount"`
	PaidAmount    float64                           `json:"paid_amount"`
	Status        string                            `json:"status"`
	Notes         *string                           `json:"notes"`
	CreatedAt     time.Time                         `json:"created_at"`
	UpdatedAt     time.Time                         `json:"updated_at"`
}

type SupplierInvoiceAddGoodsReceipt struct {
	ID                      string                               `json:"id"`
	Code                    string                               `json:"code"`
	PurchaseOrder           *SupplierInvoicePurchaseOrderMini    `json:"purchase_order,omitempty"`
	Supplier                *SupplierInvoiceAddSupplierMini      `json:"supplier,omitempty"`
	ReceiptDate             *time.Time                           `json:"receipt_date,omitempty"`
	Status                  string                               `json:"status"`
	Items                   []SupplierInvoiceAddGoodsReceiptItem `json:"items"`
	InvoiceDP               *SupplierInvoiceAddDownPaymentMini   `json:"invoice_dp,omitempty"`
	DefaultPaymentTermsID   *string                              `json:"default_payment_terms_id,omitempty"`
	DefaultPaymentTermsName *string                              `json:"default_payment_terms_name,omitempty"`
}

// SupplierInvoiceAddPurchaseOrderItem kept for backward compat with DP add-data.
type SupplierInvoiceAddPurchaseOrderItem struct {
	ID       string                         `json:"id"`
	Product  *SupplierInvoiceAddProductMini `json:"product,omitempty"`
	Quantity float64                        `json:"quantity"`
	Price    float64                        `json:"price"`
	Subtotal float64                        `json:"subtotal"`
}

// SupplierInvoiceAddPurchaseOrder kept for DP add-data endpoint.
type SupplierInvoiceAddPurchaseOrder struct {
	ID          string                                `json:"id"`
	Supplier    *SupplierInvoiceAddSupplierMini       `json:"supplier,omitempty"`
	Code        string                                `json:"code"`
	OrderDate   string                                `json:"order_date"`
	Status      string                                `json:"status"`
	TotalAmount float64                               `json:"total_amount"`
	Items       []SupplierInvoiceAddPurchaseOrderItem `json:"items"`
	InvoiceDP   *SupplierInvoiceAddDownPaymentMini    `json:"invoice_dp,omitempty"`
}

type SupplierInvoiceAddResponse struct {
	PaymentTerms  []SupplierInvoiceAddPaymentTerms `json:"payment_terms"`
	GoodsReceipts []SupplierInvoiceAddGoodsReceipt `json:"goods_receipts"`
}

// Keep core model imports referenced to avoid unused when generated conditionally.
var _ = coreModels.PaymentTerms{}
var _ = purchaseModels.PurchaseOrder{}
