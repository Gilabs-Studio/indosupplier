package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterWorkScheduleRoutes registers work schedule routes
func RegisterWorkScheduleRoutes(rg *gin.RouterGroup, h *handler.WorkScheduleHandler) {
	g := rg.Group("/work-schedules")
	g.GET("", middleware.RequirePermission("work_schedule.read"), h.List)
	g.GET("/default", middleware.RequirePermission("work_schedule.read"), h.GetDefault)
	g.GET("/form-data", middleware.RequirePermission("work_schedule.read"), h.GetFormData)
	g.GET("/:id", middleware.RequirePermission("work_schedule.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("work_schedule.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("work_schedule.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("work_schedule.delete"), h.Delete)
	g.POST("/:id/set-default", middleware.RequirePermission("work_schedule.update"), h.SetDefault)
}
