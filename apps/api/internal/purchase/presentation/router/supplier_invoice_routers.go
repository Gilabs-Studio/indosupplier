package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	supplierInvoiceRead    = "supplier_invoice.read"
	supplierInvoiceCreate  = "supplier_invoice.create"
	supplierInvoiceUpdate  = "supplier_invoice.update"
	supplierInvoiceDelete  = "supplier_invoice.delete"
	supplierInvoiceSubmit  = "supplier_invoice.submit"
	supplierInvoiceApprove = "supplier_invoice.approve"
	supplierInvoiceReject  = "supplier_invoice.reject"
	supplierInvoiceCancel  = "supplier_invoice.cancel"
	supplierInvoicePending = "supplier_invoice.pending"
	supplierInvoiceExport  = "supplier_invoice.export"
	supplierInvoicePrint   = "supplier_invoice.print"
)

func RegisterSupplierInvoiceRoutes(r *gin.RouterGroup, h *handler.SupplierInvoiceHandler, printH *handler.SupplierInvoicePrintHandler) {
	g := r.Group("/supplier-invoices")
	g.GET("/add", middleware.RequirePermission(supplierInvoiceCreate), h.Add)
	g.GET("", middleware.RequirePermission(supplierInvoiceRead), h.List)
	g.GET("/export", middleware.RequirePermission(supplierInvoiceExport), h.Export)
	g.POST("/preview", middleware.RequirePermission(supplierInvoiceCreate), h.PreviewJournal)
	g.POST("", middleware.RequirePermission(supplierInvoiceCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(supplierInvoiceRead), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission(supplierInvoiceRead), h.AuditTrail)
	g.GET("/:id/print", middleware.RequirePermission(supplierInvoicePrint), printH.PrintSupplierInvoice)
	g.PUT("/:id", middleware.RequirePermission(supplierInvoiceUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(supplierInvoiceDelete), h.Delete)
	g.POST("/:id/submit", middleware.RequirePermission(supplierInvoiceSubmit), h.Submit)
	g.POST("/:id/approve", middleware.RequirePermission(supplierInvoiceApprove), h.Approve)
	g.POST("/:id/reject", middleware.RequirePermission(supplierInvoiceReject), h.Reject)
	g.POST("/:id/cancel", middleware.RequirePermission(supplierInvoiceCancel), h.Cancel)
	g.POST("/:id/pending", middleware.RequirePermission(supplierInvoicePending), h.Pending)
	g.POST("/:id/reverse", middleware.RequirePermission(supplierInvoiceCancel), h.Reverse)
	g.PATCH("/:id/submit", middleware.RequirePermission(supplierInvoiceSubmit), h.Submit)
	g.PATCH("/:id/approve", middleware.RequirePermission(supplierInvoiceApprove), h.Approve)
	g.PATCH("/:id/reject", middleware.RequirePermission(supplierInvoiceReject), h.Reject)
	g.PATCH("/:id/cancel", middleware.RequirePermission(supplierInvoiceCancel), h.Cancel)
	g.PATCH("/:id/pending", middleware.RequirePermission(supplierInvoicePending), h.Pending)
	g.PATCH("/:id/post", middleware.RequirePermission(supplierInvoicePending), h.Pending)
	g.PATCH("/:id/reverse", middleware.RequirePermission(supplierInvoiceCancel), h.Reverse)
}
