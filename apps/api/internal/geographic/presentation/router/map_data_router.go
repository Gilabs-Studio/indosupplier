package router

import (
	"github.com/gilabs/gims/api/internal/geographic/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterMapDataRoutes registers map data routes
func RegisterMapDataRoutes(rg *gin.RouterGroup, h *handler.MapDataHandler) {
	// Map data endpoint is placed before parameterized routes for route specificity
	rg.GET("/map-data", h.GetMapData)
	rg.GET("/reverse-geocode", h.ReverseGeocode)
}
