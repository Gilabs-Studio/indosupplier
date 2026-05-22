package router

import (
	"github.com/gilabs/gims/api/internal/product/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterUnitOfMeasureRoutes(rg *gin.RouterGroup, h *handler.UnitOfMeasureHandler) {
	g := rg.Group("/units-of-measure")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
