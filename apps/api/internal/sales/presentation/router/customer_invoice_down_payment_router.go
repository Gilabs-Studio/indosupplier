package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCustomerInvoiceDownPaymentRoutes registers the routes for customer invoice down payment
func RegisterCustomerInvoiceDownPaymentRoutes(router *gin.RouterGroup, h *handler.CustomerInvoiceDownPaymentHandler, printH *handler.CustomerInvoiceDPPrintHandler) {
	group := router.Group("/customer-invoice-down-payments")
	{
		group.GET("", h.List)
		group.GET("/add", h.Add)
		group.GET("/export", h.Export)
		group.GET("/:id", h.GetByID)
		group.GET("/:id/audit-trail", middleware.RequirePermission("customer_invoice_dp.read"), h.AuditTrail)
		group.GET("/:id/print", middleware.RequirePermission("customer_invoice_dp.print"), printH.PrintDownPaymentInvoice)

		group.POST("", h.Create)
		group.POST("/:id/pending", h.Pending)
		group.POST("/:id/approve", middleware.RequirePermission("customer_invoice_dp.approve"), h.Approve)
		group.POST("/:id/cancel", middleware.RequirePermission("customer_invoice_dp.cancel"), h.Cancel)

		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}
