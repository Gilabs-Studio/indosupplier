package dto

import "time"

type ReturnOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type ReturnWarehouseOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PurchaseReturnFormDataResponse struct {
	Warehouses     []ReturnWarehouseOption `json:"warehouses"`
	ReturnReasons  []ReturnOption          `json:"return_reasons"`
	ItemConditions []ReturnOption          `json:"item_conditions"`
	Actions        []ReturnOption          `json:"actions"`
}

type CreatePurchaseReturnItemRequest struct {
	GoodsReceiptItemID *string `json:"goods_receipt_item_id"`
	ProductID          string  `json:"product_id" binding:"required"`
	UOMID              *string `json:"uom_id"`
	Condition          string  `json:"condition" binding:"required"`
	Notes              *string `json:"notes"`
	Qty                float64 `json:"qty" binding:"required,gt=0"`
	UnitCost           float64 `json:"unit_cost" binding:"required,gte=0"`
}

type CreatePurchaseReturnRequest struct {
	GoodsReceiptID  string                            `json:"goods_receipt_id" binding:"required"`
	PurchaseOrderID *string                           `json:"purchase_order_id"`
	SupplierID      string                            `json:"supplier_id"`
	WarehouseID     string                            `json:"warehouse_id" binding:"required"`
	Reason          string                            `json:"reason" binding:"required"`
	Action          string                            `json:"action" binding:"required"`
	Notes           *string                           `json:"notes"`
	Items           []CreatePurchaseReturnItemRequest `json:"items" binding:"required,min=1"`
}

type UpdatePurchaseReturnItemRequest struct {
	GoodsReceiptItemID *string `json:"goods_receipt_item_id"`
	ProductID          string  `json:"product_id" binding:"required"`
	UOMID              *string `json:"uom_id"`
	Condition          string  `json:"condition" binding:"required"`
	Notes              *string `json:"notes"`
	Qty                float64 `json:"qty" binding:"required,gt=0"`
	UnitCost           float64 `json:"unit_cost" binding:"required,gte=0"`
}

type UpdatePurchaseReturnRequest struct {
	WarehouseID string                           `json:"warehouse_id" binding:"required"`
	Reason        string                           `json:"reason" binding:"required"`
	Action        string                           `json:"action" binding:"required"`
	Notes         *string                          `json:"notes"`
	Items         []UpdatePurchaseReturnItemRequest `json:"items" binding:"required,min=1"`
}

type UpdatePurchaseReturnStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type PurchaseReturnItemResponse struct {
	ID                 string  `json:"id"`
	GoodsReceiptItemID *string `json:"goods_receipt_item_id,omitempty"`
	ProductID          string  `json:"product_id"`
	UOMID              *string `json:"uom_id,omitempty"`
	Condition          string  `json:"condition"`
	Notes              *string `json:"notes,omitempty"`
	Qty                float64 `json:"qty"`
	UnitCost           float64 `json:"unit_cost"`
	Subtotal           float64 `json:"subtotal"`
}

type PurchaseReturnResponse struct {
	ID                string                       `json:"id"`
	Code              string                       `json:"return_number"`
	GoodsReceiptID    string                       `json:"goods_receipt_id"`
	PurchaseOrderID   *string                      `json:"purchase_order_id,omitempty"`
	SupplierID        string                       `json:"supplier_id"`
	CompanyID         *string                      `json:"company_id"`
	FiscalYearID      *string                      `json:"fiscal_year_id"`
	WarehouseID       string                       `json:"warehouse_id"`
	Reason            string                       `json:"reason"`
	Action            string                       `json:"action"`
	Status            string                       `json:"status"`
	Notes             *string                      `json:"notes,omitempty"`
	TotalAmount       float64                      `json:"total_amount"`
	StockAdjustmentID *string                      `json:"stock_adjustment_id,omitempty"`
	DebitNoteID       *string                      `json:"debit_note_id,omitempty"`
	Items             []PurchaseReturnItemResponse `json:"items"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

type PurchaseReturnAuditTrailEntry struct {
	ID             string                 `json:"id"`
	Action         string                 `json:"action"`
	PermissionCode string                 `json:"permission_code"`
	TargetID       string                 `json:"target_id"`
	Metadata       map[string]interface{} `json:"metadata"`
	User           *AuditTrailUser        `json:"user"`
	CreatedAt      time.Time              `json:"created_at"`
}
