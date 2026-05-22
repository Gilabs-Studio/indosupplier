package dto

type GetStockMovementsRequest struct {
	Page        int    `json:"page" form:"page"`
	PerPage     int    `json:"per_page" form:"per_page"`
	Search      string `json:"search" form:"search"`
	WarehouseID string `json:"warehouse_id" form:"warehouse_id"`
	ProductID   string `json:"product_id" form:"product_id"`
	Type        string `json:"type" form:"type"`
	StartDate   string `json:"start_date" form:"start_date"`
	EndDate     string `json:"end_date" form:"end_date"`
}

// StockMovementRequest represents the request to create a stock movement
type StockMovementRequest struct {
	InventoryBatchID string  `json:"inventory_batch_id" binding:"omitempty,uuid"`
	ProductID        string  `json:"product_id" binding:"required,uuid"`
	WarehouseID      string  `json:"warehouse_id" binding:"required,uuid"`
	Type             string  `json:"type" binding:"required,oneof=IN OUT ADJUST TRANSFER"`
	Quantity         float64 `json:"quantity" binding:"required,gt=0"`
	ReferenceType     string  `json:"reference_type" binding:"required,oneof=PO DO OPNAME TRANSFER GOODS_RECEIPT DELIVERY_ORDER STOCK_OPNAME INVENTORY_ADJUSTMENT"`
	ReferenceID      string  `json:"reference_id" binding:"required,uuid"`
	ReferenceNumber  string  `json:"reference_number" binding:"required"`
	Source           string  `json:"source"`
	Description       string  `json:"description"`
	Cost              float64 `json:"cost"`
	CreatedBy         *string `json:"created_by"`
	SkipJournaling    bool    `json:"-"`
	MovementDirection string  `json:"-"`
}

type CreateManualMovementRequest struct {
	ProductID         string   `json:"product_id" binding:"required,uuid"`
	WarehouseID       string   `json:"warehouse_id" binding:"required,uuid"`
	TargetWarehouseID *string `json:"target_warehouse_id"` // required for TRANSFER, handled in usecase
	Type              string   `json:"type" binding:"required,oneof=IN OUT ADJUST TRANSFER"`
	Quantity          float64  `json:"quantity" binding:"required,gt=0"`
	BatchID           *string  `json:"batch_id"`  // Optional: specific batch to deduct (for OUT/TRANSFER)
	ReferenceNumber   string   `json:"reference_number"`
	Description       string   `json:"description"`
	CreatedBy         string   `json:"-"`
	MovementDirection string   `json:"-"`
}
