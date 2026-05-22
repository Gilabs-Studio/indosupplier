package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/pos/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPublicPOSRoutes mounts customer-facing self-order routes.
// These routes are UNAUTHENTICATED but rate-limited using the "public" tier.
// They are mounted directly on the root gin.Engine (not under the auth group).
//
// Route pattern: /api/v1/public/pos/tables/:token
func RegisterPublicPOSRoutes(r *gin.Engine, h *handler.PublicPOSHandler) {
	rg := r.Group("/api/v1/public/pos/tables/:token")
	rg.Use(middleware.RateLimitMiddleware("public"))

	rg.GET("", h.GetTableInfo)
	rg.POST("/orders", h.CreateCustomerOrder)
	rg.GET("/orders/:orderId", h.GetOrderStatus)
	rg.POST("/orders/:orderId/pay/digital", h.InitiateDigitalPayment)
	rg.POST("/orders/:orderId/pay/cashier", h.MarkPayAtCashier)
	rg.POST("/orders/:orderId/cancel", h.CancelCustomerOrder)
	// Staff can use this state to switch a customer from pending digital to cashier flow.
	rg.POST("/orders/:orderId/payments/:paymentId/cancel", h.CancelDigitalPayment)
}
