package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterOvertimeRequestRoutes registers overtime request routes
func RegisterOvertimeRequestRoutes(rg *gin.RouterGroup, h *handler.OvertimeRequestHandler) {
	g := rg.Group("/overtime")

	// Employee self-service routes
	g.POST("", h.Create)                      // Submit overtime request
	g.GET("/my", h.GetMyRequests)             // Get own overtime requests
	g.GET("/my-summary", h.GetMonthlySummary) // Get own monthly summary
	g.POST("/:id/cancel", h.Cancel)           // Cancel own request

	// Manager routes (for approval workflow)
	g.GET("/pending", middleware.RequirePermission("overtime.approve"), h.GetPending)
	g.POST("/:id/approve", middleware.RequirePermission("overtime.approve"), h.Approve)
	g.POST("/:id/reject", middleware.RequirePermission("overtime.approve"), h.Reject)

	// Admin routes
	g.GET("", middleware.RequirePermission("overtime.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("overtime.read"), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission("overtime.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("overtime.delete"), h.Delete)

	// Notification polling endpoint
	g.GET("/notifications", middleware.RequirePermission("overtime.approve"), h.GetPendingNotifications)
}
