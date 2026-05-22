package router

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	permissionHandler "github.com/gilabs/gims/api/internal/permission/presentation/handler"
	"github.com/gilabs/gims/api/internal/user/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup, h *handler.UserHandler, ph *permissionHandler.PermissionHandler, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	g := rg.Group("/users")
	g.Use(middleware.AuthMiddleware(jwtManager, permService))
	{
		// Static routes BEFORE parameterized /:id to avoid path conflicts
		g.GET("/available", middleware.RequirePermission("employee.read"), h.GetAvailable)
		g.GET("/limit", middleware.RequirePermission("user.read"), h.GetLimit)

		g.GET("", middleware.RequirePermission("user.read"), h.List)
		g.GET("/:id", middleware.RequirePermission("user.read"), h.GetByID)
		g.POST("", middleware.RequirePermission("user.create"), h.Create)
		g.PUT("/:id", middleware.RequirePermission("user.update"), h.Update)
		g.DELETE("/:id", middleware.RequirePermission("user.delete"), h.Delete)

		// Add permissions route
		g.GET("/:id/permissions", ph.GetUserPermissions)
	}

	// Profile routes - separate from /users CRUD to avoid conflict with /:id and clearer intent
	// Attach to the parent group (which is likely /api/v1)
	p := rg.Group("/profile")
	p.Use(middleware.AuthMiddleware(jwtManager, permService))
	{
		p.PUT("", h.UpdateProfile)
		p.PUT("/password", h.ChangePassword)
		p.POST("/avatar", h.UploadAvatar)
		p.POST("/delete-account", h.RequestAccountDeletion)
	}
}
