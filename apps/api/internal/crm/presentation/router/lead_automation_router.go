package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterLeadAutomationRoutes registers routes used to run n8n lead generation through backend API.
func RegisterLeadAutomationRoutes(r *gin.RouterGroup, h *handler.LeadAutomationHandler) {
	g := r.Group("/leads/automation")

	g.POST("/test-connection", middleware.RequirePermission(leadRead), h.TestConnection)
	g.POST("/trigger", middleware.RequirePermission(leadCreate), h.Trigger)
}
