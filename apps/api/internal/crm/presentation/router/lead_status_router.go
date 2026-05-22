package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	leadStatusRead   = "crm_lead_status.read"
	leadStatusCreate = "crm_lead_status.create"
	leadStatusUpdate = "crm_lead_status.update"
	leadStatusDelete = "crm_lead_status.delete"
)

// RegisterLeadStatusRoutes registers lead status routes
func RegisterLeadStatusRoutes(r *gin.RouterGroup, h *handler.LeadStatusHandler) {
	g := r.Group("/lead-statuses")
	g.GET("", middleware.RequirePermission(leadStatusRead), h.List)
	g.POST("", middleware.RequirePermission(leadStatusCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(leadStatusRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(leadStatusUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(leadStatusDelete), h.Delete)
}
