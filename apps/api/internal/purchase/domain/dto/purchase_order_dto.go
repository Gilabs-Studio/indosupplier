package dto

import "time"

// PurchaseOrderPartySummary is a minimal supplier snapshot used in list responses.
type PurchaseOrderPartySummary struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// PurchaseOrderRequsitionRef is a minimal PR reference.
type PurchaseOrderRequisitionRef struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

// GoodsReceiptSummary is a minimal GR view used in list responses.
type GoodsReceiptSummary struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	Status     string    `json:"status"`
	TotalItems int       `json:"total_items"`
	// TotalItemsReceived is the sum of QuantityReceived across the GR items
	TotalItemsReceived float64   `json:"total_items_received"`
	CreatedAt          time.Time `json:"created_at"`
}

// POFulfillmentSummary tracks how much of the ordered qty has been received via confirmed GRs.
type POFulfillmentSummary struct {
	TotalOrdered  float64 `json:"total_ordered"`
	TotalReceived float64 `json:"total_received"`
	TotalPending  float64 `json:"total_pending"`
	TotalRemaining float64 `json:"total_remaining"`
}

// SupplierInvoiceSummary is a minimal Supplier Invoice view used in list responses.
type SupplierInvoiceSummary struct {
	ID               string    `json:"id"`
	Code             string    `json:"code"`
	Status           string    `json:"status"`
	Amount           float64   `json:"amount"`
	PaidAmount       float64   `json:"paid_amount"`
	GoodsReceiptID   *string   `json:"goods_receipt_id,omitempty"`
	GoodsReceiptCode *string   `json:"goods_receipt_code,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type PurchaseOrderItemRequest struct {
	ProductID string   `json:"product_id" binding:"required,uuid"`
	Quantity  float64  `json:"quantity" binding:"required,gt=0"`
	Price     float64  `json:"price" binding:"required,gte=0"`
	Discount  float64  `json:"discount" binding:"gte=0,lte=100"`
	Notes     *string  `json:"notes"`
}

type CreatePurchaseOrderRequest struct {
	SupplierID           *string `json:"supplier_id" binding:"omitempty,uuid"`
	PaymentTermsID       *string `json:"payment_terms_id" binding:"omitempty,uuid"`
	BusinessUnitID       *string `json:"business_unit_id" binding:"omitempty,uuid"`
	PurchaseRequisitionID *string `json:"purchase_requisitions_id" binding:"omitempty,uuid"`
	SalesOrderID         *string `json:"sales_order_id" binding:"omitempty,uuid"`

	OrderDate string  `json:"order_date" binding:"required"`
	DueDate   *string `json:"due_date"`

	TaxRate      float64 `json:"tax_rate" binding:"gte=0,lte=100"`
	DeliveryCost float64 `json:"delivery_cost" binding:"gte=0"`
	OtherCost    float64 `json:"other_cost" binding:"gte=0"`

	Notes string `json:"notes"`

	Items []PurchaseOrderItemRequest `json:"items" binding:"required,min=1"`
}

type UpdatePurchaseOrderRequest struct {
	SupplierID     *string `json:"supplier_id" binding:"omitempty,uuid"`
	PaymentTermsID *string `json:"payment_terms_id" binding:"omitempty,uuid"`
	BusinessUnitID *string `json:"business_unit_id" binding:"omitempty,uuid"`

	OrderDate string  `json:"order_date" binding:"required"`
	DueDate   *string `json:"due_date"`

	TaxRate      float64 `json:"tax_rate" binding:"gte=0,lte=100"`
	DeliveryCost float64 `json:"delivery_cost" binding:"gte=0"`
	OtherCost    float64 `json:"other_cost" binding:"gte=0"`

	Notes string `json:"notes"`

	Items []PurchaseOrderItemRequest `json:"items" binding:"required,min=1"`
}

type RevisePurchaseOrderRequest struct {
	RevisionComment string `json:"revision_comment" binding:"required"`
}

// Responses

type PurchaseOrderItemResponse struct {
	ID                string      `json:"id"`
	ProductID         string      `json:"product_id"`
	Quantity          float64     `json:"quantity"`
	Price             float64     `json:"price"`
	Discount          float64     `json:"discount"`
	Subtotal          float64     `json:"subtotal"`
	Notes             *string     `json:"notes"`
	Product           interface{} `json:"product,omitempty"`
	QuantityReceived  float64     `json:"quantity_received"`
	QuantityRemaining float64     `json:"quantity_remaining"`
}

type PurchaseOrderListResponse struct {
	ID                  string                       `json:"id"`
	Code                string                       `json:"code"`
	CompanyID           *string                      `json:"company_id,omitempty"`
	FiscalYearID        *string                      `json:"fiscal_year_id,omitempty"`
	OrderDate           string                       `json:"order_date"`
	DueDate             *string                      `json:"due_date"`
	Status              string                       `json:"status"`
	TotalAmount         float64                      `json:"total_amount"`
	Supplier            *PurchaseOrderPartySummary   `json:"supplier,omitempty"`
	PurchaseRequisition *PurchaseOrderRequisitionRef `json:"purchase_requisition,omitempty"`
	GoodsReceipts       []GoodsReceiptSummary        `json:"goods_receipts,omitempty"`
	SupplierInvoices    []SupplierInvoiceSummary     `json:"supplier_invoices,omitempty"`
	Fulfillment         *POFulfillmentSummary        `json:"fulfillment,omitempty"`
}

type PurchaseOrderDetailResponse struct {
	ID         string      `json:"id"`
	Code       string      `json:"code"`
	CompanyID   *string    `json:"company_id,omitempty"`
	FiscalYearID *string   `json:"fiscal_year_id,omitempty"`
	SupplierID *string     `json:"supplier_id"`
	PaymentTermsID *string `json:"payment_terms_id"`
	BusinessUnitID *string `json:"business_unit_id"`
	CreatedBy  string      `json:"created_by"`
	PurchaseRequisitionID *string `json:"purchase_requisitions_id"`
	SalesOrderID *string   `json:"sales_order_id"`
	OrderDate  string      `json:"order_date"`
	DueDate    *string     `json:"due_date"`
	RevisionComment *string `json:"revision_comment"`
	Notes      string      `json:"notes"`
	Status     string      `json:"status"`
	TaxRate      float64   `json:"tax_rate"`
	TaxAmount    float64   `json:"tax_amount"`
	DeliveryCost float64   `json:"delivery_cost"`
	OtherCost    float64   `json:"other_cost"`
	SubTotal     float64   `json:"sub_total"`
	TotalAmount  float64   `json:"total_amount"`
	Supplier            interface{}                 `json:"supplier,omitempty"`
	PaymentTerms        interface{}                 `json:"payment_terms,omitempty"`
	BusinessUnit        interface{}                 `json:"business_unit,omitempty"`
	Creator             interface{}                 `json:"creator,omitempty"`
	PurchaseRequisition interface{}                 `json:"purchase_requisition,omitempty"`
	Items               []PurchaseOrderItemResponse `json:"items"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	SubmittedAt         *time.Time                  `json:"submitted_at"`
	ApprovedAt          *time.Time                  `json:"approved_at"`
	ClosedAt            *time.Time                  `json:"closed_at"`
}
