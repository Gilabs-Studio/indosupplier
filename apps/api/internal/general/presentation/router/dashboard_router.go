package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/general/presentation/handler"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterDashboardRoutes registers all dashboard routes under the given group
func RegisterDashboardRoutes(rg *gin.RouterGroup, h *handler.DashboardHandler, db *gorm.DB) {
	g := rg.Group("/dashboard")
	{
		// Inject dashboard-specific scope from role_menu_access before permission check
		g.Use(middleware.InjectDashboardScope(db))
		
		g.GET("/overview", middleware.RequirePermission("dashboard.view"), h.GetOverview)
		g.GET("/layout", middleware.RequirePermission("dashboard.view"), h.GetLayout)
		g.PUT("/layout", middleware.RequirePermission("dashboard.view"), h.SaveLayout)
	}
}
