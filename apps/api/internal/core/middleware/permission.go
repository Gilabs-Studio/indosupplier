package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gin-gonic/gin"
)

// RequirePermission checks user_permissions map populated by AuthMiddleware.
func RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requiredPermission == "" {
			errors.ForbiddenResponse(c, "invalid permission check", nil)
			c.Abort()
			return
		}

		if _, exists := c.Get("user_role"); !exists {
			errors.UnauthorizedResponse(c, "authentication required")
			c.Abort()
			return
		}

		if roleRaw, exists := c.Get("user_role"); exists {
			if role, ok := roleRaw.(string); ok && strings.EqualFold(strings.TrimSpace(role), "admin") {
				c.Set("permission_scope", "ALL")
				reqCtx := c.Request.Context()
				reqCtx = context.WithValue(reqCtx, "permission_scope", "ALL")
				c.Request = c.Request.WithContext(reqCtx)
				c.Next()
				return
			}
		}

		perms, exists := c.Get("user_permissions")
		if !exists {
			errors.ForbiddenResponse(c, "permission check failed", nil)
			c.Abort()
			return
		}

		permMap, ok := perms.(map[string]bool)
		if !ok {
			errors.ForbiddenResponse(c, "permission format error", nil)
			c.Abort()
			return
		}

		if !permMap[requiredPermission] {
			errors.ForbiddenResponse(c, fmt.Sprintf("Missing permission: %s", requiredPermission), nil)
			c.Abort()
			return
		}

		scope := "ALL"
		if scopeMap, exists := c.Get("user_permissions_scope"); exists {
			if sm, ok := scopeMap.(map[string]string); ok {
				if s, found := sm[requiredPermission]; found && s != "" {
					scope = s
				}
			}
		}

		c.Set("permission_scope", scope)
		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "permission_scope", scope)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}

// PermissionMiddleware is deprecated, use RequirePermission.
func PermissionMiddleware(permission string) gin.HandlerFunc {
	return RequirePermission(permission)
}

// RoleMiddleware checks if user has one of the required roles.
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			errors.UnauthorizedResponse(c, "authentication required")
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			errors.UnauthorizedResponse(c, "invalid role format")
			c.Abort()
			return
		}

		if strings.EqualFold(roleStr, "admin") {
			c.Next()
			return
		}

		for _, role := range roles {
			if role == roleStr {
				c.Next()
				return
			}
		}

		errors.ForbiddenResponse(c, "Required one of: "+strings.Join(roles, ", "), []string{roleStr})
		c.Abort()
	}
}
