package middleware

import (
	"context"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gin-gonic/gin"
)

// Context keys for tenant isolation
const (
	TenantIDKey = "tenant_id"
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

// TenantGuard is a Gin middleware that ensures tenant_id is present in context.
// It must run AFTER the auth middleware which sets tenant_id.
func TenantGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
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
