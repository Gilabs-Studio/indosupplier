package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

// Contact routes use customer.* permissions — contacts are child entities of customers
// and do not require a separate permission set.
const (
	customerRead   = "customer.read"
	customerCreate = "customer.create"
	customerUpdate = "customer.update"
	customerDelete = "customer.delete"
)

// RegisterContactRoutes registers contact routes
func RegisterContactRoutes(r *gin.RouterGroup, h *handler.ContactHandler) {
	g := r.Group("/contacts")
	g.GET("/form-data", middleware.RequirePermission(customerRead), h.GetFormData)
	g.GET("", middleware.RequirePermission(customerRead), h.List)
	g.POST("", middleware.RequirePermission(customerCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(customerRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(customerUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(customerDelete), h.Delete)
}
