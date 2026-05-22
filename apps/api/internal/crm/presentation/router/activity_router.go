package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	activityRead   = "crm_task.read"
	activityCreate = "crm_task.create"
)

// RegisterActivityRoutes registers all activity-related routes
func RegisterActivityRoutes(r *gin.RouterGroup, h *handler.ActivityHandler) {
	g := r.Group("/activities")

	// Static routes first (before parameterized routes)
	g.GET("/timeline", middleware.RequirePermission(activityRead), h.Timeline)
	g.GET("/my-activities", h.MyActivities)

	// List and create
	g.GET("", middleware.RequirePermission(activityRead), h.List)
	g.POST("", middleware.RequirePermission(activityCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(activityRead), h.GetByID)
}
