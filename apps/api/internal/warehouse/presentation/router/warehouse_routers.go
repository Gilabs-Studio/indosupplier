package router

import (
	"github.com/gilabs/gims/api/internal/warehouse/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterWarehouseRoutes registers all warehouse routes
func RegisterWarehouseRoutes(rg *gin.RouterGroup, h *handler.WarehouseHandler) {
	g := rg.Group("/warehouses")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
