package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterDivisionRoutes registers division routes
func RegisterDivisionRoutes(rg *gin.RouterGroup, h *handler.DivisionHandler) {
	g := rg.Group("/divisions")
	g.GET("", middleware.RequirePermission("division.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("division.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("division.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("division.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("division.delete"), h.Delete)
}
