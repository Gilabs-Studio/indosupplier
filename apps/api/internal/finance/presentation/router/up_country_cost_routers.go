package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
)

func RegisterUpCountryCostRoutes(r *gin.RouterGroup, h *handler.UpCountryCostHandler) {
	g := r.Group("/up-country-cost")
	{
		g.GET("", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionView)), h.List)
		g.POST("", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionCreate)), h.Create)
		// CRITICAL: Place fixed-path routes BEFORE /:id to avoid shadowing
		g.GET("/stats", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionView)), h.GetStats)
		g.GET("/form-data", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionView)), h.GetFormData)
		g.GET("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionView)), h.GetByID)
		g.PUT("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionUpdate)), h.Update)
		g.DELETE("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionDelete)), h.Delete)
		// Workflow actions
		g.POST("/:id/submit", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionSubmit)), h.Submit)
		g.POST("/:id/manager-approve", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionApprove)), h.ManagerApprove)
		g.POST("/:id/manager-reject", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionReject)), h.ManagerReject)
		g.POST("/:id/finance-approve", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionApprove)), h.FinanceApprove)
		g.POST("/:id/mark-paid", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionPay)), h.MarkPaid)
		// Legacy approve endpoint (kept for backward compatibility)
		g.POST("/:id/approve", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeUpCountryCost, reference.ActionApprove)), h.Approve)
	}
}
