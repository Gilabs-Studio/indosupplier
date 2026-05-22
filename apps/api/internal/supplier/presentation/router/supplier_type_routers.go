package router

import (
	"github.com/gilabs/gims/api/internal/supplier/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSupplierTypeRoutes registers supplier type routes
func RegisterSupplierTypeRoutes(rg *gin.RouterGroup, h *handler.SupplierTypeHandler) {
	g := rg.Group("/supplier-types")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
