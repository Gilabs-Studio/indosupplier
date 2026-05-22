package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/report/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSupplierResearchRoutes registers all supplier research report routes.
func RegisterSupplierResearchRoutes(rg *gin.RouterGroup, h *handler.SupplierResearchHandler) {
	g := rg.Group("/supplier-research")

	g.GET("/kpis", middleware.RequirePermission("report_supplier_research.read"), h.GetKpis)
	g.GET("/purchase-volume", middleware.RequirePermission("report_supplier_research.read"), h.ListPurchaseVolume)
	g.GET("/delivery-time", middleware.RequirePermission("report_supplier_research.read"), h.ListDeliveryTime)
	g.GET("/spend-trend", middleware.RequirePermission("report_supplier_research.read"), h.GetSpendTrend)
	g.GET("/suppliers", middleware.RequirePermission("report_supplier_research.read"), h.ListSuppliers)
	g.GET("/suppliers/:supplierId", middleware.RequirePermission("report_supplier_research.read"), h.GetSupplierDetail)
}
