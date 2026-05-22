package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	taskRead   = "crm_task.read"
	taskCreate = "crm_task.create"
	taskUpdate = "crm_task.update"
	taskDelete = "crm_task.delete"
	taskAssign = "crm_task.assign"
)

// RegisterTaskRoutes registers all task-related routes
func RegisterTaskRoutes(r *gin.RouterGroup, h *handler.TaskHandler) {
	g := r.Group("/tasks")

	// Static routes first (before parameterized routes)
	g.GET("/form-data", middleware.RequirePermission(taskRead), h.GetFormData)

	// CRUD routes
	g.GET("", middleware.RequirePermission(taskRead), h.List)
	g.POST("", middleware.RequirePermission(taskCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(taskRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(taskUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(taskDelete), h.Delete)

	// Workflow actions
	g.POST("/:id/assign", middleware.RequirePermission(taskAssign), h.Assign)
	g.POST("/:id/complete", middleware.RequirePermission(taskUpdate), h.Complete)
	g.POST("/:id/in-progress", middleware.RequirePermission(taskUpdate), h.MarkInProgress)
	g.POST("/:id/cancel", middleware.RequirePermission(taskUpdate), h.Cancel)

	// Nested reminder routes
	g.GET("/:id/reminders", middleware.RequirePermission(taskRead), h.ListReminders)
	g.POST("/:id/reminders", middleware.RequirePermission(taskCreate), h.CreateReminder)
	g.GET("/:id/reminders/:reminderID", middleware.RequirePermission(taskRead), h.GetReminderByID)
	g.PUT("/:id/reminders/:reminderID", middleware.RequirePermission(taskUpdate), h.UpdateReminder)
	g.DELETE("/:id/reminders/:reminderID", middleware.RequirePermission(taskDelete), h.DeleteReminder)
}
