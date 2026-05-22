package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAreaMappingRoutes registers area mapping visualization routes
func RegisterAreaMappingRoutes(r *gin.RouterGroup, h *handler.AreaMappingHandler) {
	g := r.Group("/area-mapping/map")

	// Get area mapping data for map visualization (customer + lead locations with activity metrics)
	g.GET("", middleware.RequirePermission("crm_area_mapping.read"), h.GetAreaMapping)
}
