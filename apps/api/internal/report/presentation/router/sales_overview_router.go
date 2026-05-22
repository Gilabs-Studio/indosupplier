package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/report/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSalesOverviewRoutes registers all sales overview report routes
func RegisterSalesOverviewRoutes(rg *gin.RouterGroup, h *handler.SalesOverviewHandler) {
	g := rg.Group("/sales-overview")

	g.GET("/performance", middleware.RequirePermission("report_sales_overview.read"), h.ListPerformance)
	g.GET("/monthly-overview", middleware.RequirePermission("report_sales_overview.read"), h.GetMonthlyOverview)
	// Profile metrics endpoint - for current logged-in user's dashboard/profile
	g.GET("/profile-metrics", h.GetEmployeeDashboardMetrics)
	
	g.GET("/sales-rep/:employeeId", middleware.RequirePermission("report_sales_overview.read"), h.GetSalesRepDetail)
	g.GET("/sales-rep/:employeeId/check-in-locations", middleware.RequirePermission("report_sales_overview.read"), h.GetCheckInLocations)
	g.GET("/sales-rep/:employeeId/products", middleware.RequirePermission("report_sales_overview.read"), h.GetProducts)
	g.GET("/sales-rep/:employeeId/customers", middleware.RequirePermission("report_sales_overview.read"), h.GetCustomers)
}
