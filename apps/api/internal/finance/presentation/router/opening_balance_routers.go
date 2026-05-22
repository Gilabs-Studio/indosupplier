package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	openingBalanceReadPermission     = "opening_balance.read"
	openingBalanceWritePermission    = "opening_balance.write"
	openingBalanceValidatePermission = "opening_balance.validate"
	openingBalanceSimulatePermission = "opening_balance.validate"
	openingBalancePostPermission     = "opening_balance.post"
)

func RegisterOpeningBalanceRoutes(group *gin.RouterGroup, h *handler.OpeningBalanceHandler) {
	opening := group.Group("/opening-balance")
	opening.GET("", middleware.RequirePermission(openingBalanceReadPermission), h.Get)
	opening.PUT("", middleware.RequirePermission(openingBalanceWritePermission), h.Upsert)
	opening.POST("/validate", middleware.RequirePermission(openingBalanceValidatePermission), h.Validate)
	opening.POST("/simulate", middleware.RequirePermission(openingBalanceSimulatePermission), h.Simulate)
	opening.POST("/post", middleware.RequirePermission(openingBalancePostPermission), h.Post)
	opening.GET("/summary", middleware.RequirePermission(openingBalanceReadPermission), h.Summary)
}
