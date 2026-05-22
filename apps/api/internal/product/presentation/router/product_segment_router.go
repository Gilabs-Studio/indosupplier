package router

import (
	"github.com/gilabs/gims/api/internal/product/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductSegmentRoutes(rg *gin.RouterGroup, h *handler.ProductSegmentHandler) {
	g := rg.Group("/product-segments")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
