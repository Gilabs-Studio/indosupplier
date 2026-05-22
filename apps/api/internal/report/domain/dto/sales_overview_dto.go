package dto

// ListSalesRepPerformanceRequest filters sales rep performance list
type ListSalesRepPerformanceRequest struct {
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// SalesRepPerformanceResponse represents a sales rep's aggregated performance
type SalesRepPerformanceResponse struct {
	EmployeeID                  string   `json:"employee_id"`
	EmployeeCode                string   `json:"employee_code"`
	Name                        string   `json:"name"`
	Email                       string   `json:"email,omitempty"`
	AvatarURL                   string   `json:"avatar_url,omitempty"`
	PositionName                string   `json:"position_name,omitempty"`
	DivisionName                string   `json:"division_name,omitempty"`
	TotalRevenue                float64  `json:"total_revenue"`
	TotalRevenueFormatted       string   `json:"total_revenue_formatted"`
	TotalOrders                 int      `json:"total_orders"`
	TotalDeliveries             int      `json:"total_deliveries"`
	TotalInvoices               int      `json:"total_invoices"`
	VisitsCompleted             int      `json:"visits_completed"`
	TasksCompleted              int      `json:"tasks_completed"`
	ConversionRate              float64  `json:"conversion_rate"`
	AverageOrderValue           float64  `json:"average_order_value"`
	AverageOrderValueFormatted  string   `json:"average_order_value_formatted"`
	TargetAmount                *float64 `json:"target_amount,omitempty"`
	TargetAmountFormatted       *string  `json:"target_amount_formatted,omitempty"`
	TargetAchievementPercentage *float64 `json:"target_achievement_percentage,omitempty"`
}

// MonthlySalesOverviewRequest filters monthly overview data
type MonthlySalesOverviewRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// MonthlySalesDataResponse represents a single month's aggregated sales data
type MonthlySalesDataResponse struct {
	Month           int     `json:"month"`
	MonthName       string  `json:"month_name"`
	Year            int     `json:"year"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalCashIn     float64 `json:"total_cash_in"`
	TotalOrders     int     `json:"total_orders"`
	TotalVisits     int     `json:"total_visits"`
	TotalDeliveries int     `json:"total_deliveries"`
	TargetAmount    float64 `json:"target_amount"`
}

// MonthlySalesOverviewResponse is the aggregated monthly overview
type MonthlySalesOverviewResponse struct {
	MonthlyData     []MonthlySalesDataResponse `json:"monthly_data"`
	TotalRevenue    float64                    `json:"total_revenue"`
	TotalCashIn     float64                    `json:"total_cash_in"`
	TotalOrders     int                        `json:"total_orders"`
	TotalVisits     int                        `json:"total_visits"`
	TotalDeliveries int                        `json:"total_deliveries"`
}

// GetSalesRepDetailRequest filters sales rep detail
type GetSalesRepDetailRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// SalesRepStatisticsResponse contains a sales rep's statistics for a period
type SalesRepStatisticsResponse struct {
	TotalRevenue               float64               `json:"total_revenue"`
	TotalRevenueFormatted      string                `json:"total_revenue_formatted"`
	TotalOrders                int                   `json:"total_orders"`
	VisitsCompleted            int                   `json:"visits_completed"`
	TasksCompleted             int                   `json:"tasks_completed"`
	ConversionRate             float64               `json:"conversion_rate"`
	AverageOrderValue          float64               `json:"average_order_value"`
	AverageOrderValueFormatted string                `json:"average_order_value_formatted"`
	TargetAmount               *float64              `json:"target_amount,omitempty"`
	TargetAmountFormatted      *string               `json:"target_amount_formatted,omitempty"`
	TargetAchievementPercent   *float64              `json:"target_achievement_percent,omitempty"`
	PeriodComparison           *PeriodComparisonData `json:"period_comparison,omitempty"`
}

// PeriodComparisonData holds the change percentages against the previous period
type PeriodComparisonData struct {
	RevenueChange float64 `json:"revenue_change"`
	OrdersChange  float64 `json:"orders_change"`
	VisitsChange  float64 `json:"visits_change"`
}

// SalesRepDetailResponse contains full details for a sales rep
type SalesRepDetailResponse struct {
	EmployeeID   string                      `json:"employee_id"`
	EmployeeCode string                      `json:"employee_code"`
	Name         string                      `json:"name"`
	Email        string                      `json:"email,omitempty"`
	AvatarURL    string                      `json:"avatar_url,omitempty"`
	PositionName string                      `json:"position_name,omitempty"`
	DivisionName string                      `json:"division_name,omitempty"`
	Statistics   *SalesRepStatisticsResponse `json:"statistics,omitempty"`
}

// GetSalesRepCheckInLocationsRequest filters check-in location data
type GetSalesRepCheckInLocationsRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
}

// LocationData represents geographic coordinates
type LocationData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

// CustomerRefData is a brief customer reference
type CustomerRefData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CheckInLocationResponse represents a single check-in record
type CheckInLocationResponse struct {
	VisitNumber   int              `json:"visit_number"`
	VisitReportID string           `json:"visit_report_id"`
	VisitDate     string           `json:"visit_date"`
	CheckInTime   string           `json:"check_in_time"`
	Location      *LocationData    `json:"location,omitempty"`
	Customer      *CustomerRefData `json:"customer,omitempty"`
	Purpose       string           `json:"purpose"`
}

// SalesRepRefData is a brief sales rep reference
type SalesRepRefData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// PeriodData represents a date range
type PeriodData struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// SalesRepCheckInLocationsResponse wraps check-in data with meta info
type SalesRepCheckInLocationsResponse struct {
	SalesRep         *SalesRepRefData          `json:"sales_rep,omitempty"`
	CheckInLocations []CheckInLocationResponse `json:"check_in_locations"`
	TotalVisits      int                       `json:"total_visits"`
	Period           *PeriodData               `json:"period,omitempty"`
}

// ListSalesRepProductsRequest filters product data for a sales rep
type ListSalesRepProductsRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// SalesRepProductResponse represents a product sold by a sales rep
type SalesRepProductResponse struct {
	ProductID             string  `json:"product_id"`
	ProductName           string  `json:"product_name"`
	ProductSKU            string  `json:"product_sku,omitempty"`
	ProductImage          string  `json:"product_image,omitempty"`
	CategoryName          string  `json:"category_name,omitempty"`
	TotalQuantity         float64 `json:"total_quantity"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	AveragePrice          float64 `json:"average_price"`
	AveragePriceFormatted string  `json:"average_price_formatted"`
	LastSoldDate          string  `json:"last_sold_date,omitempty"`
}

// ListSalesRepCustomersRequest filters customer data for a sales rep
type ListSalesRepCustomersRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
	SortBy    string `form:"sort_by,default=revenue"`
	Order     string `form:"order,default=desc"`
}

// SalesRepCustomerResponse represents a customer associated with a sales rep
type SalesRepCustomerResponse struct {
	CustomerID            string  `json:"customer_id"`
	CustomerName          string  `json:"customer_name"`
	CustomerCode          string  `json:"customer_code,omitempty"`
	CustomerType          string  `json:"customer_type,omitempty"`
	City                  string  `json:"city,omitempty"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalRevenueFormatted string  `json:"total_revenue_formatted"`
	TotalOrders           int     `json:"total_orders"`
	Status                string  `json:"status,omitempty"`
}

// EmployeeDashboardMetricsRequest filters dashboard metrics for current employee
type EmployeeDashboardMetricsRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// CheckInLocationsSummary represents aggregated check-in locations data
type CheckInLocationsSummary struct {
	TotalLocations int    `json:"total_locations"`
	TotalVisits    int    `json:"total_visits"`
	Period         *PeriodData `json:"period,omitempty"`
}

// ProductsSoldSummary represents aggregated products sold data
type ProductsSoldSummary struct {
	TotalProducts      int     `json:"total_products"`
	TotalQuantity      float64 `json:"total_quantity"`
	TotalRevenue       float64 `json:"total_revenue"`
	TotalRevenueFormatted string `json:"total_revenue_formatted"`
	AverageRevenue     float64 `json:"average_revenue"`
	AverageRevenueFormatted string `json:"average_revenue_formatted"`
}

// CustomersSummary represents aggregated customers data
type CustomersSummary struct {
	TotalCustomers     int     `json:"total_customers"`
	TotalRevenue       float64 `json:"total_revenue"`
	TotalRevenueFormatted string `json:"total_revenue_formatted"`
	AverageOrderValue  float64 `json:"average_order_value"`
	AverageOrderValueFormatted string `json:"average_order_value_formatted"`
	TotalOrders        int     `json:"total_orders"`
}

// EmployeeDashboardMetricsResponse contains aggregated metrics for employee dashboard/profile
type EmployeeDashboardMetricsResponse struct {
	CheckInLocations *CheckInLocationsSummary `json:"check_in_locations,omitempty"`
	ProductsSold     *ProductsSoldSummary     `json:"products_sold,omitempty"`
	Customers        *CustomersSummary        `json:"customers,omitempty"`
}
