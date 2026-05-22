package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	receivablesRecapRead   = "sales_payment.read"
	receivablesRecapExport = "sales_payment.export"
)

func RegisterReceivablesRecapRoutes(r *gin.RouterGroup, h *handler.ReceivablesRecapHandler) {
	g := r.Group("/receivables-recap")
	g.GET("", middleware.RequirePermission(receivablesRecapRead), h.List)
	g.GET("/summary", middleware.RequirePermission(receivablesRecapRead), h.Summary)
	g.GET("/export", middleware.RequirePermission(receivablesRecapExport), h.Export)
}
