package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	budgetRead    = "budget.read"
	budgetCreate  = "budget.create"
	budgetUpdate  = "budget.update"
	budgetDelete  = "budget.delete"
	budgetApprove = "budget.approve"
)

func RegisterBudgetRoutes(r *gin.RouterGroup, h *handler.BudgetHandler) {
	g := r.Group("/budget")
	g.GET("", middleware.RequirePermission(budgetRead), h.List)
	g.GET("/", middleware.RequirePermission(budgetRead), h.List)
	g.POST("", middleware.RequirePermission(budgetCreate), h.Create)
	g.POST("/", middleware.RequirePermission(budgetCreate), h.Create)
	// CRITICAL: Place form-data BEFORE parameterized routes (/:id) for route specificity
	g.GET("/form-data", middleware.RequirePermission(budgetRead), h.GetFormData)
	g.GET("/:id", middleware.RequirePermission(budgetRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(budgetUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(budgetDelete), h.Delete)
	g.POST("/:id/approve", middleware.RequirePermission(budgetApprove), h.Approve)
	g.POST("/:id/sync", middleware.RequirePermission(budgetUpdate), h.SyncActuals)
}
