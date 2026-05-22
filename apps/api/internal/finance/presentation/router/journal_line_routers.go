package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	journalLineRead = "journal_line.read"
)

// RegisterJournalLineRoutes registers the journal lines sub-ledger view routes.
// These MUST be registered BEFORE /:id routes to avoid Gin path conflict.
func RegisterJournalLineRoutes(rg *gin.RouterGroup, h *handler.JournalLineHandler) {
	g := rg.Group("/journal-lines")
	g.GET("", middleware.RequirePermission(journalLineRead), h.ListLines)
	g.GET("/", middleware.RequirePermission(journalLineRead), h.ListLines)
	g.GET("/export", middleware.RequirePermission(journalLineRead), h.ExportLines)
}
