package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPayableRecapRoutes(r *gin.RouterGroup, h *handler.PayableRecapHandler) {
	g := r.Group("/payable-recap")
	g.GET("", middleware.RequirePermission(purchasePaymentRead), h.List)
	g.GET("/summary", middleware.RequirePermission(purchasePaymentRead), h.Summary)
	g.GET("/export", middleware.RequirePermission(purchasePaymentExport), h.Export)
}
