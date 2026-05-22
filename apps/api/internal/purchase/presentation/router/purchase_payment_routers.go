package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	purchasePaymentRead       = "purchase_payment.read"
	purchasePaymentCreate     = "purchase_payment.create"
	purchasePaymentUpdate     = "purchase_payment.update"
	purchasePaymentDelete     = "purchase_payment.delete"
	purchasePaymentConfirm    = "purchase_payment.confirm"
	purchasePaymentExport     = "purchase_payment.export"
	purchasePaymentPrint      = "purchase_payment.print"
)

func RegisterPurchasePaymentRoutes(r *gin.RouterGroup, h *handler.PurchasePaymentHandler, printH *handler.PurchasePaymentPrintHandler) {
	g := r.Group("/payments")
	g.GET("/add", middleware.RequirePermission(purchasePaymentCreate), h.Add)
	g.GET("", middleware.RequirePermission(purchasePaymentRead), h.List)
	g.GET("/export", middleware.RequirePermission(purchasePaymentExport), h.Export)
	g.POST("", middleware.RequirePermission(purchasePaymentCreate), h.Create)
	g.POST("/batch", middleware.RequirePermission(purchasePaymentCreate), h.CreateBatch)
	g.PATCH("/batch/confirm", middleware.RequirePermission(purchasePaymentConfirm), h.ConfirmBatch)
	g.PATCH("/batch/post", middleware.RequirePermission(purchasePaymentConfirm), h.ConfirmBatch)
	g.GET("/:id", middleware.RequirePermission(purchasePaymentRead), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission(purchasePaymentRead), h.AuditTrail)
	g.GET("/:id/print", middleware.RequirePermission(purchasePaymentPrint), printH.PrintPurchasePayment)
	g.PUT("/:id", middleware.RequirePermission(purchasePaymentUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(purchasePaymentDelete), h.Delete)
	g.POST("/:id/confirm", middleware.RequirePermission(purchasePaymentConfirm), h.Confirm)
	g.PATCH("/:id/confirm", middleware.RequirePermission(purchasePaymentConfirm), h.Confirm)
	g.PATCH("/:id/post", middleware.RequirePermission(purchasePaymentConfirm), h.Confirm)
}
