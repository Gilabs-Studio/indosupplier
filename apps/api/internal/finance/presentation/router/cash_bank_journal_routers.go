package router

import (
	"net/http"

	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	cashBankRead = "cash_bank.read"
)

// HARDENING (CBJ-003): Write operations are deprecated.
// Cash & Bank transactions must originate from operational modules
// (Banking & Payments, Sales, Purchase) and flow through PostOrUpdateJournal().
// Only read endpoints remain for monitoring/audit purposes.
func deprecatedCashBankWrite(c *gin.Context) {
	response.ErrorResponse(c, http.StatusMethodNotAllowed,
		"CASH_BANK_WRITE_DEPRECATED",
		"Direct cash/bank journal creation is deprecated. Use Banking & Payments module for cash transactions. All journals are auto-generated from operational modules.",
		nil, nil,
	)
}

func RegisterCashBankJournalRoutes(r *gin.RouterGroup, h *handler.CashBankJournalHandler) {
	g := r.Group("/cash-bank")

	// READ-ONLY endpoints (active)
	g.GET("", middleware.RequirePermission(cashBankRead), h.List)
	g.GET("/", middleware.RequirePermission(cashBankRead), h.List)
	g.GET("/form-data", middleware.RequirePermission(cashBankRead), h.GetFormData)
	g.GET("/:id", middleware.RequirePermission(cashBankRead), h.GetByID)
	g.GET("/:id/lines", middleware.RequirePermission(cashBankRead), h.ListLines)

	// DEPRECATED write endpoints — return 405 with deprecation notice
	g.POST("", deprecatedCashBankWrite)
	g.POST("/", deprecatedCashBankWrite)
	g.PUT("/:id", deprecatedCashBankWrite)
	g.DELETE("/:id", deprecatedCashBankWrite)
	g.POST("/:id/post", deprecatedCashBankWrite)
}

