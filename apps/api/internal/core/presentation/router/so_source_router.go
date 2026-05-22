package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterSOSourceRoutes(rg *gin.RouterGroup, h *handler.SOSourceHandler) {
	g := rg.Group("/so-sources")
	{
		g.POST("", middleware.RequirePermission("so_source.create"), h.Create)
		g.GET("", middleware.RequirePermission("so_source.read"), h.List)
		g.GET("/:id", middleware.RequirePermission("so_source.read"), h.GetByID)
		g.PUT("/:id", middleware.RequirePermission("so_source.update"), h.Update)
		g.DELETE("/:id", middleware.RequirePermission("so_source.delete"), h.Delete)
	}
}
