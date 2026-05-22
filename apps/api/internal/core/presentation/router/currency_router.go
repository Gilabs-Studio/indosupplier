package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterCurrencyRoutes(rg *gin.RouterGroup, h *handler.CurrencyHandler) {
	g := rg.Group("/currencies")
	{
		g.GET("", middleware.RequirePermission("currency.read"), h.List)
		g.GET("/:id", middleware.RequirePermission("currency.read"), h.GetByID)
		g.POST("", middleware.RequireSystemAdmin(), middleware.RequirePermission("currency.create"), h.Create)
		g.PUT("/:id", middleware.RequireSystemAdmin(), middleware.RequirePermission("currency.update"), h.Update)
		g.DELETE("/:id", middleware.RequireSystemAdmin(), middleware.RequirePermission("currency.delete"), h.Delete)
	}
}
