package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	purchaseReturnReadPermission   = "purchase_return.read"
	purchaseReturnCreatePermission = "purchase_return.create"
	purchaseReturnUpdatePermission = "purchase_return.update"
	purchaseReturnDeletePermission = "purchase_return.delete"
)

func RegisterPurchaseReturnRoutes(r *gin.RouterGroup, h *handler.PurchaseReturnHandler) {
	g := r.Group("/returns")
	g.GET("/form-data", middleware.RequirePermission(purchaseReturnReadPermission), h.GetFormData)
	g.GET("", middleware.RequirePermission(purchaseReturnReadPermission), h.List)
	g.GET("/:id/audit-trail", middleware.RequirePermission(purchaseReturnReadPermission), h.AuditTrail)
	g.POST("", middleware.RequirePermission(purchaseReturnCreatePermission), h.Create)
	g.PUT("/:id", middleware.RequirePermission(purchaseReturnUpdatePermission), h.Update)
	g.PATCH("/:id/status", middleware.RequirePermission(purchaseReturnUpdatePermission), h.UpdateStatus)
	g.PATCH("/:id/submit", middleware.RequirePermission(purchaseReturnUpdatePermission), h.Submit)
	g.PATCH("/:id/confirm", middleware.RequirePermission(purchaseReturnUpdatePermission), h.Confirm)
	g.PATCH("/:id/reject", middleware.RequirePermission(purchaseReturnUpdatePermission), h.Reject)
	g.DELETE("/:id", middleware.RequirePermission(purchaseReturnDeletePermission), h.Delete)
	g.GET("/:id", middleware.RequirePermission(purchaseReturnReadPermission), h.GetByID)
}
