package dto

import "time"

type GoodsReceiptListResponse struct {
	ID                 string                         `json:"id"`
	Code               string                         `json:"code"`
	CompanyID          *string                        `json:"company_id,omitempty"`
	FiscalYearID       *string                        `json:"fiscal_year_id,omitempty"`
	PurchaseOrder      *GoodsReceiptPurchaseOrderMini `json:"purchase_order,omitempty"`
	Warehouse          *GoodsReceiptWarehouseMini     `json:"warehouse,omitempty"`
	WarehouseID        *string                        `json:"warehouse_id,omitempty"`
	Supplier           *GoodsReceiptSupplierMini      `json:"supplier,omitempty"`
	ReceiptDate        *string                        `json:"receipt_date,omitempty"`
	Notes              *string                        `json:"notes,omitempty"`
	ProofImageURL      *string                        `json:"proof_image_url,omitempty"`
	Status             string                         `json:"status"`
	CreatedBy          string                         `json:"created_by"`
	TotalItemsReceived float64                        `json:"total_items_received"`

	SubmittedAt                  *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt                   *time.Time `json:"approved_at,omitempty"`
	ClosedAt                     *time.Time `json:"closed_at,omitempty"`
	RejectedAt                   *time.Time `json:"rejected_at,omitempty"`
	ConvertedAt                  *time.Time `json:"converted_at,omitempty"`
	ConvertedToSupplierInvoiceID *string    `json:"converted_to_supplier_invoice_id,omitempty"`
	JournalEntryID               *string    `json:"journal_entry_id,omitempty"`
}

type GoodsReceiptPurchaseOrderMini struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type GoodsReceiptSupplierMini struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GoodsReceiptWarehouseMini struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GoodsReceiptDetailResponse struct {
	ID            string                           `json:"id"`
	Code          string                           `json:"code"`
	CompanyID     *string                          `json:"company_id,omitempty"`
	FiscalYearID  *string                          `json:"fiscal_year_id,omitempty"`
	PurchaseOrder *GoodsReceiptPurchaseOrderDetail `json:"purchase_order,omitempty"`
	Warehouse     *GoodsReceiptWarehouseMini       `json:"warehouse,omitempty"`
	WarehouseID   *string                          `json:"warehouse_id,omitempty"`
	Supplier      *GoodsReceiptSupplierMini        `json:"supplier,omitempty"`
	ReceiptDate   *string                          `json:"receipt_date,omitempty"`
	Notes         *string                          `json:"notes,omitempty"`
	ProofImageURL *string                          `json:"proof_image_url,omitempty"`
	Status        string                           `json:"status"`
	CreatedBy     string                           `json:"created_by"`
	Items         []GoodsReceiptItemResponse       `json:"items"`

	SubmittedAt                  *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt                   *time.Time `json:"approved_at,omitempty"`
	ClosedAt                     *time.Time `json:"closed_at,omitempty"`
	RejectedAt                   *time.Time `json:"rejected_at,omitempty"`
	ConvertedAt                  *time.Time `json:"converted_at,omitempty"`
	ConvertedToSupplierInvoiceID *string    `json:"converted_to_supplier_invoice_id,omitempty"`
	JournalEntryID               *string    `json:"journal_entry_id,omitempty"`
}

type GoodsReceiptPurchaseOrderDetail struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Status string `json:"status"`
}

type GoodsReceiptItemResponse struct {
	ID                  string       `json:"id"`
	PurchaseOrderItemID string       `json:"purchase_order_item_id"`
	Product             *ProductMini `json:"product,omitempty"`
	QuantityReceived    float64      `json:"quantity_received"`
	Notes               *string      `json:"notes,omitempty"`
}

type ProductMini struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	SKU  *string `json:"sku,omitempty"`
}

// GoodsReceiptConvertResponse is returned when a GR is converted to a Supplier Invoice.
type GoodsReceiptConvertResponse struct {
	GoodsReceiptID    string `json:"goods_receipt_id"`
	SupplierInvoiceID string `json:"supplier_invoice_id"`
}
