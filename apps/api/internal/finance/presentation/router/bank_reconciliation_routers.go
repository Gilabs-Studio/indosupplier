package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	bankReconciliationRead      = "bank_reconciliation.read"
	bankReconciliationWrite     = "bank_reconciliation.write"
	bankReconciliationReconcile = "bank_reconciliation.reconcile"
	bankReconciliationLock      = "bank_reconciliation.lock"
)

func RegisterBankReconciliationRoutes(rg *gin.RouterGroup, h *handler.BankReconciliationHandler) {
	g := rg.Group("/bank-reconciliations")

	g.GET("", middleware.RequirePermission(bankReconciliationRead), h.List)
	g.GET("/form-data", middleware.RequirePermission(bankReconciliationRead), h.GetFormData)
	g.POST("/import", middleware.RequirePermission(bankReconciliationWrite), h.Import)
	g.GET("/:id", middleware.RequirePermission(bankReconciliationRead), h.GetByID)
	g.POST("/:id/auto-match", middleware.RequirePermission(bankReconciliationWrite), h.AutoMatch)
	g.POST("/:id/line/:line_id/match", middleware.RequirePermission(bankReconciliationWrite), h.MatchLine)
	g.POST("/:id/line/:line_id/exclude", middleware.RequirePermission(bankReconciliationWrite), h.ExcludeLine)
	g.POST("/:id/confirm", middleware.RequirePermission(bankReconciliationReconcile), h.Confirm)
	g.POST("/:id/lock", middleware.RequirePermission(bankReconciliationLock), h.Lock)
}
