package middleware

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gin-gonic/gin"
)

// Context keys for tenant isolation
const (
	TenantIDKey      = "tenant_id"
	IsSystemAdminKey = "is_system_admin"
)

// TenantFromContext extracts the tenant_id from a context.Context.
func TenantFromContext(ctx context.Context) string {
	if v := ctx.Value(TenantIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// IsSystemAdmin returns true when the request comes from a system admin session.
func IsSystemAdmin(ctx context.Context) bool {
	if v := ctx.Value(IsSystemAdminKey); v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// TenantGuard is a Gin middleware that ensures tenant_id is present in context.
// It must run AFTER the auth middleware which sets tenant_id.
// System admin requests bypass this check.
func TenantGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// System admins bypass tenant isolation
		if isAdmin, exists := c.Get(IsSystemAdminKey); exists {
			if b, ok := isAdmin.(bool); ok && b {
				c.Next()
				return
			}
		}

		tenantID, exists := c.Get(TenantIDKey)
		if !exists || tenantID == "" {
			errors.ErrorResponse(c, "FORBIDDEN", map[string]interface{}{
				"reason": "tenant context is required",
			}, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
