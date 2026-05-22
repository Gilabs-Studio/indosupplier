package router

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/permission/presentation/handler"
	"github.com/gin-gonic/gin"
)

const permissionRead = "permission.read"

func RegisterPermissionRoutes(rg *gin.RouterGroup, h *handler.PermissionHandler, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	g := rg.Group("/permissions")
	g.Use(middleware.AuthMiddleware(jwtManager, permService))
	{
		g.GET("", middleware.RequirePermission(permissionRead), h.List)
		g.GET("/:id", middleware.RequirePermission(permissionRead), h.GetByID)
		g.GET("/categories/hierarchy", middleware.RequirePermission(permissionRead), h.GetMenuCategories)
	}
}
