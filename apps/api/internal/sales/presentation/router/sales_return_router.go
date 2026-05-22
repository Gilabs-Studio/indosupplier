package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	salesReturnReadPermission   = "sales_return.read"
	salesReturnCreatePermission = "sales_return.create"
	salesReturnUpdatePermission = "sales_return.update"
	salesReturnDeletePermission = "sales_return.delete"
)

func RegisterSalesReturnRoutes(rg *gin.RouterGroup, h *handler.SalesReturnHandler) {
	g := rg.Group("/returns")
	g.GET("/form-data", middleware.RequirePermission(salesReturnReadPermission), h.GetFormData)
	g.GET("", middleware.RequirePermission(salesReturnReadPermission), h.List)
	g.GET("/:id/audit-trail", middleware.RequirePermission(salesReturnReadPermission), h.AuditTrail)
	g.POST("", middleware.RequirePermission(salesReturnCreatePermission), h.Create)
	g.PUT("/:id", middleware.RequirePermission(salesReturnUpdatePermission), h.Update)
	g.PATCH("/:id/status", middleware.RequirePermission(salesReturnUpdatePermission), h.UpdateStatus)
	g.DELETE("/:id", middleware.RequirePermission(salesReturnDeletePermission), h.Delete)
	g.GET("/:id", middleware.RequirePermission(salesReturnReadPermission), h.GetByID)
}
