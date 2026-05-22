package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	scheduleRead   = "crm_task.read"
	scheduleCreate = "crm_task.create"
	scheduleUpdate = "crm_task.update"
	scheduleDelete = "crm_task.delete"
)

// RegisterScheduleRoutes registers all schedule-related routes
func RegisterScheduleRoutes(r *gin.RouterGroup, h *handler.ScheduleHandler) {
	g := r.Group("/schedules")

	// Static routes first (before parameterized routes)
	g.GET("/form-data", middleware.RequirePermission(scheduleRead), h.GetFormData)

	// CRUD routes
	g.GET("", middleware.RequirePermission(scheduleRead), h.List)
	g.POST("", middleware.RequirePermission(scheduleCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(scheduleRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(scheduleUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(scheduleDelete), h.Delete)
}
