package middleware

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gin-gonic/gin"
)

// SystemAdminAuthMiddleware validates JWT and ensures the caller is a system admin.
// System admin tokens have role="system_admin" which is never assigned to tenant users.
func SystemAdminAuthMiddleware(jwtManager *jwt.JWTManager, sysAdminRepo repositories.SystemAdminRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Check cookie first
		if cookie, err := c.Cookie("gims_sys_access_token"); err == nil && cookie != "" {
			tokenString = cookie
		}

		// Fallback to Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			}
		}

		if tokenString == "" {
			errors.UnauthorizedResponse(c, "token missing")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				errors.ErrorResponse(c, "TOKEN_EXPIRED", nil, nil)
			} else {
				errors.ErrorResponse(c, "TOKEN_INVALID", nil, nil)
			}
			c.Abort()
			return
		}

		// Enforce system_admin role
		if claims.Role != "system_admin" {
			errors.ErrorResponse(c, "FORBIDDEN", map[string]interface{}{
				"reason": "system admin access required",
			}, nil)
			c.Abort()
			return
		}

		// Verify admin still exists and is active
		admin, err := sysAdminRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil || admin.Status != "active" {
			errors.ErrorResponse(c, "ACCOUNT_DISABLED", nil, nil)
			c.Abort()
			return
		}

		// Set system admin context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", admin.Email)
		c.Set("user_name", admin.Name)
		c.Set("user_role", "system_admin")
		c.Set(IsSystemAdminKey, true)

		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "user_id", claims.UserID)
		reqCtx = context.WithValue(reqCtx, "user_email", admin.Email)
		reqCtx = context.WithValue(reqCtx, "user_name", admin.Name)
		reqCtx = context.WithValue(reqCtx, "user_role", "system_admin")
		reqCtx = context.WithValue(reqCtx, IsSystemAdminKey, true)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}
