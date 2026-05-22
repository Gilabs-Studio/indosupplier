package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/geographic/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCountryRoutes registers country routes
func RegisterCountryRoutes(rg *gin.RouterGroup, h *handler.CountryHandler) {
	g := rg.Group("/countries")
	g.GET("", middleware.RequirePermission("geographic.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("geographic.read"), h.GetByID)
}
