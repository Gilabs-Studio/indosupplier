package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	assetCategoryRead   = "asset_category.read"
	assetCategoryCreate = "asset_category.create"
	assetCategoryUpdate = "asset_category.update"
	assetCategoryDelete = "asset_category.delete"
)

func RegisterAssetCategoryRoutes(r *gin.RouterGroup, h *handler.AssetCategoryHandler) {
	registerAssetCategoryRoutesInGroup(r.Group("/asset-categories"), h)
	registerAssetCategoryRoutesInGroup(r.Group("/fixed-assets/categories"), h)
}

func registerAssetCategoryRoutesInGroup(g *gin.RouterGroup, h *handler.AssetCategoryHandler) {
	g.GET("", middleware.RequirePermission(assetCategoryRead), h.List)
	g.GET("/", middleware.RequirePermission(assetCategoryRead), h.List)
	g.POST("", middleware.RequirePermission(assetCategoryCreate), h.Create)
	g.POST("/", middleware.RequirePermission(assetCategoryCreate), h.Create)
	// CRITICAL: Place form-data BEFORE parameterized routes (/:id) for route specificity
	g.GET("/form-data", middleware.RequirePermission(assetCategoryRead), h.GetFormData)
	g.GET("/:id", middleware.RequirePermission(assetCategoryRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(assetCategoryUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(assetCategoryDelete), h.Delete)
}
