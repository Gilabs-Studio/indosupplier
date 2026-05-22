package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	coaRead   = "coa.read"
	coaCreate = "coa.create"
	coaUpdate = "coa.update"
	coaDelete = "coa.delete"
)

func RegisterChartOfAccountRoutes(rg *gin.RouterGroup, h *handler.ChartOfAccountHandler) {
	registerChartOfAccountRoutesInGroup(rg.Group("/chart-of-accounts"), h)
	registerChartOfAccountRoutesInGroup(rg.Group("/accounting/coa"), h)
}

func registerChartOfAccountRoutesInGroup(g *gin.RouterGroup, h *handler.ChartOfAccountHandler) {
	g.GET("/tree", middleware.RequirePermission(coaRead), h.Tree)
	g.GET("", middleware.RequirePermission(coaRead), h.List)
	g.GET("/", middleware.RequirePermission(coaRead), h.List)
	g.POST("", middleware.RequirePermission(coaCreate), h.Create)
	g.POST("/", middleware.RequirePermission(coaCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(coaRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(coaUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(coaDelete), h.Delete)
}
