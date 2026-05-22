package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	areaMappingRead   = "crm_area_mapping.read"
	areaMappingCreate = "crm_area_mapping.create"
)

// RegisterAreaCaptureRoutes registers all area mapping / capture routes
func RegisterAreaCaptureRoutes(r *gin.RouterGroup, h *handler.AreaCaptureHandler) {
	g := r.Group("/area-mapping")

	// Capture GPS location
	g.POST("/capture", middleware.RequirePermission(areaMappingCreate), h.Capture)

	// List captured GPS points
	g.GET("/captures", middleware.RequirePermission(areaMappingRead), h.ListCaptures)

	// Heatmap data for visit density visualization
	g.GET("/heatmap", middleware.RequirePermission(areaMappingRead), h.GetHeatmap)

	// Coverage analysis per area
	g.GET("/coverage", middleware.RequirePermission(areaMappingRead), h.GetCoverage)
}
