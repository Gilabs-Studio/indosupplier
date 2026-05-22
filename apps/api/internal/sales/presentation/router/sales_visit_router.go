package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/sales/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSalesVisitRoutes registers sales visit routes
func RegisterSalesVisitRoutes(rg *gin.RouterGroup, h *handler.SalesVisitHandler) {
	g := rg.Group("/sales-visits")
	g.GET("", middleware.RequirePermission("sales_visit.read"), h.List)
	g.GET("/calendar", middleware.RequirePermission("sales_visit.read"), h.GetCalendarSummary)
	g.GET("/:id", middleware.RequirePermission("sales_visit.read"), h.GetByID)
	g.GET("/:id/details", middleware.RequirePermission("sales_visit.read"), h.ListDetails)
	g.GET("/:id/history", middleware.RequirePermission("sales_visit.read"), h.ListProgressHistory)
	g.POST("", middleware.RequirePermission("sales_visit.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("sales_visit.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("sales_visit.delete"), h.Delete)
	g.PATCH("/:id/status", middleware.RequirePermission("sales_visit.update"), h.UpdateStatus)
	g.POST("/:id/check-in", middleware.RequirePermission("sales_visit.update"), h.CheckIn)
	g.POST("/:id/check-out", middleware.RequirePermission("sales_visit.update"), h.CheckOut)
}
