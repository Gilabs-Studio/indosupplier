package dto

type TotalSalesMetric struct {
	CurrentAmount    float64 `json:"current_amount"`
	FormattedAmount  string  `json:"formatted_amount"`
	PreviousAmount   float64 `json:"previous_amount"`
	GrowthPercentage string  `json:"growth_percentage"`
	PaidOrderCount   int64   `json:"paid_order_count"`
}

type ActiveProductsMetric struct {
	TotalCount    int64 `json:"total_count"`
	ActiveCount   int64 `json:"active_count"`
	DraftCount    int64 `json:"draft_count"`
	FeaturedCount int64 `json:"featured_count"`
}

type MatchingRFQsMetric struct {
	TotalOpen         int64 `json:"total_open"`
	ExpiringSoonCount int64 `json:"expiring_soon_count"`
	NewThisWeekCount  int64 `json:"new_this_week_count"`
}

type MonthlySalesPerformance struct {
	Month           string  `json:"month"`
	MonthName       string  `json:"month_name"`
	MonthNumber     int     `json:"month_number"`
	Year            int     `json:"year"`
	Amount          float64 `json:"amount"`
	FormattedAmount string  `json:"formatted_amount"`
	OrderCount      int64   `json:"order_count"`
}

type SellerPerformanceMetric struct {
	VerificationLevel int     `json:"verification_level"`
	VerificationBadge string  `json:"verification_badge"`
	TrustDescription  string  `json:"trust_description"`
	StarRating        float64 `json:"star_rating"`
	ReviewCount       int     `json:"review_count"`
	ChatResponseRate  string  `json:"chat_response_rate"`
	ResponseTimeText  string  `json:"response_time_text"`
	MonthlyVisitors   int64   `json:"monthly_visitors"`
}

type RecentRFQItem struct {
	ID               string `json:"id"`
	RFQNumber        string `json:"rfq_number"`
	Product          string `json:"product"`
	Category         string `json:"category"`
	Quantity         string `json:"quantity"`
	TargetPort       string `json:"target_port"`
	Date             string `json:"date"`
	Replies          int64  `json:"replies"`
	Status           string `json:"status"`
	MatchingCategory bool   `json:"matching_category"`
}

type SupplierDashboardResponse struct {
	TotalSales        TotalSalesMetric          `json:"total_sales"`
	ActiveProducts    ActiveProductsMetric      `json:"active_products"`
	MatchingRFQs      MatchingRFQsMetric        `json:"matching_rfqs"`
	SalesPerformance  []MonthlySalesPerformance `json:"sales_performance"`
	SellerPerformance SellerPerformanceMetric   `json:"seller_performance"`
	RecentRFQs        []RecentRFQItem           `json:"recent_rfqs"`
}
