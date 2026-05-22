package dto

// GeoPerformanceRequest filters geo performance report data
type GeoPerformanceRequest struct {
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Mode       string `form:"mode,default=sales_order"` // "sales_order" or "paid_invoice"
	SalesRepID string `form:"sales_rep_id"`
	Level      string `form:"level,default=province"` // "province" or "city"
}

// GeoPerformanceAreaResponse represents a single geographic area's aggregated metrics
type GeoPerformanceAreaResponse struct {
	AreaID        string  `json:"area_id"`
	AreaName      string  `json:"area_name"`
	ParentName    string  `json:"parent_name,omitempty"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalOrders   int     `json:"total_orders"`
	AvgOrderValue float64 `json:"avg_order_value"`
}

// GeoPerformanceSummaryResponse wraps the area list with summary totals
type GeoPerformanceSummaryResponse struct {
	Areas           []GeoPerformanceAreaResponse `json:"areas"`
	TotalRevenue    float64                      `json:"total_revenue"`
	TotalOrders     int                          `json:"total_orders"`
	AvgOrderValue   float64                      `json:"avg_order_value"`
	AreasWithData   int                          `json:"areas_with_data"`
	Mode            string                       `json:"mode"`
	Level           string                       `json:"level"`
}

// GeoPerformanceFormDataResponse provides filter options for the report page
type GeoPerformanceFormDataResponse struct {
	SalesReps []GeoSalesRepOption `json:"sales_reps"`
}

// GeoSalesRepOption represents a sales rep option for filter dropdowns
type GeoSalesRepOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
