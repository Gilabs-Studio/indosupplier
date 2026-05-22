package router

import (
	"github.com/gilabs/gims/api/internal/ai/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterChatRoutes registers AI chat routes (legacy pipeline)
func RegisterChatRoutes(r *gin.RouterGroup, chatHandler *handler.ChatHandler) {
	// Models endpoint (before /chat group for route specificity)
	r.GET("/models", chatHandler.ListModels)

	chat := r.Group("/chat")
	{
		chat.POST("/send", chatHandler.SendMessage)
		chat.POST("/confirm", chatHandler.ConfirmAction)
	}
}

// RegisterV2ChatRoutes registers engine-based routes with streaming support.
// These run alongside the legacy routes for backward compatibility.
func RegisterV2ChatRoutes(r *gin.RouterGroup, streamHandler *handler.StreamHandler) {
	r.GET("/v2/models", streamHandler.ListModelsV2)

	v2Chat := r.Group("/chat/v2")
	{
		v2Chat.POST("/send", streamHandler.SendMessageV2)
		v2Chat.POST("/stream", streamHandler.StreamMessage)
		v2Chat.POST("/confirm", streamHandler.ConfirmActionV2)
	}
}
