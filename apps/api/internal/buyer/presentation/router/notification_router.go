package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterNotificationRoutes(rg *gin.RouterGroup, h *handler.NotificationHandler, jwtManager *jwt.JWTManager) {
	buyerGroup := rg.Group("/buyer/notifications")
	buyerGroup.Use(middleware.AuthMiddleware(jwtManager))
	{
		buyerGroup.GET("", h.List)
		buyerGroup.POST("/mark-read", h.MarkAllRead)
	}

	supplierGroup := rg.Group("/supplier/notifications")
	supplierGroup.Use(middleware.AuthMiddleware(jwtManager))
	{
		supplierGroup.GET("", h.List)
		supplierGroup.POST("/mark-read", h.MarkAllRead)
	}
}
