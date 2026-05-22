package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterHolidayRoutes registers holiday routes
func RegisterHolidayRoutes(rg *gin.RouterGroup, h *handler.HolidayHandler) {
	g := rg.Group("/holidays")
	g.GET("", middleware.RequirePermission("holiday.read"), h.List)
	g.GET("/check", middleware.RequirePermission("holiday.read"), h.CheckHoliday)
	g.GET("/year/:year", middleware.RequirePermission("holiday.read"), h.GetByYear)
	g.GET("/calendar/:year", middleware.RequirePermission("holiday.read"), h.GetCalendar)
	g.GET("/:id", middleware.RequirePermission("holiday.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("holiday.create"), h.Create)
	g.POST("/batch", middleware.RequirePermission("holiday.create"), h.CreateBatch)
	g.PUT("/:id", middleware.RequirePermission("holiday.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("holiday.delete"), h.Delete)
}
