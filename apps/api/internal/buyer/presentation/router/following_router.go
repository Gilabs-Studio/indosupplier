package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterFollowingRoutes(rg *gin.RouterGroup, h *handler.FollowingHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/following")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.DELETE("/:supplierProfileId", h.Delete)
	}
}
