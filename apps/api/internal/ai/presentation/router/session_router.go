package router

import (
	"github.com/gilabs/gims/api/internal/ai/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSessionRoutes registers AI session management routes
func RegisterSessionRoutes(r *gin.RouterGroup, sessionHandler *handler.SessionHandler) {
	sessions := r.Group("/sessions")
	{
		sessions.GET("", sessionHandler.ListSessions)
		sessions.GET("/:id", sessionHandler.GetSessionDetail)
		sessions.DELETE("/:id", sessionHandler.DeleteSession)
	}
}
