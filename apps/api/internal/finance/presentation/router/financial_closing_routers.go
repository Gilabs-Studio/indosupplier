package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	financialClosingRead    = "financial_closing.read"
	financialClosingCreate  = "financial_closing.create"
	financialClosingApprove = "financial_closing.approve"
	financialClosingReopen  = "financial_closing.reopen"
	financialClosingYearEnd = "financial_closing.year_end"
	financialClosingDelete  = "financial_closing.delete"
)

func RegisterFinancialClosingRoutes(r *gin.RouterGroup, h *handler.FinancialClosingHandler) {
	registerFinancialClosingRoutesInGroup(r.Group("/closing"), h)
	registerFinancialClosingRoutesInGroup(r.Group("/accounting/closing"), h)
}

func registerFinancialClosingRoutesInGroup(g *gin.RouterGroup, h *handler.FinancialClosingHandler) {
	g.GET("", middleware.RequirePermission(financialClosingRead), h.List)
	g.GET("/", middleware.RequirePermission(financialClosingRead), h.List)
	g.POST("", middleware.RequirePermission(financialClosingCreate), h.Create)
	g.POST("/", middleware.RequirePermission(financialClosingCreate), h.Create)
	g.POST("/year-end-close", middleware.RequirePermission(financialClosingYearEnd), h.YearEndClose)
	g.GET("/:id", middleware.RequirePermission(financialClosingRead), h.GetByID)
	g.GET("/:id/analysis", middleware.RequirePermission(financialClosingRead), h.GetAnalysis)
	g.POST("/:id/approve", middleware.RequirePermission(financialClosingApprove), h.Approve)
	g.POST("/:id/reopen", middleware.RequirePermission(financialClosingReopen), h.Reopen)
	g.DELETE("/:id", middleware.RequirePermission(financialClosingDelete), h.Delete)
}
