package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterBookmarkRoutes(rg *gin.RouterGroup, h *handler.BookmarkHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/bookmarks")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.POST("/toggle", h.Toggle)
		g.DELETE("/:id", h.Delete)
	}
}
