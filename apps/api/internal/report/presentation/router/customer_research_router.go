package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/report/presentation/handler"
	"github.com/gin-gonic/gin"
)

const customerResearchReadPermission = "report_customer_research.read"

// RegisterCustomerResearchRoutes registers all customer research report routes.
func RegisterCustomerResearchRoutes(rg *gin.RouterGroup, h *handler.CustomerResearchHandler) {
	g := rg.Group("/customer-research")

	g.GET("/kpis", middleware.RequirePermission(customerResearchReadPermission), h.GetKPIs)
	g.GET("/revenue-by-customer", middleware.RequirePermission(customerResearchReadPermission), h.ListRevenueByCustomer)
	g.GET("/purchase-frequency", middleware.RequirePermission(customerResearchReadPermission), h.ListPurchaseFrequency)
	g.GET("/revenue-trend", middleware.RequirePermission(customerResearchReadPermission), h.GetRevenueTrend)
	g.GET("/customers", middleware.RequirePermission(customerResearchReadPermission), h.ListCustomers)
	g.GET("/customers/:customerId", middleware.RequirePermission(customerResearchReadPermission), h.GetCustomerDetail)
	g.GET("/customers/:customerId/products", middleware.RequirePermission(customerResearchReadPermission), h.GetCustomerTopProducts)
}
