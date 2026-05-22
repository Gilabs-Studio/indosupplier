package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	salaryRead    = "salary.read"
	salaryCreate  = "salary.create"
	salaryUpdate  = "salary.update"
	salaryDelete  = "salary.delete"
	salaryApprove = "salary.approve"
)

func RegisterSalaryStructureRoutes(r *gin.RouterGroup, h *handler.SalaryStructureHandler) {
	g := r.Group("/salary")
	{
		g.GET("", middleware.RequirePermission(salaryRead), h.List)
		g.POST("", middleware.RequirePermission(salaryCreate), h.Create)
		g.GET("/stats", middleware.RequirePermission(salaryRead), h.GetStats)
		g.GET("/grouped", middleware.RequirePermission(salaryRead), h.ListGrouped)
		g.GET("/form-data", middleware.RequirePermission(salaryRead), h.GetFormData)
		g.GET("/:id", middleware.RequirePermission(salaryRead), h.GetByID)
		g.PUT("/:id", middleware.RequirePermission(salaryUpdate), h.Update)
		g.POST("/:id/toggle-status", middleware.RequirePermission(salaryUpdate), h.ToggleStatus)
		g.DELETE("/:id", middleware.RequirePermission(salaryDelete), h.Delete)
		g.POST("/:id/approve", middleware.RequirePermission(salaryApprove), h.Approve)
	}
}
