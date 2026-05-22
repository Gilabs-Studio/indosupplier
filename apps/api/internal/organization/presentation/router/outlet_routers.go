package router

import (
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterOutletRoutes registers all outlet routes
func RegisterOutletRoutes(rg *gin.RouterGroup, h *handler.OutletHandler) {
	g := rg.Group("/outlets")
	{
		g.GET("/form-data", h.GetFormData)
		g.GET("/limit", h.GetLimit)
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
