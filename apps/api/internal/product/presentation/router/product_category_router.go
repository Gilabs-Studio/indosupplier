package router

import (
	"github.com/gilabs/gims/api/internal/product/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductCategoryRoutes(rg *gin.RouterGroup, h *handler.ProductCategoryHandler) {
	g := rg.Group("/product-categories")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/tree", h.GetTree)         // Get hierarchical tree structure
		g.GET("/:id", h.GetByID)
		g.GET("/:id/children", h.GetChildren) // Get direct children (lazy load)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}

