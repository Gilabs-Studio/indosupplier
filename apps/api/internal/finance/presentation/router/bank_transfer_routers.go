package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	bankTransferRead   = "bank_transfer.read"
	bankTransferWrite  = "bank_transfer.write"
	bankTransferCancel = "bank_transfer.cancel"
)

func RegisterBankTransferRoutes(rg *gin.RouterGroup, h *handler.BankTransferHandler) {
	g := rg.Group("/bank-transfers")

	g.GET("", middleware.RequirePermission(bankTransferRead), h.List)
	g.POST("", middleware.RequirePermission(bankTransferWrite), h.Create)
	g.GET("/:id", middleware.RequirePermission(bankTransferRead), h.GetByID)
	g.POST("/:id/complete", middleware.RequirePermission(bankTransferWrite), h.Complete)
	g.POST("/:id/cancel", middleware.RequirePermission(bankTransferCancel), h.Cancel)
}
