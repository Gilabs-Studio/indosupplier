package router

import (
	"github.com/gilabs/gims/api/internal/supplier/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterBankRoutes registers bank routes
func RegisterBankRoutes(rg *gin.RouterGroup, h *handler.BankHandler) {
	g := rg.Group("/banks")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
