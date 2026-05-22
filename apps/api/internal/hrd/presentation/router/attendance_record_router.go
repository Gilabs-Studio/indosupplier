package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAttendanceRecordRoutes registers attendance record routes
func RegisterAttendanceRecordRoutes(rg *gin.RouterGroup, h *handler.AttendanceRecordHandler, wsH *handler.AttendanceWSHandler) {
	g := rg.Group("/attendance")

	// Employee self-service routes (no special permission needed beyond auth)
	g.GET("/ws/today", wsH.SubscribeToday)
	g.GET("/today", h.GetTodayAttendance)
	g.POST("/clock-in", h.ClockIn)
	g.POST("/clock-out", h.ClockOut)
	g.GET("/my-stats", h.GetMonthlyStats)
	g.GET("/my-history", h.ListMyAttendance)

	// Admin routes
	g.GET("/form-data", middleware.RequirePermission("attendance.read"), h.GetFormData)
	g.GET("/employee-schedule/:employeeId", middleware.RequirePermission("attendance.read"), h.GetEmployeeSchedule)
	g.GET("", middleware.RequirePermission("attendance.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("attendance.read"), h.GetByID)
	g.POST("/manual", middleware.RequirePermission("attendance.create"), h.CreateManualEntry)
	g.POST("/process-absent", middleware.RequirePermission("attendance.create"), h.ProcessAutoAbsent)
	g.PUT("/:id", middleware.RequirePermission("attendance.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("attendance.delete"), h.Delete)
}
