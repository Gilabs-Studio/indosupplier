package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSalesOrderRoutes registers sales order routes
func RegisterSalesOrderRoutes(rg *gin.RouterGroup, h *handler.SalesOrderHandler, printH *handler.SalesOrderPrintHandler) {
	g := rg.Group("/sales-orders")
	g.GET("", middleware.RequirePermission("sales_order.read"), h.List)
	g.GET("/export", middleware.RequirePermission("sales_order.read"), h.Export)
	g.GET("/:id", middleware.RequirePermission("sales_order.read"), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission("sales_order.read"), h.AuditTrail)
	g.GET("/:id/items", middleware.RequirePermission("sales_order.read"), h.ListItems)
	g.GET("/:id/print", middleware.RequirePermission("sales_order.print"), printH.PrintOrder)
	g.POST("", middleware.RequirePermission("sales_order.create"), h.Create)
	g.POST("/convert-from-quotation", middleware.RequirePermission("sales_order.create"), h.ConvertFromQuotation)
	g.PUT("/:id", middleware.RequirePermission("sales_order.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("sales_order.delete"), h.Delete)
	g.PATCH("/:id/status", middleware.RequirePermission("sales_order.update"), h.UpdateStatus)
	g.POST("/:id/approve", middleware.RequirePermission("sales_order.approve"), h.Approve)
}
