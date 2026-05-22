package dto

import "time"

type GetProductStockLedgersRequest struct {
	Page            int        `form:"page"`
	Limit           int        `form:"limit"`
	TransactionType string     `form:"transaction_type"`
	DateFrom        string     `form:"date_from"`
	DateTo          string     `form:"date_to"`
	ParsedDateFrom  *time.Time `form:"-" json:"-"`
	ParsedDateTo    *time.Time `form:"-" json:"-"`
}

type ProductStockLedgerItem struct {
	ID                   string    `json:"id"`
	ProductID            string    `json:"product_id"`
	TransactionID        string    `json:"transaction_id"`
	TransactionType      string    `json:"transaction_type"`
	TransactionTypeLabel string    `json:"transaction_type_label"`
	Qty                  float64   `json:"qty"`
	UnitCost             float64   `json:"unit_cost"`
	AverageCost          float64   `json:"average_cost"`
	StockValue           float64   `json:"stock_value"`
	RunningQty           float64   `json:"running_qty"`
	CreatedAt            time.Time `json:"created_at"`
}

type GetProductStockLedgersResponse struct {
	Data []ProductStockLedgerItem `json:"data"`
	Meta PaginationMeta           `json:"meta"`
}
