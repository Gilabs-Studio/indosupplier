package dto

import "time"

type ReceiveStockItem struct {
	ProductID   string     `json:"product_id"`
	Quantity    float64    `json:"quantity"`
	CostPrice   float64    `json:"cost_price"`   // Price per unit from PO
	BatchNumber *string    `json:"batch_number"` // Optional override
	ExpiryDate  *time.Time `json:"expiry_date"`
}

type ReceiveStockRequest struct {
	SourceID     string             `json:"source_id"`     // GR ID
	SourceNumber string             `json:"source_number"` // GR Code
	SourceType   string             `json:"source_type"`   // "GR"
	Source       string             `json:"source"`
	WarehouseID  string             `json:"warehouse_id"`  // Default warehouse? Or per item? purchase usually implies receiving to a warehouse.
	Items        []ReceiveStockItem `json:"items"`
	ReceivedAt   time.Time          `json:"received_at"`
	ReceivedBy   string             `json:"received_by"`
	Notes        string             `json:"notes"`
}
