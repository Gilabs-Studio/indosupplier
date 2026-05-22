package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	salesPaymentRead    = "sales_payment.read"
	salesPaymentCreate  = "sales_payment.create"
	salesPaymentDelete  = "sales_payment.delete"
	salesPaymentConfirm = "sales_payment.confirm"
	salesPaymentReverse = "sales_payment.reverse"
	salesPaymentExport  = "sales_payment.export"
)

func RegisterSalesPaymentRoutes(r *gin.RouterGroup, h *handler.SalesPaymentHandler, printH *handler.SalesPaymentPrintHandler) {
	g := r.Group("/payments")
	g.GET("/add", middleware.RequirePermission(salesPaymentCreate), h.Add)
	g.GET("", middleware.RequirePermission(salesPaymentRead), h.List)
	g.GET("/export", middleware.RequirePermission(salesPaymentExport), h.Export)
	g.POST("", middleware.RequirePermission(salesPaymentCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(salesPaymentRead), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission(salesPaymentRead), h.AuditTrail)
	g.GET("/:id/print", middleware.RequirePermission("sales_payment.print"), printH.PrintPayment)
	g.DELETE("/:id", middleware.RequirePermission(salesPaymentDelete), h.Delete)
	g.POST("/:id/confirm", middleware.RequirePermission(salesPaymentConfirm), h.Confirm)
	g.POST("/:id/reverse", middleware.RequirePermission(salesPaymentReverse), h.Reverse)
}
