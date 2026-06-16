package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/chat/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterChatRoutes(rg *gin.RouterGroup, h *handler.ChatHandler, jwtManager *jwt.JWTManager) {
	// REST Routes
	g := rg.Group("/buyer/chat/rooms")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.ListRooms)
		g.POST("", h.GetOrCreateRoom)
		g.GET("/:id/messages", h.GetRoomMessages)
		g.POST("/:id/messages", h.SendMessage)
		g.POST("/:id/read", h.MarkAsRead)
	}

	// WebSocket Route (WebSocket manual auth check inside handler)
	rg.GET("/chat/ws", h.ConnectWebSocket)
}
