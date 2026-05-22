package router

import (
	"github.com/gilabs/gims/api/internal/ai/presentation/handler"
	middleware "github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes registers AI admin routes (action logs, intent registry)
func RegisterAdminRoutes(r *gin.RouterGroup, adminHandler *handler.AdminHandler) {
	admin := r.Group("/admin")
	admin.Use(middleware.RequirePermission("ai.admin"))
	{
		admin.GET("/actions", adminHandler.ListActions)
		admin.GET("/intents", adminHandler.GetIntentRegistry)
	}
}
