package dto

import "time"

type StockOpnameStatus string

const (
	StockOpnameStatusDraft           StockOpnameStatus = "draft"
	StockOpnameStatusPendingApproval StockOpnameStatus = "pending_approval"
	StockOpnameStatusPending         StockOpnameStatus = "pending"
	StockOpnameStatusApproved        StockOpnameStatus = "approved"
	StockOpnameStatusCompleted       StockOpnameStatus = "completed"
	StockOpnameStatusRejected        StockOpnameStatus = "rejected"
	StockOpnameStatusPosted          StockOpnameStatus = "posted"
)

type CreateStockOpnameRequest struct {
	WarehouseID  string  `json:"warehouse_id" validate:"required,uuid"`
	Date         string  `json:"date" validate:"required,datetime=2006-01-02"`
	Description  string  `json:"description"`
	ScopeType    string   `json:"scope_type" validate:"omitempty,oneof=all category brand"`
	CategoryIDs  []string `json:"category_ids" validate:"omitempty,dive,uuid"`
	BrandIDs     []string `json:"brand_ids" validate:"omitempty,dive,uuid"`
	OrderedByID  *string `json:"ordered_by_id" validate:"omitempty,uuid"`
	AssignedToID *string `json:"assigned_to_id" validate:"omitempty,uuid"`
}

type UpdateStockOpnameRequest struct {
	Date         *string `json:"date" validate:"omitempty,datetime=2006-01-02"`
	Description  *string `json:"description"`
	OrderedByID  *string `json:"ordered_by_id" validate:"omitempty,uuid"`
	AssignedToID *string `json:"assigned_to_id" validate:"omitempty,uuid"`
}

type SaveStockOpnameItemsRequest struct {
	Items []StockOpnameItemRequest `json:"items" validate:"required,dive"`
}

type StockOpnameItemRequest struct {
	ProductID        string   `json:"product_id" validate:"required,uuid"`
	SystemQty        float64  `json:"system_qty"`
	PhysicalQty      *float64 `json:"physical_qty"`
	Notes            string   `json:"notes"`
	InventoryBatchID *string  `json:"inventory_batch_id"`
	BatchNumber      string   `json:"batch_number"`
	BatchQty         float64  `json:"batch_qty"`
}

type StockOpnameResponse struct {
	ID                       string            `json:"id"`
	OpnameNumber             string            `json:"opname_number"`
	WarehouseID              string            `json:"warehouse_id"`
	WarehouseName            string            `json:"warehouse_name,omitempty"`
	JournalID                *string           `json:"journal_id,omitempty"`
	Date                     time.Time         `json:"date"`
	Status                   StockOpnameStatus `json:"status"`
	Description              string            `json:"description"`
	TotalItems               int               `json:"total_items"`
	TotalVarianceQty         float64           `json:"total_variance_qty"`
	TotalNegativeVarianceQty float64           `json:"total_negative_variance_qty"`
	TotalPositiveVarianceQty float64           `json:"total_positive_variance_qty"`
	OrderedByID              *string           `json:"ordered_by_id,omitempty"`
	OrderedByName            *string           `json:"ordered_by_name,omitempty"`
	AssignedToID             *string           `json:"assigned_to_id,omitempty"`
	AssignedToName           *string           `json:"assigned_to_name,omitempty"`
	CreatedBy                *string           `json:"created_by"`
	CreatedByName            string            `json:"created_by_name,omitempty"`
	CreatedAt                time.Time         `json:"created_at"`
	UpdatedAt                time.Time         `json:"updated_at"`
}

type StockOpnameItemResponse struct {
	ID               string   `json:"id"`
	StockOpnameID    string   `json:"stock_opname_id"`
	ProductID        string   `json:"product_id"`
	ProductName      string   `json:"product_name,omitempty"`
	ProductCode      string   `json:"product_code,omitempty"`
	ProductImageURL  *string  `json:"product_image_url,omitempty"`
	SystemQty        float64  `json:"system_qty"`
	PhysicalQty      *float64 `json:"physical_qty"`
	VarianceQty      float64  `json:"variance_qty"`
	UnitCost         float64  `json:"unit_cost"`
	Notes            string   `json:"notes"`
	InventoryBatchID *string  `json:"inventory_batch_id,omitempty"`
	BatchNumber      string   `json:"batch_number,omitempty"`
	BatchQty         float64  `json:"batch_qty,omitempty"`
}

type ListStockOpnamesRequest struct {
	Page        int    `form:"page"`
	PerPage     int    `form:"per_page"`
	Search      string `form:"search"`
	WarehouseID string `form:"warehouse_id"`
	Status      string `form:"status"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
}

type ListStockOpnameItemsRequest struct {
	Page    int `form:"page" query:"page"`
	PerPage int `form:"per_page" query:"per_page"`
}

type WarehouseStockSnapshot struct {
	ProductID   string
	SystemQty   float64
	BatchID     string
	BatchNumber string
	BatchQty    float64
}

// UserWarehouseInfo holds minimal warehouse data for a user's assigned warehouse.
type UserWarehouseInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type UpdateStockOpnameStatusRequest struct {
	Status string `json:"status" validate:"required"`
}
