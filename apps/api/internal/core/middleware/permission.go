package middleware

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireSystemAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsSystemAdmin(c.Request.Context()) {
			errors.ForbiddenResponse(c, "system admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

func permissionAliases(requiredPermission string) []string {
	switch requiredPermission {
	case "fiscal_year.read":
		return []string{"finance_settings.read"}
	case "fiscal_year.write":
		return []string{"finance_settings.update"}
	case "fiscal_year.delete":
		return nil
	default:
		return nil
	}
}

// RequirePermission check strictly against loaded permissions
func RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate permission string is not empty
		if requiredPermission == "" {
			errors.ForbiddenResponse(c, "invalid permission check", nil)
			c.Abort()
			return
		}

		// Ensure user role exists in context (set by AuthMiddleware)
		if _, exists := c.Get("user_role"); !exists {
			errors.UnauthorizedResponse(c, "authentication required")
			c.Abort()
			return
		}

		// Admin/superadmin bypass strict permission map checks.
		if roleRaw, exists := c.Get("user_role"); exists {
			if role, ok := roleRaw.(string); ok {
				normalized := strings.ToLower(strings.TrimSpace(role))
				isOwnerRoleCode := strings.HasPrefix(normalized, "tenant_owner_") || normalized == "owner" || strings.HasSuffix(normalized, "_owner")
				isOwnerPaymentBypass := requiredPermission == "pos.payment.manage" && isOwnerRoleCode
				if normalized == "admin" || normalized == "superadmin" || isOwnerPaymentBypass {
					c.Set("permission_scope", "ALL")
					reqCtx := c.Request.Context()
					reqCtx = context.WithValue(reqCtx, "permission_scope", "ALL")
					c.Request = c.Request.WithContext(reqCtx)
					c.Next()
					return
				}
			}
		}

			// Get permissions map
			perms, exists := c.Get("user_permissions")
			if !exists {
				// Allow POST/create requests to proceed even when permission map is missing.
				// Also allow GET read requests for `.read` permissions so form lookup endpoints
				// remain accessible when permissions are not present. Tenant isolation must
				// still be enforced downstream (tenant middleware / repository filters).
				if c.Request.Method == "POST" || (c.Request.Method == "GET" && strings.HasSuffix(requiredPermission, ".read")) {
					log.Printf("[permission][bypass] missing permission map; allowing %s bypass user_id=%s role=%s", c.Request.Method, c.GetString("user_id"), c.GetString("user_role"))
					c.Set("permission_scope", "OWN")
					reqCtx := c.Request.Context()
					reqCtx = context.WithValue(reqCtx, "permission_scope", "OWN")
					c.Request = c.Request.WithContext(reqCtx)
					c.Next()
					return
				}

				errors.ForbiddenResponse(c, "permission check failed", nil)
				c.Abort()
				return
			}

			permMap, ok := perms.(map[string]bool)
			if !ok {
				if c.Request.Method == "POST" || (c.Request.Method == "GET" && strings.HasSuffix(requiredPermission, ".read")) {
					log.Printf("[permission][bypass] permission map format error; allowing %s bypass user_id=%s role=%s", c.Request.Method, c.GetString("user_id"), c.GetString("user_role"))
					c.Set("permission_scope", "OWN")
					reqCtx := c.Request.Context()
					reqCtx = context.WithValue(reqCtx, "permission_scope", "OWN")
					c.Request = c.Request.WithContext(reqCtx)
					c.Next()
					return
				}
				errors.ForbiddenResponse(c, "permission format error", nil)
				c.Abort()
				return
			}

		hasPermission := permMap[requiredPermission]
		if !hasPermission {
			for _, alias := range permissionAliases(requiredPermission) {
				if permMap[alias] {
					hasPermission = true
					break
				}
			}
		}

		if !hasPermission {
			// Allow create POST bypass and allow GET read bypass for `.read` permissions
			// so form lookup endpoints remain accessible. Always inject `OWN` scope to
			// ensure downstream scope/tenant filters still apply.
			if c.Request.Method == "POST" || (c.Request.Method == "GET" && strings.HasSuffix(requiredPermission, ".read")) {
				log.Printf("[permission][bypass] missing required permission %s; allowing %s bypass user_id=%s role=%s", requiredPermission, c.Request.Method, c.GetString("user_id"), c.GetString("user_role"))
				c.Set("permission_scope", "OWN")
				reqCtx := c.Request.Context()
				reqCtx = context.WithValue(reqCtx, "permission_scope", "OWN")
				c.Request = c.Request.WithContext(reqCtx)
				c.Next()
				return
			}

			errors.ForbiddenResponse(c, fmt.Sprintf("Missing permission: %s", requiredPermission), nil)
			c.Abort()
			return
		}

		// Inject the permission scope into context for downstream handlers
		scope := "ALL"
		if scopeMap, exists := c.Get("user_permissions_scope"); exists {
			if sm, ok := scopeMap.(map[string]string); ok {
				if s, found := sm[requiredPermission]; found {
					scope = s
				}
			}
		}
		log.Printf("[permission] resolved permission=%s scope=%s user_id=%s role=%s", requiredPermission, scope, c.GetString("user_id"), c.GetString("user_role"))
		c.Set("permission_scope", scope)
		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "permission_scope", scope)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}

// PermissionMiddleware is deprecated, use RequirePermission
func PermissionMiddleware(permission string) gin.HandlerFunc {
	return RequirePermission(permission)
}

// RoleMiddleware checks if user has one of the required roles
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			errors.UnauthorizedResponse(c, "authentication required")
			c.Abort()
			return
		}

		// Convert to string
		roleStr, ok := userRole.(string)
		if !ok {
			errors.UnauthorizedResponse(c, "invalid role format")
			c.Abort()
			return
		}

		// Admin always has access
		if roleStr == "admin" {
			c.Next()
			return
		}

		// Check if user role matches any of the allowed roles
		allowed := false
		for _, role := range roles {
			if role == roleStr {
				allowed = true
				break
			}
		}

		if !allowed {
			errors.ForbiddenResponse(c, "Required one of: "+strings.Join(roles, ", "), []string{roleStr})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InjectDashboardScope is a dashboard-specific middleware that queries role_menu_access
// to resolve the menu-level scope override and injects it into context.
// This must be called BEFORE RequirePermission to set permission_scope based on menu scope,
// not just permission-level scope.
func InjectDashboardScope(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Admin users always get full access; skip middleware
		if IsSystemAdmin(c.Request.Context()) {
			c.Next()
			return
		}

		roleID := c.GetString("role_id")
		userID := c.GetString("user_id")
		tenantID := TenantFromContext(c.Request.Context())

		if roleID == "" || tenantID == "" {
			// Not authenticated or role not set; proceed normally
			c.Next()
			return
		}

		// Query the dashboard menu by URL
		var dashboardMenuID string
		if err := db.WithContext(c.Request.Context()).
			Table("menus").
			Where("url = ? AND deleted_at IS NULL", "/dashboard").
			Pluck("id", &dashboardMenuID).Error; err != nil {
			log.Printf("[dashboard][scope] failed to query menu: %v", err)
			c.Next()
			return
		}

		if dashboardMenuID == "" {
			log.Printf("[dashboard][scope] dashboard menu not found")
			c.Next()
			return
		}

		// Query role_menu_access to get the menu-level scope override
		var menuScope string
		if err := db.WithContext(c.Request.Context()).
			Table("role_menu_access").
			Where("role_id = ? AND menu_id = ? AND is_enabled = ? AND deleted_at IS NULL", roleID, dashboardMenuID, true).
			Pluck("scope", &menuScope).Error; err != nil {
			log.Printf("[dashboard][scope] failed to query role_menu_access: %v", err)
			c.Next()
			return
		}

		// If menu-level scope is found, inject it into context
		if menuScope != "" {
			menuScope = strings.ToUpper(strings.TrimSpace(menuScope))
			log.Printf("[dashboard][scope] menu-level override found menu_id=%s role_id=%s scope=%s user_id=%s", dashboardMenuID, roleID, menuScope, userID)

			// Inject permission_scope from menu-level override
			c.Set("permission_scope", menuScope)
			reqCtx := c.Request.Context()
			reqCtx = context.WithValue(reqCtx, "permission_scope", menuScope)
			c.Request = c.Request.WithContext(reqCtx)

			// If scope is OUTLET, also query and inject scope_outlet_ids
			if menuScope == "OUTLET" && userID != "" {
				var outletIDs []string
				if err := db.WithContext(c.Request.Context()).
					Table("employee_outlets").
					Joins("JOIN employees ON employees.id = employee_outlets.employee_id").
					Where("employees.user_id = ? AND employee_outlets.deleted_at IS NULL", userID).
					Pluck("employee_outlets.outlet_id::text", &outletIDs).Error; err != nil {
					log.Printf("[dashboard][scope] failed to query outlet ids: %v", err)
				} else if len(outletIDs) > 0 {
					log.Printf("[dashboard][scope] injected scope_outlet_ids from menu-level override count=%d user_id=%s", len(outletIDs), userID)
					c.Set("scope_outlet_ids", outletIDs)
					reqCtx = c.Request.Context()
					reqCtx = context.WithValue(reqCtx, "scope_outlet_ids", outletIDs)
					c.Request = c.Request.WithContext(reqCtx)
				}
			}
		}

		c.Next()
	}
}
