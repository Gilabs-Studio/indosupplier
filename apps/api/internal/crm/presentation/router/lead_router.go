package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	leadRead    = "crm_lead.read"
	leadCreate  = "crm_lead.create"
	leadUpdate  = "crm_lead.update"
	leadDelete  = "crm_lead.delete"
	leadConvert = "crm_lead.convert"
)

// RegisterLeadRoutes registers all lead-related routes
func RegisterLeadRoutes(r *gin.RouterGroup, h *handler.LeadHandler) {
	g := r.Group("/leads")

	// Static routes first (before parameterized routes)
	g.GET("/form-data", middleware.RequirePermission(leadRead), h.GetFormData)
	g.GET("/analytics", middleware.RequirePermission(leadRead), h.GetAnalytics)
	g.GET("/unprocessed", middleware.RequirePermission(leadRead), h.GetUnprocessed)
	g.POST("/upsert", middleware.RequirePermission(leadCreate), h.BulkUpsert)

	// CRUD routes
	g.GET("", middleware.RequirePermission(leadRead), h.List)
	g.POST("", middleware.RequirePermission(leadCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(leadRead), h.GetByID)
	g.GET("/:id/product-items", middleware.RequirePermission(leadRead), h.GetProductItems)
	g.PUT("/:id", middleware.RequirePermission(leadUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(leadDelete), h.Delete)

	// Special actions
	g.POST("/:id/convert", middleware.RequirePermission(leadConvert), h.Convert)
}
