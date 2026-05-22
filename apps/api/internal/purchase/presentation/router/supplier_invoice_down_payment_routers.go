package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	supplierInvoiceDPRead    = "supplier_invoice_dp.read"
	supplierInvoiceDPCreate  = "supplier_invoice_dp.create"
	supplierInvoiceDPUpdate  = "supplier_invoice_dp.update"
	supplierInvoiceDPDelete  = "supplier_invoice_dp.delete"
	supplierInvoiceDPPending = "supplier_invoice_dp.pending"
	supplierInvoiceDPSubmit  = "supplier_invoice_dp.submit"
	supplierInvoiceDPApprove = "supplier_invoice_dp.approve"
	supplierInvoiceDPReject  = "supplier_invoice_dp.reject"
	supplierInvoiceDPCancel  = "supplier_invoice_dp.cancel"
	supplierInvoiceDPExport  = "supplier_invoice_dp.export"
	supplierInvoiceDPPrint   = "supplier_invoice_dp.print"
)

func RegisterSupplierInvoiceDownPaymentRoutes(r *gin.RouterGroup, h *handler.SupplierInvoiceDownPaymentHandler, printH *handler.SupplierInvoiceDPPrintHandler) {
	g := r.Group("/supplier-invoice-down-payments")
	g.GET("/add", middleware.RequirePermission(supplierInvoiceDPCreate), h.Add)
	g.GET("", middleware.RequirePermission(supplierInvoiceDPRead), h.List)
	g.GET("/export", middleware.RequirePermission(supplierInvoiceDPExport), h.Export)
	g.POST("", middleware.RequirePermission(supplierInvoiceDPCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(supplierInvoiceDPRead), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission(supplierInvoiceDPRead), h.AuditTrail)
	g.GET("/:id/print", middleware.RequirePermission(supplierInvoiceDPPrint), printH.PrintSupplierInvoiceDP)
	g.PUT("/:id", middleware.RequirePermission(supplierInvoiceDPUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(supplierInvoiceDPDelete), h.Delete)
	g.POST("/:id/pending", middleware.RequirePermission(supplierInvoiceDPPending), h.Pending)
	g.POST("/:id/submit", middleware.RequirePermission(supplierInvoiceDPSubmit), h.Submit)
	g.POST("/:id/approve", middleware.RequirePermission(supplierInvoiceDPApprove), h.Approve)
	g.POST("/:id/reject", middleware.RequirePermission(supplierInvoiceDPReject), h.Reject)
	g.POST("/:id/cancel", middleware.RequirePermission(supplierInvoiceDPCancel), h.Cancel)
	g.PATCH("/:id/pending", middleware.RequirePermission(supplierInvoiceDPPending), h.Pending)
	g.PATCH("/:id/post", middleware.RequirePermission(supplierInvoiceDPPending), h.Pending)
	g.PATCH("/:id/submit", middleware.RequirePermission(supplierInvoiceDPSubmit), h.Submit)
	g.PATCH("/:id/approve", middleware.RequirePermission(supplierInvoiceDPApprove), h.Approve)
	g.PATCH("/:id/reject", middleware.RequirePermission(supplierInvoiceDPReject), h.Reject)
	g.PATCH("/:id/cancel", middleware.RequirePermission(supplierInvoiceDPCancel), h.Cancel)
}
