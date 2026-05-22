package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterBusinessTypeRoutes registers business type routes
func RegisterBusinessTypeRoutes(rg *gin.RouterGroup, h *handler.BusinessTypeHandler) {
	g := rg.Group("/business-types")
	g.GET("", middleware.RequirePermission("business_type.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("business_type.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("business_type.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("business_type.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("business_type.delete"), h.Delete)
}
