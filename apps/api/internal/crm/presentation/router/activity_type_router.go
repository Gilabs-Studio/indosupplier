package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	activityTypeRead   = "crm_activity_type.read"
	activityTypeCreate = "crm_activity_type.create"
	activityTypeUpdate = "crm_activity_type.update"
	activityTypeDelete = "crm_activity_type.delete"
)

// RegisterActivityTypeRoutes registers activity type routes
func RegisterActivityTypeRoutes(r *gin.RouterGroup, h *handler.ActivityTypeHandler) {
	g := r.Group("/activity-types")
	g.GET("", middleware.RequirePermission(activityTypeRead), h.List)
	g.POST("", middleware.RequirePermission(activityTypeCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(activityTypeRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(activityTypeUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(activityTypeDelete), h.Delete)
}
