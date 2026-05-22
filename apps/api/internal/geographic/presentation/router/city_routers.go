package router

import (
	"github.com/gilabs/gims/api/internal/geographic/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCityRoutes registers city routes
func RegisterCityRoutes(rg *gin.RouterGroup, h *handler.CityHandler) {
	g := rg.Group("/cities")
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
}
