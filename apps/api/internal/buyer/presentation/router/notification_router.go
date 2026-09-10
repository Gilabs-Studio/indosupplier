package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterNotificationRoutes(rg *gin.RouterGroup, h *handler.NotificationHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/notifications")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.List)
		g.POST("/mark-read", h.MarkAllRead)
	}
}
