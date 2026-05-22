package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// SetupEvaluationCriteriaRoutes sets up routes for evaluation criteria management
func SetupEvaluationCriteriaRoutes(router *gin.RouterGroup, handler *handler.EvaluationCriteriaHandler) {
	criteria := router.Group("/evaluation-criteria")
	{
		// Get criteria by group (static route before /:id for route specificity)
		criteria.GET("/group/:group_id", middleware.RequirePermission("evaluation.read"), handler.GetByGroupID)

		// CRUD routes
		criteria.GET("/:id", middleware.RequirePermission("evaluation.read"), handler.GetByID)
		criteria.POST("", middleware.RequirePermission("evaluation.create"), handler.Create)
		criteria.PUT("/:id", middleware.RequirePermission("evaluation.update"), handler.Update)
		criteria.DELETE("/:id", middleware.RequirePermission("evaluation.delete"), handler.Delete)
	}
}
