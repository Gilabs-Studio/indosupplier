package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/auth/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/auth")
	{
		g.POST("/login", middleware.RateLimitMiddleware("login"), h.Login)
		g.POST("/register", middleware.RateLimitMiddleware("login"), h.Register)
		g.POST("/refresh-token", middleware.RateLimitMiddleware("refresh"), h.RefreshToken)
		g.GET("/csrf", middleware.RateLimitMiddleware("public"), h.GetCSRFToken)

		protected := g.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.POST("/supplier-profile", h.BecomeSupplier)
			protected.POST("/logout", h.Logout)
		}
	}
}
