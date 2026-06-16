package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterReviewRoutes(rg *gin.RouterGroup, h *handler.ReviewHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/reviews")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("/eligible", h.ListEligible)
		g.GET("/history", h.ListHistory)
		g.POST("", h.Create)
	}
}
