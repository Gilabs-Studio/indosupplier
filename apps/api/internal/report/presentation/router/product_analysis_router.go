package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/report/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterProductAnalysisRoutes registers all product analysis report routes
func RegisterProductAnalysisRoutes(rg *gin.RouterGroup, h *handler.ProductAnalysisHandler) {
	g := rg.Group("/product-analysis")

	g.GET("/performance", middleware.RequirePermission("report_product_analysis.read"), h.ListPerformance)
	g.GET("/category-performance", middleware.RequirePermission("report_product_analysis.read"), h.ListCategoryPerformance)
	g.GET("/segment-performance", middleware.RequirePermission("report_product_analysis.read"), h.ListSegmentPerformance)
	g.GET("/type-performance", middleware.RequirePermission("report_product_analysis.read"), h.ListTypePerformance)
	g.GET("/packaging-performance", middleware.RequirePermission("report_product_analysis.read"), h.ListPackagingPerformance)
	g.GET("/procurement-type-performance", middleware.RequirePermission("report_product_analysis.read"), h.ListProcurementTypePerformance)
	g.GET("/monthly-overview", middleware.RequirePermission("report_product_analysis.read"), h.GetMonthlyOverview)
	g.GET("/product/:productId", middleware.RequirePermission("report_product_analysis.read"), h.GetProductDetail)
	g.GET("/product/:productId/customers", middleware.RequirePermission("report_product_analysis.read"), h.GetProductCustomers)
	g.GET("/product/:productId/sales-reps", middleware.RequirePermission("report_product_analysis.read"), h.GetProductSalesReps)
	g.GET("/product/:productId/monthly-trend", middleware.RequirePermission("report_product_analysis.read"), h.GetProductMonthlyTrend)
}
