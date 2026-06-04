package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/platform/presentation/handler"
)

func RegisterPlatformRoutes(rg *gin.RouterGroup, h *handler.PlatformHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/platform")
	{
		g.GET("/feature-status", h.FeatureStatus)
		g.GET("/catalog", h.Catalog)
	}

	protected := rg.Group("/platform")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		protected.GET("/dashboard/:persona", h.Dashboard)
	}
}
