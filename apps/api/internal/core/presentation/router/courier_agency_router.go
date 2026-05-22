package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterCourierAgencyRoutes(rg *gin.RouterGroup, h *handler.CourierAgencyHandler) {
	g := rg.Group("/courier-agencies")
	{
		g.POST("", middleware.RequirePermission("courier_agency.create"), h.Create)
		g.GET("", middleware.RequirePermission("courier_agency.read"), h.List)
		g.GET("/:id", middleware.RequirePermission("courier_agency.read"), h.GetByID)
		g.PUT("/:id", middleware.RequirePermission("courier_agency.update"), h.Update)
		g.DELETE("/:id", middleware.RequirePermission("courier_agency.delete"), h.Delete)
	}
}
