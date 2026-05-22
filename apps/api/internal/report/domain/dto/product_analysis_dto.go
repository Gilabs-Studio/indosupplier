package dto

// --- Product Performance List ---

// ListProductPerformanceRequest filters the product performance list
type ListProductPerformanceRequest struct {
	Search     string `form:"search"`
	CategoryID string `form:"category_id"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page,default=1"`
	PerPage    int    `form:"per_page,default=20"`
	SortBy     string `form:"sort_by,default=revenue"`
	Order      string `form:"order,default=desc"`
}

// ProductPerformanceResponse represents a product's aggregated sales performance
type ProductPerformanceResponse struct {
	ProductID             string  `json:"product_id"`
	ProductCode           string  `json:"product_code"`
	ProductName           string  `json:"product_name"`
	ProductSKU            string  `json:"product_sku,omitempty"`
	ProductImage          string  `json:"product_image,omitempty"`
	CategoryName          string  `json:"category_name,omitempty"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
	AvgPrice              float64 `json:"avg_price"`
	AvgPriceFormatted     string  `json:"avg_price_formatted"`
}

// --- Monthly Product Sales ---

// MonthlyProductSalesRequest filters the monthly product sales chart
type MonthlyProductSalesRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// MonthlyProductSalesDataResponse represents a single month aggregation
type MonthlyProductSalesDataResponse struct {
	Month        int     `json:"month"`
	MonthName    string  `json:"month_name"`
	Year         int     `json:"year"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalQty     float64 `json:"total_qty"`
	TotalOrders  int     `json:"total_orders"`
}

// MonthlyProductSalesResponse wraps monthly data with totals
type MonthlyProductSalesResponse struct {
	MonthlyData  []MonthlyProductSalesDataResponse `json:"monthly_data"`
	TotalRevenue float64                           `json:"total_revenue"`
	TotalQty     float64                           `json:"total_qty"`
	TotalOrders  int                               `json:"total_orders"`
}

// --- Product Detail ---

// GetProductDetailRequest filters product detail
type GetProductDetailRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// ProductDetailStatistics contains a product's statistics for a period
type ProductDetailStatistics struct {
	TotalQty              float64              `json:"total_qty"`
	TotalRevenue          float64              `json:"total_revenue"`
	TotalRevenueFormatted string               `json:"total_revenue_formatted"`
	TotalOrders           int                  `json:"total_orders"`
	AvgPrice              float64              `json:"avg_price"`
	AvgPriceFormatted     string               `json:"avg_price_formatted"`
	PeriodComparison      *ProductPeriodChange `json:"period_comparison,omitempty"`
}

// ProductPeriodChange holds change percentages compared to previous period
type ProductPeriodChange struct {
	RevenueChange float64 `json:"revenue_change"`
	QtyChange     float64 `json:"qty_change"`
	OrdersChange  float64 `json:"orders_change"`
}

// ProductDetailResponse contains full details for a product analysis
type ProductDetailResponse struct {
	ProductID    string                   `json:"product_id"`
	ProductCode  string                   `json:"product_code"`
	ProductName  string                   `json:"product_name"`
	ProductSKU   string                   `json:"product_sku,omitempty"`
	ProductImage string                   `json:"product_image,omitempty"`
	CategoryName string                   `json:"category_name,omitempty"`
	BrandName    string                   `json:"brand_name,omitempty"`
	SellingPrice float64                  `json:"selling_price"`
	CostPrice    float64                  `json:"cost_price"`
	CurrentStock float64                  `json:"current_stock"`
	Statistics   *ProductDetailStatistics `json:"statistics,omitempty"`
}

// --- Product Customers ---

// ListProductCustomersRequest filters customer data for a product
type ListProductCustomersRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// ProductCustomerResponse represents a customer buying a specific product
type ProductCustomerResponse struct {
	CustomerID            string  `json:"customer_id"`
	CustomerName          string  `json:"customer_name"`
	CustomerCode          string  `json:"customer_code,omitempty"`
	CustomerType          string  `json:"customer_type,omitempty"`
	CityName              string  `json:"city_name,omitempty"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
}

// --- Product Sales Reps ---

// ListProductSalesRepsRequest filters sales rep data for a product
type ListProductSalesRepsRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// ProductSalesRepResponse represents a sales rep selling a specific product
type ProductSalesRepResponse struct {
	EmployeeID            string  `json:"employee_id"`
	EmployeeCode          string  `json:"employee_code"`
	Name                  string  `json:"name"`
	AvatarURL             string  `json:"avatar_url,omitempty"`
	PositionName          string  `json:"position_name,omitempty"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
}

// --- Category Performance ---

// ListCategoryPerformanceRequest filters the category performance list
type ListCategoryPerformanceRequest struct {
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// CategoryPerformanceResponse represents a product category's aggregated sales performance
type CategoryPerformanceResponse struct {
	CategoryID            string  `json:"category_id"`
	CategoryName          string  `json:"category_name"`
	ProductCount          int     `json:"product_count"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
	AvgPrice              float64 `json:"avg_price"`
	AvgPriceFormatted     string  `json:"avg_price_formatted"`
}

// --- Segment Performance ---

// ListSegmentPerformanceRequest filters the segment performance list
type ListSegmentPerformanceRequest struct {
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// SegmentPerformanceResponse aggregates sales performance by product segment
type SegmentPerformanceResponse struct {
	SegmentID             string  `json:"segment_id"`
	SegmentName           string  `json:"segment_name"`
	ProductCount          int     `json:"product_count"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
	AvgPrice              float64 `json:"avg_price"`
	AvgPriceFormatted     string  `json:"avg_price_formatted"`
}

// --- Type Performance ---

// ListTypePerformanceRequest filters the product type performance list
type ListTypePerformanceRequest struct {
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// TypePerformanceResponse aggregates sales performance by product type
type TypePerformanceResponse struct {
	TypeID                string  `json:"type_id"`
	TypeName              string  `json:"type_name"`
	ProductCount          int     `json:"product_count"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
	AvgPrice              float64 `json:"avg_price"`
	AvgPriceFormatted     string  `json:"avg_price_formatted"`
}

// --- Packaging Performance ---

// ListPackagingPerformanceRequest filters the packaging performance list
type ListPackagingPerformanceRequest struct {
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// PackagingPerformanceResponse aggregates sales performance by packaging type
type PackagingPerformanceResponse struct {
	PackagingID           string  `json:"packaging_id"`
	PackagingName         string  `json:"packaging_name"`
	ProductCount          int     `json:"product_count"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
	AvgPrice              float64 `json:"avg_price"`
	AvgPriceFormatted     string  `json:"avg_price_formatted"`
}

// --- Procurement Type Performance ---

// ListProcurementTypePerformanceRequest filters the procurement type performance list
type ListProcurementTypePerformanceRequest struct {
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// ProcurementTypePerformanceResponse aggregates sales performance by procurement type
type ProcurementTypePerformanceResponse struct {
	ProcurementTypeID     string  `json:"procurement_type_id"`
	ProcurementTypeName   string  `json:"procurement_type_name"`
	ProductCount          int     `json:"product_count"`
	TotalQty              float64 `json:"total_qty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
	AvgPrice              float64 `json:"avg_price"`
	AvgPriceFormatted     string  `json:"avg_price_formatted"`
}

// --- Product Monthly Trend ---

// GetProductMonthlyTrendRequest filters monthly trend for a specific product
type GetProductMonthlyTrendRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// ProductMonthlyTrendDataResponse represents a single month for a product
type ProductMonthlyTrendDataResponse struct {
	Month        int     `json:"month"`
	MonthName    string  `json:"month_name"`
	Year         int     `json:"year"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalQty     float64 `json:"total_qty"`
	TotalOrders  int     `json:"total_orders"`
}

// ProductMonthlyTrendResponse wraps monthly trend data for a product
type ProductMonthlyTrendResponse struct {
	ProductID   string                            `json:"product_id"`
	ProductName string                            `json:"product_name"`
	MonthlyData []ProductMonthlyTrendDataResponse `json:"monthly_data"`
}
