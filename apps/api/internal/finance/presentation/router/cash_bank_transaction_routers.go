package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	cashBankTransactionRead    = "cash_bank_transaction.read"
	cashBankTransactionWrite   = "cash_bank_transaction.write"
	cashBankTransactionReverse = "cash_bank_transaction.reverse"
)

func RegisterCashBankTransactionRoutes(rg *gin.RouterGroup, h *handler.CashBankTransactionHandler) {
	g := rg.Group("/cash-bank-transactions")

	g.GET("", middleware.RequirePermission(cashBankTransactionRead), h.List)
	g.GET("/form-data", middleware.RequirePermission(cashBankTransactionRead), h.GetFormData)
	g.POST("", middleware.RequirePermission(cashBankTransactionWrite), h.Create)
	g.GET("/:id", middleware.RequirePermission(cashBankTransactionRead), h.GetByID)
	g.POST("/:id/reverse", middleware.RequirePermission(cashBankTransactionReverse), h.Reverse)
}
