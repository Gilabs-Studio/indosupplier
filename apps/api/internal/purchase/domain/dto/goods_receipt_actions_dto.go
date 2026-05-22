package dto

type CreateGoodsReceiptRequest struct {
	PurchaseOrderID string                          `json:"purchase_order_id" binding:"required,uuid"`
	WarehouseID     string                          `json:"warehouse_id" binding:"required,uuid"`
	ReceiptDate     *string                         `json:"receipt_date,omitempty" binding:"omitempty,datetime=2006-01-02"`
	Notes           *string                         `json:"notes,omitempty"`
	ProofImageURL   *string                         `json:"proof_image_url,omitempty" binding:"omitempty,max=500"`
	Items           []CreateGoodsReceiptItemRequest `json:"items" binding:"required,dive"`
}

type CreateGoodsReceiptItemRequest struct {
	PurchaseOrderItemID string  `json:"purchase_order_item_id" binding:"required,uuid"`
	ProductID           string  `json:"product_id" binding:"required,uuid"`
	QuantityReceived    float64 `json:"quantity_received" binding:"gte=0"`
	Notes               *string `json:"notes,omitempty"`
}

type UpdateGoodsReceiptRequest struct {
	WarehouseID   string                          `json:"warehouse_id" binding:"required,uuid"`
	ReceiptDate   *string                         `json:"receipt_date,omitempty" binding:"omitempty,datetime=2006-01-02"`
	Notes         *string                         `json:"notes,omitempty"`
	ProofImageURL *string                         `json:"proof_image_url,omitempty" binding:"omitempty,max=500"`
	Items         []CreateGoodsReceiptItemRequest `json:"items" binding:"required,dive"`
}
