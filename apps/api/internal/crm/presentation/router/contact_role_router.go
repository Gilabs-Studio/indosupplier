package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	contactRoleRead   = "crm_contact_role.read"
	contactRoleCreate = "crm_contact_role.create"
	contactRoleUpdate = "crm_contact_role.update"
	contactRoleDelete = "crm_contact_role.delete"
)

// RegisterContactRoleRoutes registers contact role routes
func RegisterContactRoleRoutes(r *gin.RouterGroup, h *handler.ContactRoleHandler) {
	g := r.Group("/contact-roles")
	g.GET("", middleware.RequirePermission(contactRoleRead), h.List)
	g.POST("", middleware.RequirePermission(contactRoleCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(contactRoleRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(contactRoleUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(contactRoleDelete), h.Delete)
}
