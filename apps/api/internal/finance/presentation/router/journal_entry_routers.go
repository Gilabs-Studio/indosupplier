package router

import (
	"net/http"

	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	journalRead              = "journal.read"
	journalCreate            = "journal.create"
	journalUpdate            = "journal.update"
	journalDelete            = "journal.delete"
	journalPost              = "journal.post"
	journalCancel            = "journal.cancel"
	journalReverse           = "journal.reverse"
	salesJournalRead         = "sales_journal.read"
	purchaseJournalRead      = "purchase_journal.read"
	adjustmentJournalRead    = "adjustment_journal.read"
	adjustmentJournalCreate  = "adjustment_journal.create"
	adjustmentJournalSubmit  = "adjustment_journal.submit"
	adjustmentJournalApprove = "adjustment_journal.approve"
	adjustmentJournalReject  = "adjustment_journal.reject"
	adjustmentJournalUpdate  = "adjustment_journal.update"
	adjustmentJournalPost    = "adjustment_journal.post"
	adjustmentJournalReverse = "adjustment_journal.reverse"
	journalTemplateRead      = "journal_template.read"
	journalTemplateCreate    = "journal_template.create"
	journalTemplateUse       = "journal_template.use"
	cashBankJournalRead      = "cash_bank_journal.read"
)

func deprecatedJournalRoute(message string, replacement string) gin.HandlerFunc {
	return func(c *gin.Context) {
		details := map[string]interface{}{}
		if replacement != "" {
			details["replacement"] = replacement
		}

		response.ErrorResponse(
			c,
			http.StatusGone,
			"FINANCE_ROUTE_DEPRECATED",
			message,
			details,
			nil,
		)
	}
}

func RegisterJournalEntryRoutes(rg *gin.RouterGroup, h *handler.JournalEntryHandler) {
	registerJournalEntryRoutesInGroup(rg.Group("/journal-entries"), h)
	registerJournalEntryRoutesInGroup(rg.Group("/journals"), h)
	registerJournalEntryRoutesInGroup(rg.Group("/accounting/journal-entries"), h)
	registerJournalEntryRoutesInGroup(rg.Group("/accounting/journals"), h)
	registerJournalTemplateRoutes(rg.Group("/journal-templates"), h)
}

func registerJournalTemplateRoutes(g *gin.RouterGroup, h *handler.JournalEntryHandler) {
	g.GET("", middleware.RequirePermission(journalTemplateRead), h.ListJournalTemplates)
	g.GET("/", middleware.RequirePermission(journalTemplateRead), h.ListJournalTemplates)
	g.POST("", middleware.RequirePermission(journalTemplateCreate), h.CreateJournalTemplate)
	g.POST("/", middleware.RequirePermission(journalTemplateCreate), h.CreateJournalTemplate)
	g.POST("/:id/use", middleware.RequirePermission(journalTemplateUse), h.UseJournalTemplate)
}

func registerJournalEntryRoutesInGroup(g *gin.RouterGroup, h *handler.JournalEntryHandler) {
	g.Use(middleware.CheckActiveFiscalYear())

	// CRITICAL: Place form-data BEFORE parameterized routes (/:id) for route specificity
	g.GET("/form-data", middleware.RequirePermission(journalRead), h.GetFormData)
	g.GET("", middleware.RequirePermission(journalRead), h.List)
	g.GET("/", middleware.RequirePermission(journalRead), h.List)
	g.POST("", middleware.RequirePermission(journalCreate), h.Create)
	g.POST("/", middleware.RequirePermission(journalCreate), h.Create)

	// Domain-specific read-only journal endpoints
	g.GET("/sales", middleware.RequirePermission(salesJournalRead), h.ListSalesJournals)
	g.GET("/purchase", middleware.RequirePermission(purchaseJournalRead), h.ListPurchaseJournals)
	g.GET("/cash-bank", middleware.RequirePermission(cashBankJournalRead), h.ListCashBankSubLedger)
	g.Any("/inventory", deprecatedJournalRoute("Inventory journal endpoint was moved out of Finance module.", "/stock"))
	g.Any("/inventory/*path", deprecatedJournalRoute("Inventory journal endpoint was moved out of Finance module.", "/stock"))
	g.Any("/valuation", deprecatedJournalRoute("Journal valuation endpoint was moved out of Finance module.", "/stock"))
	g.Any("/valuation/*path", deprecatedJournalRoute("Journal valuation endpoint was moved out of Finance module.", "/stock"))

	// Adjustment journal endpoints (operational, Finance-controlled)
	g.GET("/adjustment", middleware.RequirePermission(adjustmentJournalRead), h.ListAdjustmentJournals)
	g.POST("/adjustment", middleware.RequirePermission(adjustmentJournalCreate), h.CreateAdjustment)
	g.PUT("/adjustment/:id", middleware.RequirePermission(adjustmentJournalUpdate), h.UpdateAdjustment)
	g.PATCH("/adjustment/:id/submit-approval", middleware.RequirePermission(adjustmentJournalSubmit), h.SubmitAdjustmentApproval)
	g.PATCH("/adjustment/:id/approve", middleware.RequirePermission(adjustmentJournalApprove), h.ApproveAdjustment)
	g.PATCH("/adjustment/:id/reject", middleware.RequirePermission(adjustmentJournalReject), h.RejectAdjustment)
	g.GET("/adjustment/:id/approvals", middleware.RequirePermission(adjustmentJournalRead), h.GetAdjustmentApprovalHistory)
	g.POST("/adjustment/:id/post", middleware.RequirePermission(adjustmentJournalPost), h.PostAdjustment)
	g.POST("/adjustment/:id/reverse", middleware.RequirePermission(adjustmentJournalReverse), h.ReverseAdjustment)
	g.PATCH("/:id/submit-approval", middleware.RequirePermission(adjustmentJournalSubmit), h.SubmitAdjustmentApproval)
	g.PATCH("/:id/approve", middleware.RequirePermission(adjustmentJournalApprove), h.ApproveAdjustment)
	g.PATCH("/:id/reject", middleware.RequirePermission(adjustmentJournalReject), h.RejectAdjustment)
	g.GET("/:id/approvals", middleware.RequirePermission(adjustmentJournalRead), h.GetAdjustmentApprovalHistory)
	g.GET("/:id", middleware.RequirePermission(journalRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(journalUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(journalDelete), h.Delete)
	g.POST("/:id/post", middleware.RequirePermission(journalPost), h.Post)
	g.PATCH("/:id/post", middleware.RequirePermission(journalPost), h.Post)
	g.PATCH("/:id/cancel", middleware.RequirePermission(journalCancel), h.Cancel)
	g.POST("/:id/cancel", middleware.RequirePermission(journalCancel), h.Cancel)
	g.POST("/:id/reverse", middleware.RequirePermission(journalReverse), h.Reverse)
	g.PATCH("/:id/reverse", middleware.RequirePermission(journalReverse), h.Reverse)
}
