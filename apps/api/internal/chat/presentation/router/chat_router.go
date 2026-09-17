package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/chat/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterChatRoutes(rg *gin.RouterGroup, h *handler.ChatHandler, jwtManager *jwt.JWTManager) {
	// REST Routes
	chatGroup := rg.Group("/buyer/chat")
	chatGroup.Use(middleware.AuthMiddleware(jwtManager))
	{
		chatGroup.GET("/ws-token", h.GetWebSocketToken)
		chatGroup.GET("/rooms", h.ListRooms)
		chatGroup.POST("/rooms", h.GetOrCreateRoom)
		chatGroup.GET("/rooms/:id/messages", h.GetRoomMessages)
		chatGroup.POST("/rooms/:id/messages", h.SendMessage)
		chatGroup.POST("/rooms/:id/read", h.MarkAsRead)
	}

	// WebSocket Route (WebSocket manual auth check inside handler)
	rg.GET("/chat/ws", h.ConnectWebSocket)
}
