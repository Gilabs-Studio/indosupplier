package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/repositories"
)

func SysadminAuthMiddleware(jwtManager *jwt.JWTManager, repo repositories.SystemAdminRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			cookie, err := c.Cookie("indosupplier_admin_token")
			if err == nil && cookie != "" {
				tokenString = cookie
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

		if claims.UserID == "" || claims.Email == "" {
			errors.ErrorResponse(c, "TOKEN_INVALID", nil, nil)
			c.Abort()
			return
		}
		if claims.SubjectType != jwt.TokenSubjectSystemAdmin {
			errors.ErrorResponse(c, "TOKEN_INVALID", nil, nil)
			c.Abort()
			return
		}

		// Verify admin exists in the database
		admin, err := repo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil || admin.Status != "active" {
			errors.ForbiddenResponse(c, "admin_access_required", nil)
			c.Abort()
			return
		}

		c.Set("admin_id", admin.ID)
		c.Set("admin_email", admin.Email)
		c.Set("admin_permission_set", admin.PermissionSet)

		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "admin_id", admin.ID)
		reqCtx = context.WithValue(reqCtx, "admin_email", admin.Email)
		reqCtx = context.WithValue(reqCtx, "admin_permission_set", admin.PermissionSet)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}
