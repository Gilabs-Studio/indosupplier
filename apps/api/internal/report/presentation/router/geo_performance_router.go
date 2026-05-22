package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/report/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterGeoPerformanceRoutes registers all geo performance report routes
func RegisterGeoPerformanceRoutes(rg *gin.RouterGroup, h *handler.GeoPerformanceHandler) {
	g := rg.Group("/geo-performance")

	g.GET("/form-data", middleware.RequirePermission("report_geo_performance.read"), h.GetFormData)
	g.GET("", middleware.RequirePermission("report_geo_performance.read"), h.GetGeoPerformance)
}
