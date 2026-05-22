package dto

// AdjustStockItem represents a single product variance from Stock Opname
type AdjustStockItem struct {
	ProductID   string  `json:"product_id"`
	VarianceQty float64 `json:"variance_qty"` // Positive = surplus, Negative = shortage
	BatchID     *string `json:"batch_id,omitempty"`
}

// AdjustStockFromOpnameRequest contains all data needed to create ADJUST movements from a posted Stock Opname
type AdjustStockFromOpnameRequest struct {
	OpnameID     string            `json:"opname_id"`
	OpnameNumber string            `json:"opname_number"`
	WarehouseID  string            `json:"warehouse_id"`
	Items        []AdjustStockItem `json:"items"`
	PostedBy     string            `json:"posted_by"`
	Notes        string            `json:"notes"`
}
