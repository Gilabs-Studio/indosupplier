package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterNonTradePayableRoutes(r *gin.RouterGroup, h *handler.NonTradePayableHandler) {
	registerNonTradePayableRoutesInGroup(r.Group("/non-trade-payables"), h)
	registerNonTradePayableRoutesInGroup(r.Group("/ap/non-trade-payables"), h)
}

func registerNonTradePayableRoutesInGroup(g *gin.RouterGroup, h *handler.NonTradePayableHandler) {
	g.GET("", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionView)), h.List)
	g.GET("/", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionView)), h.List)
	g.POST("", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionCreate)), h.Create)
	g.POST("/", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionCreate)), h.Create)
	// CRITICAL: Place form-data BEFORE parameterized routes (/:id) for route specificity
	g.GET("/form-data", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionView)), h.GetFormData)
	g.GET("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionView)), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionUpdate)), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionDelete)), h.Delete)

	// Canonical lifecycle endpoints (Phase 8)
	g.POST("/:id/post", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionSubmit)), h.Post)
	g.POST("/:id/cancel", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionReject)), h.Cancel)
	g.POST("/:id/payments", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionPay)), h.PostPayment)

	// Backward-compatible aliases
	g.POST("/:id/submit", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionSubmit)), h.Submit)
	g.POST("/:id/approve", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionApprove)), h.Approve)
	g.POST("/:id/reject", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionReject)), h.Reject)
	g.POST("/:id/pay", middleware.RequirePermission(reference.PermissionKey(reference.RefTypeNonTradePayable, reference.ActionPay)), h.Pay)
}
