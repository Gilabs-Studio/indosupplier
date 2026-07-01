package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterProfileRoutes(rg *gin.RouterGroup, h *handler.ProfileHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/profile")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.GetProfile)
		g.PUT("/personal", h.UpdatePersonal)
		g.PUT("/company", h.UpdateCompany)
		g.GET("/documents", h.ListDocuments)
		g.POST("/documents", h.UploadDocument)
	}
}
