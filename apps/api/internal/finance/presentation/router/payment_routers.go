package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPaymentRoutes(r *gin.RouterGroup, h *handler.PaymentHandler) {
	registerPaymentRoutesInGroup(r.Group("/payments"), h)
	registerPaymentRoutesInGroup(r.Group("/ap/payments"), h)
}

func registerPaymentRoutesInGroup(g *gin.RouterGroup, h *handler.PaymentHandler) {
	g.GET("", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionView)), h.List)
	g.GET("/", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionView)), h.List)
	g.POST("", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionCreate)), h.Create)
	g.POST("/", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionCreate)), h.Create)
	// CRITICAL: Place form-data BEFORE parameterized routes (/:id) for route specificity
	g.GET("/form-data", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionView)), h.GetFormData)
	g.GET("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionView)), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionUpdate)), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionDelete)), h.Delete)
	g.POST("/:id/approve", middleware.RequirePermission(reference.PermissionKey(reference.RefTypePayment, reference.ActionApprove)), h.Approve)
}
