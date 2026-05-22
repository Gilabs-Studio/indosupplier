package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSalesQuotationRoutes registers sales quotation routes
func RegisterSalesQuotationRoutes(rg *gin.RouterGroup, h *handler.SalesQuotationHandler, printH *handler.SalesQuotationPrintHandler) {
	g := rg.Group("/sales-quotations")
	g.GET("", middleware.RequirePermission("sales_quotation.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("sales_quotation.read"), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission("sales_quotation.read"), h.AuditTrail)
	g.GET("/:id/items", middleware.RequirePermission("sales_quotation.read"), h.ListItems)
	g.GET("/:id/print", middleware.RequirePermission("sales_quotation.print"), printH.PrintQuotation)
	g.POST("", middleware.RequirePermission("sales_quotation.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("sales_quotation.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("sales_quotation.delete"), h.Delete)
	g.PATCH("/:id/status", middleware.RequirePermission("sales_quotation.update"), h.UpdateStatus)
}
