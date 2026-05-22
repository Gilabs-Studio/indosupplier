package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/geographic/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterVillageRoutes registers village routes
func RegisterVillageRoutes(rg *gin.RouterGroup, h *handler.VillageHandler) {
	g := rg.Group("/villages")
	g.GET("", middleware.RequirePermission("geographic.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("geographic.read"), h.GetByID)
}
