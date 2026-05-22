package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

// Permission codes must match the seeder (permission_seeder.go) exactly.
const (
	generalLedgerReportRead   = "general_ledger_report.read"
	generalLedgerReportExport = "general_ledger_report.export"
	balanceSheetReportRead    = "balance_sheet_report.read"
	balanceSheetReportExport  = "balance_sheet_report.export"
	profitLossReportRead      = "profit_loss_report.read"
	profitLossReportExport    = "profit_loss_report.export"
	trialBalanceReportRead    = "trial_balance_report.read"
	cashFlowReportRead        = "cash_flow_statement.read"
)

func RegisterFinanceReportExRoutes(rg *gin.RouterGroup, h *handler.FinanceReportHandler) {
	g := rg.Group("/reports")
	g.GET("/general-ledger", middleware.RequirePermission(generalLedgerReportRead), h.GeneralLedger)
	g.GET("/balance-sheet", middleware.RequirePermission(balanceSheetReportRead), h.BalanceSheet)
	g.GET("/profit-loss", middleware.RequirePermission(profitLossReportRead), h.ProfitAndLoss)
	g.GET("/trial-balance", middleware.RequirePermission(trialBalanceReportRead), h.TrialBalance)
	g.GET("/export/general-ledger", middleware.RequirePermission(generalLedgerReportExport), h.ExportGeneralLedger)
	g.GET("/export/balance-sheet", middleware.RequirePermission(balanceSheetReportExport), h.ExportBalanceSheet)
	g.GET("/export/profit-loss", middleware.RequirePermission(profitLossReportExport), h.ExportProfitAndLoss)
	g.GET("/cash-flow-statement", middleware.RequirePermission(cashFlowReportRead), h.CashFlow)
	g.GET("/cash-flow", middleware.RequirePermission(cashFlowReportRead), h.CashFlow)
}
