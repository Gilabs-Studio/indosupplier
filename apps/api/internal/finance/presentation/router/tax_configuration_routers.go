package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	taxConfigurationReadPermission  = "tax_configuration.read"
	taxConfigurationWritePermission = "tax_configuration.write"
)

func RegisterTaxConfigurationRoutes(group *gin.RouterGroup, h *handler.TaxConfigurationHandler) {
	taxes := group.Group("/taxes")
	taxes.GET("", middleware.RequirePermission(taxConfigurationReadPermission), h.List)
	taxes.POST("", middleware.RequirePermission(taxConfigurationWritePermission), h.Create)
	taxes.GET("/:id", middleware.RequirePermission(taxConfigurationReadPermission), h.GetByID)
	taxes.PUT("/:id", middleware.RequirePermission(taxConfigurationWritePermission), h.Update)
	taxes.PATCH("/:id/toggle-status", middleware.RequirePermission(taxConfigurationWritePermission), h.ToggleStatus)
}
