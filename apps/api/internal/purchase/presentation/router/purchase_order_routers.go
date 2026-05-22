package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	purchaseOrderRead    = "purchase_order.read"
	purchaseOrderCreate  = "purchase_order.create"
	purchaseOrderUpdate  = "purchase_order.update"
	purchaseOrderDelete  = "purchase_order.delete"
	purchaseOrderExport  = "purchase_order.export"
	purchaseOrderPrint   = "purchase_order.print"
	purchaseOrderSubmit  = "purchase_order.submit"
	purchaseOrderApprove = "purchase_order.approve"
	purchaseOrderReject  = "purchase_order.reject"
)

func RegisterPurchaseOrderRoutes(r *gin.RouterGroup, h *handler.PurchaseOrderHandler, printH *handler.PurchaseOrderPrintHandler) {
	g := r.Group("/purchase-orders")
	g.GET("/add", middleware.RequirePermission(purchaseOrderCreate), h.Add)
	g.GET("", middleware.RequirePermission(purchaseOrderRead), h.List)
	g.GET("/export", middleware.RequirePermission(purchaseOrderExport), h.Export)
	g.POST("", middleware.RequirePermission(purchaseOrderCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(purchaseOrderRead), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission(purchaseOrderRead), h.AuditTrail)
	g.GET("/:id/print", middleware.RequirePermission(purchaseOrderPrint), printH.PrintPurchaseOrder)
	g.PUT("/:id", middleware.RequirePermission(purchaseOrderUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(purchaseOrderDelete), h.Delete)
	// New workflow actions
	g.POST("/:id/submit", middleware.RequirePermission(purchaseOrderSubmit), h.Submit)
	g.POST("/:id/approve", middleware.RequirePermission(purchaseOrderApprove), h.Approve)
	g.POST("/:id/reject", middleware.RequirePermission(purchaseOrderReject), h.Reject)
	g.PATCH("/:id/submit", middleware.RequirePermission(purchaseOrderSubmit), h.Submit)
	g.PATCH("/:id/approve", middleware.RequirePermission(purchaseOrderApprove), h.Approve)
	g.PATCH("/:id/reject", middleware.RequirePermission(purchaseOrderReject), h.Reject)
}
