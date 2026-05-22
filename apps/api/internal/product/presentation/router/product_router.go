package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/product/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(rg *gin.RouterGroup, h *handler.ProductHandler) {
	g := rg.Group("/products")
	{
		g.POST("", middleware.RequirePermission("product.create"), h.Create)
		g.GET("", middleware.RequirePermission("product.read"), h.List)
		// Recipe endpoints BEFORE /:id to avoid Gin route conflict
		g.GET("/:id/recipe", middleware.RequirePermission("product.read"), h.GetRecipe)
		g.PUT("/:id/recipe", middleware.RequirePermission("product.recipe.update"), h.UpdateRecipe)
		g.GET("/:id/recipe/versions", middleware.RequirePermission("product.read"), h.ListRecipeVersions)
		g.POST("/:id/recipe/clone", middleware.RequirePermission("product.recipe.update"), h.CloneRecipeFromVersion)
		g.GET("/:id/recipe/compare", middleware.RequirePermission("product.read"), h.CompareRecipeVersions)
		g.GET("/:id", middleware.RequirePermission("product.read"), h.GetByID)
		g.PUT("/:id", middleware.RequirePermission("product.update"), h.Update)
		g.DELETE("/:id", middleware.RequirePermission("product.delete"), h.Delete)
		g.POST("/:id/submit", middleware.RequirePermission("product.update"), h.Submit)
		g.POST("/:id/approve", middleware.RequirePermission("product.approve"), h.Approve)
	}
}
