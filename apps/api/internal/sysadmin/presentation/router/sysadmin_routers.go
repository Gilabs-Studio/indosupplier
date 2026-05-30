package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/presentation/middleware"
)

func RegisterSysadminAuthRoutes(rg *gin.RouterGroup, h *handler.SystemAdminHandler, jwtManager *jwt.JWTManager, repo repositories.SystemAdminRepository) {
	g := rg.Group("/sysadmin/auth")
	{
		g.POST("/login", h.Login)
		
		// Protected admin routes
		protected := g.Group("")
		protected.Use(middleware.SysadminAuthMiddleware(jwtManager, repo))
		{
			protected.GET("/me", h.Me)
			protected.POST("/logout", h.Logout)
		}
	}
}
