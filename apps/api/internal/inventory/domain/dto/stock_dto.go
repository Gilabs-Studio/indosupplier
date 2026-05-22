package dto

import "time"

// InventoryStockItem represents the aggregated stock view for the list
type InventoryStockItem struct {
	ProductID       string  `json:"product_id"`
	ProductCode     string  `json:"product_code"`
	ProductName     string  `json:"product_name"`
	ProductImageURL *string `json:"product_image_url"`
	ProductCategory *string `json:"product_category"`
	ProductBrand    *string `json:"product_brand"`

	WarehouseID   string `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`

	OnHand    float64 `json:"on_hand"`
	Reserved  float64 `json:"reserved"`
	Available float64 `json:"available"`

	MinStock float64 `json:"min_stock"`
	MaxStock float64 `json:"max_stock"`
	UomName  string  `json:"uom_name"`

	Status             string `json:"status"` // "ok", "low", "overstock", "out_of_stock"
	HasExpiringBatches bool   `json:"has_expiring_batches"`
	// IsIngredient indicates this product is a raw material used in F&B recipes
	IsIngredient       bool   `json:"is_ingredient" gorm:"column:is_ingredient"`
}

type GetInventoryListRequest struct {
	Page        int    `form:"page"`
	PerPage     int    `form:"per_page"`
	Search      string `form:"search"`
	WarehouseID string `form:"warehouse_id"`
	ProductID   string `form:"product_id"`
	// LowStock is a legacy shorthand for Status=low_stock; kept for backward compat
	LowStock    bool   `form:"low_stock"`
	// Status filters items by stock status: ok | low_stock | out_of_stock | overstock
	Status      string `form:"status"`
	// HasExpiring filters items that have at least one batch expiring within 30 days
	HasExpiring bool   `form:"has_expiring"`
	// HasExpired filters items that have at least one expired batch still holding quantity
	HasExpired  bool   `form:"has_expired"`
	// IsIngredient filters items where the product is flagged as an ingredient (raw material)
	IsIngredient *bool  `form:"is_ingredient"`
}

type GetInventoryListResponse struct {
	Data []InventoryStockItem `json:"data"`
	Meta PaginationMeta       `json:"meta"`
}

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type InventoryBatchItem struct {
	ID               string     `json:"id"`
	BatchNumber      string     `json:"batch_number"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	ReceivedAt       *time.Time `json:"received_at"`
	CurrentQuantity  float64    `json:"current_quantity"`
	ReservedQuantity float64    `json:"reserved_quantity"`
	Available        float64    `json:"available"`
}

// CreateBatchParams used for repository creation
type CreateBatchParams struct {
	ProductID       string
	WarehouseID     string
	BatchNumber     string
	ExpiryDate      *time.Time
	InitialQuantity float64
	CostPrice       float64
	ReceivedAt      time.Time
}

// InventoryBatchDetail provides detailed batch info for validation
type InventoryBatchDetail struct {
	ID               string     `json:"id"`
	ProductID        string     `json:"product_id"`
	WarehouseID      string     `json:"warehouse_id"`
	BatchNumber      string     `json:"batch_number"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	CurrentQuantity  float64    `json:"current_quantity"`
	ReservedQuantity float64    `json:"reserved_quantity"`
	CostPrice        float64    `json:"cost_price"`
	Available        float64    `json:"available"`
	IsActive         bool       `json:"is_active"`
}

// Tree View DTOs

type GetInventoryTreeWarehousesResponse struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Summary StockSummary `json:"summary"`
}

type StockSummary struct {
	TotalItems int `json:"total_items"`
	Ok         int `json:"ok"`
	Low        int `json:"low"`
	OutOfStock int `json:"out_of_stock"`
	Overstock  int `json:"overstock"`
}

type GetInventoryTreeProductsRequest struct {
	WarehouseID  string `form:"warehouse_id" binding:"required"`
	Page         int    `form:"page"`
	PerPage      int    `form:"per_page"`
	Search       string `form:"search"`
	// IsIngredient filters products flagged as raw material ingredients
	IsIngredient *bool  `form:"is_ingredient"`
}


type TreeProductsSummary struct {
	TotalItems int `json:"total_items"`
	Ok         int `json:"ok"`
	Low        int `json:"low"`
	OutOfStock int `json:"out_of_stock"`
	Overstock  int `json:"overstock"`
}

type GetInventoryTreeProductsResponse struct {
	Data    []InventoryStockItem `json:"data"`
	Meta    PaginationMeta       `json:"meta"`
	Summary TreeProductsSummary  `json:"summary"`
}
type GetInventoryTreeBatchesRequest struct {
	WarehouseID string `form:"warehouse_id" binding:"required"`
	ProductID   string `form:"product_id" binding:"required"`
	Page        int    `form:"page"`
	PerPage     int    `form:"per_page"`
}

type GetInventoryTreeBatchesResponse struct {
	Data []InventoryBatchItem `json:"data"`
	Meta PaginationMeta       `json:"meta"`
}

// InventoryMetrics provides a high-level summary for owner/admin dashboards
type InventoryMetrics struct {
	TotalItems           int     `json:"total_items"`            // Unique product-warehouse combinations
	TotalProducts        int     `json:"total_products"`         // Unique products with stock
	TotalWarehouses      int     `json:"total_warehouses"`       // Warehouses with active inventory
	TotalOnHand          float64 `json:"total_on_hand"`          // Sum of all on-hand units
	OkCount              int     `json:"ok_count"`               // Items within healthy range
	LowStockCount        int     `json:"low_stock_count"`        // Items at or below min_stock
	OutOfStockCount      int     `json:"out_of_stock_count"`     // Items with zero available
	OverstockCount       int     `json:"overstock_count"`        // Items exceeding max_stock
	ExpiringBatches30Day int     `json:"expiring_batches_30_day"` // Batches expiring within 30 days
	ExpiredBatches       int     `json:"expired_batches"`        // Expired batches still with qty > 0
}
