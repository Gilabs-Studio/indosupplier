package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	leadSourceRead   = "crm_lead_source.read"
	leadSourceCreate = "crm_lead_source.create"
	leadSourceUpdate = "crm_lead_source.update"
	leadSourceDelete = "crm_lead_source.delete"
)

// RegisterLeadSourceRoutes registers lead source routes
func RegisterLeadSourceRoutes(r *gin.RouterGroup, h *handler.LeadSourceHandler) {
	g := r.Group("/lead-sources")
	g.GET("", middleware.RequirePermission(leadSourceRead), h.List)
	g.POST("", middleware.RequirePermission(leadSourceCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(leadSourceRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(leadSourceUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(leadSourceDelete), h.Delete)
}
