package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterCompareRoutes(rg *gin.RouterGroup, h *handler.CompareHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/compare")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.List)
		g.POST("", h.Add)
		g.DELETE("/:id", h.Delete)

		g.GET("/products", h.ListProducts)
		g.POST("/products", h.AddProduct)
		g.DELETE("/products/:id", h.DeleteProduct)
	}
}
