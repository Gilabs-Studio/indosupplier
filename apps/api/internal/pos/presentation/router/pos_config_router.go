package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/pos/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPOSConfigRoutes registers POS configuration routes
func RegisterPOSConfigRoutes(rg *gin.RouterGroup, h *handler.POSConfigHandler) {
	configs := rg.Group("/config/outlet/:outletID")

	// Terminal-related config now follows POS Terminal read permission.
	configs.GET("", middleware.RequirePermission("pos.order.read"), h.GetByOutlet)
	configs.PUT("", middleware.RequirePermission("pos.order.read"), h.Upsert)
	configs.PUT("/receipt-whatsapp-template", middleware.RequirePermission("pos.order.read"), h.UpdateReceiptWhatsAppTemplate)
}
