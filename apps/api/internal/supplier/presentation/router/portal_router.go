package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/supplier/presentation/handler"
)

func RegisterSupplierPortalRoutes(rg *gin.RouterGroup, h *handler.SupplierPortalHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/supplier")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("/dashboard", h.GetDashboard)
		g.GET("/profile", h.GetProfile)
		g.PUT("/profile", h.UpdateProfile)
		g.GET("/billing-overview", h.GetBillingOverview)
		g.POST("/subscription/upgrade", h.UpgradeSubscriptionPlan)
	}
}
