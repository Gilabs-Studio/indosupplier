package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const agingReportRead = "aging_report.read"

func RegisterFinanceAgingReportRoutes(rg *gin.RouterGroup, h *handler.AgingReportHandler) {
	rg.GET("/ar/aging", middleware.RequirePermission(agingReportRead), h.ARAgingFinance)
	rg.GET("/ap/aging", middleware.RequirePermission(agingReportRead), h.APAgingFinance)

	g := rg.Group("/reports")
	g.GET("/ar-aging", middleware.RequirePermission(agingReportRead), h.ARAging)
	g.GET("/ap-aging", middleware.RequirePermission(agingReportRead), h.APAging)
	g.GET("/aging/ar", middleware.RequirePermission(agingReportRead), h.ARAging)
	g.GET("/aging/ap", middleware.RequirePermission(agingReportRead), h.APAging)
}
