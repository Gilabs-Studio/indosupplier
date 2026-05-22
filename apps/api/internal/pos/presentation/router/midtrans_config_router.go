package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/pos/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterXenditConfigRoutes registers Xendit gateway configuration routes.
// Management endpoints (connect/disconnect/update) are restricted to pos.payment.manage.
// The status endpoint is accessible to any authenticated POS user (pos.order.create).
func RegisterXenditConfigRoutes(rg *gin.RouterGroup, h *handler.XenditConfigHandler) {
	// Lightweight connection status — cashiers need this to know if digital payment is available
	rg.GET("/payment/status", middleware.RequirePermission("pos.order.create"), h.GetStatus)

	// Management routes — Owner/Admin only
	payment := rg.Group("/payment/config")
	payment.Use(middleware.RequirePermission("pos.payment.manage"))

	payment.GET("", h.Get)
	payment.POST("/connect", h.Connect)
	payment.PATCH("", h.Update)
	payment.DELETE("/disconnect", h.Disconnect)
}
