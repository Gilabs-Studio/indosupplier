package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterLeaveTypeRoutes(rg *gin.RouterGroup, h *handler.LeaveTypeHandler) {
	g := rg.Group("/leave-types")
	{
		g.POST("", middleware.RequirePermission("leave_type.create"), h.Create)
		g.GET("", middleware.RequirePermission("leave_type.read"), h.List)
		g.GET("/:id", middleware.RequirePermission("leave_type.read"), h.GetByID)
		g.PUT("/:id", middleware.RequirePermission("leave_type.update"), h.Update)
		g.DELETE("/:id", middleware.RequirePermission("leave_type.delete"), h.Delete)
	}
}
