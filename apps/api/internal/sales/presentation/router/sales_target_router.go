package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSalesTargetRoutes registers sales target routes
func RegisterSalesTargetRoutes(rg *gin.RouterGroup, h *handler.SalesTargetHandler) {
	g := rg.Group("/sales-targets")
	g.GET("", middleware.RequirePermission("sales_target.read"), h.List)
	g.GET("/available-employees", middleware.RequirePermission("sales_target.create"), h.ListAvailableEmployees)
	g.GET("/:id", middleware.RequirePermission("sales_target.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("sales_target.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("sales_target.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("sales_target.delete"), h.Delete)
}
