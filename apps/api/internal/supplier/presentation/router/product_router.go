package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/supplier/presentation/handler"
)

func RegisterProductRoutes(rg *gin.RouterGroup, h *handler.ProductHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/supplier/products")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}

	cats := rg.Group("/categories")
	{
		cats.GET("", h.ListCategories)
	}
}
