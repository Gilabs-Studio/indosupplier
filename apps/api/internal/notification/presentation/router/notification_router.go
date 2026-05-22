package router

import (
	"github.com/gilabs/gims/api/internal/notification/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(r *gin.RouterGroup, h *handler.NotificationHandler) {
	r.GET("/ws/notifications", h.SubscribeWS)

	n := r.Group("/notifications")
	n.GET("", h.List)
	n.GET("/unread-count", h.GetUnreadCount)
	n.POST("/read-all", h.MarkAllAsRead)
	n.POST("/:id/read", h.MarkAsRead)
}
