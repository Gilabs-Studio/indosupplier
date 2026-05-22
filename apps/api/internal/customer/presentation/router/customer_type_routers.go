package router

import (
	"github.com/gilabs/gims/api/internal/customer/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCustomerTypeRoutes registers customer type routes
func RegisterCustomerTypeRoutes(rg *gin.RouterGroup, h *handler.CustomerTypeHandler) {
	g := rg.Group("/customer-types")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
