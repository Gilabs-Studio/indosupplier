package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterJobPositionRoutes registers job position routes
func RegisterJobPositionRoutes(rg *gin.RouterGroup, h *handler.JobPositionHandler) {
	g := rg.Group("/job-positions")
	g.GET("", middleware.RequirePermission("job_position.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("job_position.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("job_position.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("job_position.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("job_position.delete"), h.Delete)
}
