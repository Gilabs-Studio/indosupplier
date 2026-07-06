package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/verification/presentation/handler"
)

func RegisterVerificationRoutes(rg *gin.RouterGroup, h *handler.VerificationHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/supplier/verification")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.GetVerificationData)
		g.PUT("", h.UpdateVerificationData)
		g.POST("/submit", h.SubmitVerification)
	}
}
