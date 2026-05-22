package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/pos/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPOSPaymentRoutes registers POS payment routes.
// All payment actions (read + process) require pos.order.create (cashier-level).
// Digital payment routes check at the usecase level that the merchant has a connected Xendit account.
func RegisterPOSPaymentRoutes(rg *gin.RouterGroup, h *handler.POSPaymentHandler) {
	// Use :id to match the existing /orders/:id routes and avoid Gin wildcard conflicts.
	payments := rg.Group("/orders/:id/payments")
	payments.Use(middleware.RequirePermission("pos.order.create"))

	payments.GET("", h.GetByOrder)
	payments.POST("/cash", middleware.IdempotentRequest(), h.ProcessCash)
	payments.POST("/digital", h.InitiateDigitalPayment)
	payments.POST("/:paymentID/cancel", h.CancelPendingPayment)

	// Xendit webhook — no auth required (server-to-server callback from Xendit)
	// Registered directly on the parent group to avoid the permission middleware
	rg.POST("/webhook/xendit", h.XenditWebhook)
}
