package dto

// SupplierResearchKpisRequest filters KPI metrics.
type SupplierResearchKpisRequest struct {
	StartDate        string  `form:"start_date"`
	EndDate          string  `form:"end_date"`
	DateMode         string  `form:"date_mode"`
	Year             int     `form:"year"`
	CategoryIDs      string  `form:"category_ids"`
	MinPurchaseValue float64 `form:"min_purchase_value"`
	MaxPurchaseValue float64 `form:"max_purchase_value"`
}

// SupplierResearchKpisResponse contains top-level KPI cards.
type SupplierResearchKpisResponse struct {
	TotalSuppliers      int     `json:"total_suppliers"`
	ActiveSuppliers     int     `json:"active_suppliers"`
	TotalPurchaseValue  float64 `json:"total_purchase_value"`
	AverageLeadTimeDays float64 `json:"average_lead_time_days"`
}

// ListSupplierPurchaseVolumeRequest filters purchase-volume data.
type ListSupplierPurchaseVolumeRequest struct {
	Search           string  `form:"search"`
	StartDate        string  `form:"start_date"`
	EndDate          string  `form:"end_date"`
	DateMode         string  `form:"date_mode"`
	Year             int     `form:"year"`
	CategoryIDs      string  `form:"category_ids"`
	MinPurchaseValue float64 `form:"min_purchase_value"`
	MaxPurchaseValue float64 `form:"max_purchase_value"`
	Page             int     `form:"page,default=1"`
	PerPage          int     `form:"per_page,default=20"`
	SortBy           string  `form:"sort_by,default=purchase_value"`
	Order            string  `form:"order,default=desc"`
}

// SupplierPurchaseVolumeResponse represents purchase volume by supplier.
type SupplierPurchaseVolumeResponse struct {
	SupplierID          string  `json:"supplier_id"`
	SupplierCode        string  `json:"supplier_code,omitempty"`
	SupplierName        string  `json:"supplier_name"`
	CategoryName        string  `json:"category_name,omitempty"`
	TotalPurchaseValue  float64 `json:"total_purchase_value"`
	TotalPurchaseOrders int     `json:"total_purchase_orders"`
	DependencyScore     float64 `json:"dependency_score"`
}

// ListSupplierDeliveryTimeRequest filters delivery-time data.
type ListSupplierDeliveryTimeRequest struct {
	Search           string  `form:"search"`
	StartDate        string  `form:"start_date"`
	EndDate          string  `form:"end_date"`
	DateMode         string  `form:"date_mode"`
	Year             int     `form:"year"`
	CategoryIDs      string  `form:"category_ids"`
	MinPurchaseValue float64 `form:"min_purchase_value"`
	MaxPurchaseValue float64 `form:"max_purchase_value"`
	Page             int     `form:"page,default=1"`
	PerPage          int     `form:"per_page,default=20"`
	SortBy           string  `form:"sort_by,default=lead_time"`
	Order            string  `form:"order,default=desc"`
}

// SupplierDeliveryTimeResponse represents lead time and on-time metrics by supplier.
type SupplierDeliveryTimeResponse struct {
	SupplierID          string  `json:"supplier_id"`
	SupplierName        string  `json:"supplier_name"`
	AverageLeadTimeDays float64 `json:"average_lead_time_days"`
	SupplierOnTimeRate  float64 `json:"supplier_on_time_rate"`
	LateDeliveryCount   int     `json:"late_delivery_count"`
}

// SupplierSpendTrendRequest filters spend trend chart.
type SupplierSpendTrendRequest struct {
	StartDate        string  `form:"start_date"`
	EndDate          string  `form:"end_date"`
	DateMode         string  `form:"date_mode"`
	Year             int     `form:"year"`
	CategoryIDs      string  `form:"category_ids"`
	MinPurchaseValue float64 `form:"min_purchase_value"`
	MaxPurchaseValue float64 `form:"max_purchase_value"`
	Interval         string  `form:"interval,default=monthly"`
}

// SupplierSpendTrendPointResponse represents one data point in spend trend.
type SupplierSpendTrendPointResponse struct {
	Period             string  `json:"period"`
	TotalPurchaseValue float64 `json:"total_purchase_value"`
}

// SupplierSpendTrendResponse wraps spend trend points.
type SupplierSpendTrendResponse struct {
	Interval string                            `json:"interval"`
	Timeline []SupplierSpendTrendPointResponse `json:"timeline"`
}

// ListSuppliersRequest filters tabular supplier analytics.
type ListSuppliersRequest struct {
	Tab              string  `form:"tab,default=top_spenders"`
	Search           string  `form:"search"`
	StartDate        string  `form:"start_date"`
	EndDate          string  `form:"end_date"`
	DateMode         string  `form:"date_mode"`
	Year             int     `form:"year"`
	CategoryIDs      string  `form:"category_ids"`
	MinPurchaseValue float64 `form:"min_purchase_value"`
	MaxPurchaseValue float64 `form:"max_purchase_value"`
	Page             int     `form:"page,default=1"`
	PerPage          int     `form:"per_page,default=20"`
	SortBy           string  `form:"sort_by"`
	Order            string  `form:"order,default=desc"`
}

// SupplierTableRowResponse is the row shape used by tabbed supplier tables.
type SupplierTableRowResponse struct {
	SupplierID                string  `json:"supplier_id"`
	SupplierCode              string  `json:"supplier_code,omitempty"`
	SupplierName              string  `json:"supplier_name"`
	CategoryName              string  `json:"category_name,omitempty"`
	TotalPurchaseValue        float64 `json:"total_purchase_value"`
	TotalPurchaseOrders       int     `json:"total_purchase_orders"`
	AverageLeadTimeDays       float64 `json:"average_lead_time_days"`
	LateDeliveryCount         int     `json:"late_delivery_count"`
	SupplierOnTimeRate        float64 `json:"supplier_on_time_rate"`
	DependencyScore           float64 `json:"dependency_score"`
	ActivePurchaseOrderCount  int     `json:"active_purchase_order_count"`
}

// SupplierDetailResponse represents supplier details for report page.
type SupplierDetailResponse struct {
	SupplierID          string                                   `json:"supplier_id"`
	SupplierCode        string                                   `json:"supplier_code,omitempty"`
	SupplierName        string                                   `json:"supplier_name"`
	CategoryName        string                                   `json:"category_name,omitempty"`
	TotalPurchaseValue  float64                                  `json:"total_purchase_value"`
	TotalPurchaseOrders int                                      `json:"total_purchase_orders"`
	AverageLeadTimeDays float64                                  `json:"average_lead_time_days"`
	SupplierOnTimeRate  float64                                  `json:"supplier_on_time_rate"`
	LateDeliveryCount   int                                      `json:"late_delivery_count"`
	DependencyScore     float64                                  `json:"dependency_score"`
	Products            []SupplierDetailPurchasedProductResponse `json:"products,omitempty"`
	PurchaseOrders      []SupplierDetailPurchaseOrderResponse    `json:"purchase_orders,omitempty"`
}

// SupplierDetailPurchasedProductResponse is a compact product row used in supplier detail page.
type SupplierDetailPurchasedProductResponse struct {
	ProductID   string  `json:"product_id"`
	ProductCode string  `json:"product_code"`
	ProductName string  `json:"product_name"`
	TotalQty    float64 `json:"total_qty"`
	TotalOrders int     `json:"total_orders"`
	TotalAmount float64 `json:"total_amount"`
}

// SupplierDetailPurchaseOrderResponse is a compact PO row used in supplier detail page.
type SupplierDetailPurchaseOrderResponse struct {
	PurchaseOrderID string  `json:"purchase_order_id"`
	Code            string  `json:"code"`
	Status          string  `json:"status"`
	OrderDate       string  `json:"order_date"`
	TotalAmount     float64 `json:"total_amount"`
}
