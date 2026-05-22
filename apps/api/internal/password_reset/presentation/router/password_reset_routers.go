package router

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/password_reset/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPasswordResetRoutes registers password reset routes
func RegisterPasswordResetRoutes(rg *gin.RouterGroup, h *handler.PasswordResetHandler, jwtManager *jwt.JWTManager, permissionService security.PermissionService) {
	g := rg.Group("/password-reset")
	{
		// Public routes (no authentication required)
		g.POST("/forgot-password", middleware.RateLimitMiddleware("password_reset"), h.ForgotPassword)
		g.POST("/reset-password", middleware.RateLimitMiddleware("password_reset"), h.ResetPassword)
		g.GET("/validate-token", middleware.RateLimitMiddleware("password_reset"), h.ValidateResetToken)
	}
}
