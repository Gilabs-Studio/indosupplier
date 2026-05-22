package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/geographic/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterDistrictRoutes registers district routes
func RegisterDistrictRoutes(rg *gin.RouterGroup, h *handler.DistrictHandler) {
	g := rg.Group("/districts")
	g.GET("", middleware.RequirePermission("geographic.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("geographic.read"), h.GetByID)
}
