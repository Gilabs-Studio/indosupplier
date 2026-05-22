package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	assetLocationRead   = "asset_location.read"
	assetLocationCreate = "asset_location.create"
	assetLocationUpdate = "asset_location.update"
	assetLocationDelete = "asset_location.delete"
)

func RegisterAssetLocationRoutes(r *gin.RouterGroup, h *handler.AssetLocationHandler) {
	registerAssetLocationRoutesInGroup(r.Group("/asset-locations"), h)
	registerAssetLocationRoutesInGroup(r.Group("/fixed-assets/locations"), h)
}

func registerAssetLocationRoutesInGroup(g *gin.RouterGroup, h *handler.AssetLocationHandler) {
	g.GET("", middleware.RequirePermission(assetLocationRead), h.List)
	g.GET("/", middleware.RequirePermission(assetLocationRead), h.List)
	g.POST("", middleware.RequirePermission(assetLocationCreate), h.Create)
	g.POST("/", middleware.RequirePermission(assetLocationCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(assetLocationRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(assetLocationUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(assetLocationDelete), h.Delete)
}
