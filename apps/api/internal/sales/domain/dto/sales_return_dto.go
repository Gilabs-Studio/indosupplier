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

type SalesReturnFormDataResponse struct {
	Warehouses     []ReturnWarehouseOption `json:"warehouses"`
	ReturnReasons  []ReturnOption          `json:"return_reasons"`
	ItemConditions []ReturnOption          `json:"item_conditions"`
	Actions        []ReturnOption          `json:"actions"`
	RefundMethods  []ReturnOption          `json:"refund_methods"`
}

type CreateSalesReturnItemRequest struct {
	InvoiceItemID *string `json:"invoice_item_id"`
	ProductID     string  `json:"product_id" binding:"required,uuid"`
	UOMID         *string `json:"uom_id"`
	Condition     string  `json:"condition" binding:"required"`
	Notes         *string `json:"notes"`
	Qty           float64 `json:"qty" binding:"required,gt=0"`
	UnitPrice     float64 `json:"unit_price" binding:"required,gte=0"`
}

type CreateSalesReturnRequest struct {
	InvoiceID   *string                        `json:"invoice_id"`
	DeliveryID  *string                        `json:"delivery_id"`
	WarehouseID string                         `json:"warehouse_id" binding:"required"`
	CustomerID  string                         `json:"customer_id"`
	Reason      string                         `json:"reason" binding:"required"`
	Action      string                         `json:"action" binding:"required"`
	Notes       *string                        `json:"notes"`
	Items       []CreateSalesReturnItemRequest `json:"items" binding:"required,min=1"`
}

type UpdateSalesReturnItemRequest struct {
	InvoiceItemID *string `json:"invoice_item_id"`
	ProductID     string  `json:"product_id" binding:"required,uuid"`
	UOMID         *string `json:"uom_id"`
	Condition     string  `json:"condition" binding:"required"`
	Notes         *string `json:"notes"`
	Qty           float64 `json:"qty" binding:"required,gt=0"`
	UnitPrice     float64 `json:"unit_price" binding:"required,gte=0"`
}

type UpdateSalesReturnRequest struct {
	InvoiceID   *string                        `json:"invoice_id"`
	WarehouseID string                         `json:"warehouse_id" binding:"required"`
	CustomerID  string                         `json:"customer_id"`
	Reason      string                         `json:"reason" binding:"required"`
	Action      string                         `json:"action" binding:"required"`
	Notes       *string                        `json:"notes"`
	Items       []UpdateSalesReturnItemRequest `json:"items" binding:"required,min=1"`
}

type UpdateSalesReturnStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type SalesReturnItemResponse struct {
	ID            string  `json:"id"`
	InvoiceItemID *string `json:"invoice_item_id,omitempty"`
	ProductID     string  `json:"product_id"`
	UOMID         *string `json:"uom_id,omitempty"`
	Condition     string  `json:"condition"`
	Notes         *string `json:"notes,omitempty"`
	Qty           float64 `json:"qty"`
	UnitPrice     float64 `json:"unit_price"`
	Subtotal      float64 `json:"subtotal"`
}

type SalesReturnResponse struct {
	ID                string                    `json:"id"`
	Code              string                    `json:"return_number"`
	InvoiceID         *string                   `json:"invoice_id,omitempty"`
	DeliveryID        *string                   `json:"delivery_id,omitempty"`
	WarehouseID       string                    `json:"warehouse_id"`
	CustomerID        string                    `json:"customer_id"`
	Reason            string                    `json:"reason"`
	Action            string                    `json:"action"`
	Status            string                    `json:"status"`
	Notes             *string                   `json:"notes,omitempty"`
	TotalAmount       float64                   `json:"total_amount"`
	StockAdjustmentID *string                   `json:"stock_adjustment_id,omitempty"`
	CreditNoteID      *string                   `json:"credit_note_id,omitempty"`
	Items             []SalesReturnItemResponse `json:"items"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
}
