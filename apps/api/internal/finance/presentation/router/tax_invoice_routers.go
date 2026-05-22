package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	taxInvoiceRead   = "tax_invoice.read"
	taxInvoiceCreate = "tax_invoice.create"
	taxInvoiceUpdate = "tax_invoice.update"
	taxInvoiceDelete = "tax_invoice.delete"
)

func RegisterTaxInvoiceRoutes(r *gin.RouterGroup, h *handler.TaxInvoiceHandler) {
	g := r.Group("/tax-invoices")
	g.GET("", middleware.RequirePermission(taxInvoiceRead), h.List)
	g.GET("/", middleware.RequirePermission(taxInvoiceRead), h.List)
	g.POST("", middleware.RequirePermission(taxInvoiceCreate), h.Create)
	g.POST("/", middleware.RequirePermission(taxInvoiceCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(taxInvoiceRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(taxInvoiceUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(taxInvoiceDelete), h.Delete)
}
