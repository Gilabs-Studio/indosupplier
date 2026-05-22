package dto

import "time"

type GoodsReceiptAddResponse struct {
	EligiblePurchaseOrders []GoodsReceiptPurchaseOrderOption `json:"eligible_purchase_orders"`
}

type GoodsReceiptPurchaseOrderOption struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Status string `json:"status"`
	Supplier *GoodsReceiptSupplierMini `json:"supplier,omitempty"`
}

type GoodsReceiptAuditTrailEntry struct {
	ID             string                 `json:"id"`
	Action         string                 `json:"action"`
	PermissionCode string                 `json:"permission_code"`
	TargetID       string                 `json:"target_id"`
	Metadata       map[string]interface{} `json:"metadata"`
	User           *AuditTrailUser        `json:"user,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}
