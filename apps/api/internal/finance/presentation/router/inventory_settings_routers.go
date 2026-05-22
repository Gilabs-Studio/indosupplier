package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	inventorySettingsReadPermission  = "inventory_settings.read"
	inventorySettingsWritePermission = "inventory_settings.write"
)

func RegisterInventorySettingsRoutes(group *gin.RouterGroup, h *handler.InventorySettingsHandler) {
	settings := group.Group("/inventory-settings")
	settings.GET("", middleware.RequirePermission(inventorySettingsReadPermission), h.Get)
	settings.PUT("", middleware.RequirePermission(inventorySettingsWritePermission), h.Upsert)
	settings.GET("/avg-cost/:product_id", middleware.RequirePermission(inventorySettingsReadPermission), h.GetAverageCost)
}
