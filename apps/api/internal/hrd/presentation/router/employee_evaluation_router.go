package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// SetupEmployeeEvaluationRoutes sets up routes for employee evaluation management
func SetupEmployeeEvaluationRoutes(router *gin.RouterGroup, handler *handler.EmployeeEvaluationHandler) {
	evaluations := router.Group("/employee-evaluations")
	{
		// Form data endpoint
		evaluations.GET("/form-data", middleware.RequirePermission("evaluation.read"), handler.GetFormData)

		// CRUD routes
		evaluations.GET("", middleware.RequirePermission("evaluation.read"), handler.GetAll)
		evaluations.GET("/:id/audit-trail", middleware.RequirePermission("evaluation.audit_trail"), handler.AuditTrail)
		evaluations.GET("/:id", middleware.RequirePermission("evaluation.read"), handler.GetByID)
		evaluations.POST("", middleware.RequirePermission("evaluation.create"), handler.Create)
		evaluations.PUT("/:id", middleware.RequirePermission("evaluation.update"), handler.Update)
		evaluations.DELETE("/:id", middleware.RequirePermission("evaluation.delete"), handler.Delete)
	}
}
