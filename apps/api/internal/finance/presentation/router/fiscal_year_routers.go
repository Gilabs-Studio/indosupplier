package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	fiscalYearReadPermission   = "fiscal_year.read"
	fiscalYearWritePermission  = "fiscal_year.write"
	fiscalYearDeletePermission = "fiscal_year.delete"
)

func RegisterFiscalYearRoutes(group *gin.RouterGroup, h *handler.FiscalYearHandler) {
	fiscalYears := group.Group("/fiscal-years")
	fiscalYears.GET("", middleware.RequirePermission(fiscalYearReadPermission), h.List)
	fiscalYears.POST("", middleware.RequirePermission(fiscalYearWritePermission), h.Create)
	fiscalYears.GET("/:id", middleware.RequirePermission(fiscalYearReadPermission), h.GetByID)
	fiscalYears.PUT("/:id", middleware.RequirePermission(fiscalYearWritePermission), h.Update)
	fiscalYears.DELETE("/:id", middleware.RequirePermission(fiscalYearDeletePermission), h.Delete)
	fiscalYears.PATCH("/:id/activate", middleware.RequirePermission(fiscalYearWritePermission), h.Activate)
	fiscalYears.PATCH("/:id/lock", middleware.RequirePermission(fiscalYearWritePermission), h.Lock)
}
