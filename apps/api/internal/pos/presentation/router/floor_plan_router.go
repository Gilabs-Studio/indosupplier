package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/pos/presentation/handler"
)

// RegisterFloorPlanRoutes sets up floor plan routes
func RegisterFloorPlanRoutes(rg *gin.RouterGroup, h *handler.FloorPlanHandler) {
	plans := rg.Group("/floor-plans")

	// Static routes MUST come before parameterized routes
	plans.GET("/form-data", middleware.RequirePermission("pos.layout.read"), h.GetFormData)

	plans.POST("", middleware.RequirePermission("pos.layout.create"), h.Create)
	plans.GET("", middleware.RequirePermission("pos.layout.read"), h.List)
	plans.GET("/:id", middleware.RequirePermission("pos.layout.read"), h.GetByID)
	plans.PUT("/:id", middleware.RequirePermission("pos.layout.update"), h.Update)
	plans.PUT("/:id/layout", middleware.RequirePermission("pos.layout.update"), h.SaveLayoutData)
	plans.DELETE("/:id", middleware.RequirePermission("pos.layout.delete"), h.Delete)
	plans.POST("/:id/publish", middleware.RequirePermission("pos.layout.update"), h.Publish)
	plans.GET("/:id/versions", middleware.RequirePermission("pos.layout.read"), h.ListVersions)

	// Table QR token management — static /table-tokens before parameterized /:objectId/token
	plans.GET("/:id/table-tokens", middleware.RequirePermission("pos.layout.read"), h.ListTableTokens)
	plans.POST("/:id/tables/:objectId/token", middleware.RequirePermission("pos.layout.update"), h.GenerateTableToken)
	plans.DELETE("/:id/tables/:objectId/token", middleware.RequirePermission("pos.layout.update"), h.RevokeTableToken)
}
