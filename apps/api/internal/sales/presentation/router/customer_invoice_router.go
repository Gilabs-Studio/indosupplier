package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCustomerInvoiceRoutes registers customer invoice routes
func RegisterCustomerInvoiceRoutes(rg *gin.RouterGroup, h *handler.CustomerInvoiceHandler, printH *handler.CustomerInvoicePrintHandler) {
	g := rg.Group("/customer-invoices")
	g.POST("/preview", middleware.RequirePermission("customer_invoice.create"), h.PreviewJournal)
	g.GET("", middleware.RequirePermission("customer_invoice.read"), h.List)
	g.GET("/export", middleware.RequirePermission("customer_invoice.read"), h.Export)
	g.GET("/:id", middleware.RequirePermission("customer_invoice.read"), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission("customer_invoice.read"), h.AuditTrail)
	g.GET("/:id/items", middleware.RequirePermission("customer_invoice.read"), h.ListItems)
	g.GET("/:id/print", middleware.RequirePermission("customer_invoice.print"), printH.PrintInvoice)
	g.POST("", middleware.RequirePermission("customer_invoice.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("customer_invoice.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("customer_invoice.delete"), h.Delete)
	g.PATCH("/:id/status", middleware.RequirePermission("customer_invoice.update"), h.UpdateStatus)
	g.POST("/:id/approve", middleware.RequirePermission("customer_invoice.approve"), h.Approve)
}
