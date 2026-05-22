package router

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/role/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(rg *gin.RouterGroup, h *handler.RoleHandler, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	g := rg.Group("/roles")
	g.Use(middleware.AuthMiddleware(jwtManager, permService))
	{
		g.GET("", middleware.RequirePermission("role.read"), h.List)
		g.GET("/:id", middleware.RequirePermission("role.read"), h.GetByID)
		g.POST("", middleware.RequirePermission("role.create"), h.Create)
		g.PUT("/:id", middleware.RequirePermission("role.update"), h.Update)
		g.DELETE("/:id", middleware.RequirePermission("role.delete"), h.Delete)
		g.POST("/:id/permissions", middleware.RequirePermission("role.assign_permissions"), h.AssignPermissions)
		g.GET("/:id/menu-access", middleware.RequirePermission("role.read"), h.GetMenuAccess)
		g.PUT("/:id/menu-access", middleware.RequirePermission("role.assign_permissions"), h.UpdateMenuAccess)
		g.GET("/validate", h.ValidateUserRole)
	}
}
