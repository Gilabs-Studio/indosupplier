package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// SetupEvaluationGroupRoutes sets up routes for evaluation group management
func SetupEvaluationGroupRoutes(router *gin.RouterGroup, handler *handler.EvaluationGroupHandler) {
	groups := router.Group("/evaluation-groups")
	{
		// CRUD routes
		groups.GET("", middleware.RequirePermission("evaluation.read"), handler.GetAll)
		groups.GET("/:id/audit-trail", middleware.RequirePermission("evaluation.audit_trail"), handler.AuditTrail)
		groups.GET("/:id", middleware.RequirePermission("evaluation.read"), handler.GetByID)
		groups.POST("", middleware.RequirePermission("evaluation.create"), handler.Create)
		groups.PUT("/:id", middleware.RequirePermission("evaluation.update"), handler.Update)
		groups.DELETE("/:id", middleware.RequirePermission("evaluation.delete"), handler.Delete)
	}
}
