package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterBusinessUnitRoutes registers business unit routes
func RegisterBusinessUnitRoutes(rg *gin.RouterGroup, h *handler.BusinessUnitHandler) {
	g := rg.Group("/business-units")
	g.GET("", middleware.RequirePermission("business_unit.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("business_unit.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("business_unit.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("business_unit.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("business_unit.delete"), h.Delete)
}
