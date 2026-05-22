package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAreaRoutes registers area routes
func RegisterAreaRoutes(rg *gin.RouterGroup, h *handler.AreaHandler) {
	g := rg.Group("/areas")
	// Form data endpoint BEFORE parameterized routes for route specificity
	g.GET("/form-data", middleware.RequirePermission("area.read"), h.GetFormData)
	g.GET("", middleware.RequirePermission("area.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("area.read"), h.GetByID)
	g.GET("/:id/detail", middleware.RequirePermission("area.read"), h.GetDetail)
	g.POST("", middleware.RequirePermission("area.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("area.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("area.delete"), h.Delete)
	// Supervisor and member assignment routes
	g.POST("/:id/supervisors", middleware.RequirePermission("area.assign_supervisor"), h.AssignSupervisors)
	g.POST("/:id/members", middleware.RequirePermission("area.assign_member"), h.AssignMembers)
	g.DELETE("/:id/employees/:emp_id", middleware.RequirePermission("area.assign_member"), h.RemoveEmployee)
}
