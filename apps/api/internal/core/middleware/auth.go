package middleware

import (
	"context"
	"strings"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token and sets user context.
func AuthMiddleware(jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) gin.HandlerFunc {
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
			cookie, err := c.Cookie("indosupplier_access_token")
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

		if claims.UserID == "" || claims.Email == "" || claims.Role == "" {
			errors.ErrorResponse(c, "TOKEN_INVALID", nil, nil)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "user_id", claims.UserID)
		reqCtx = context.WithValue(reqCtx, "user_email", claims.Email)
		reqCtx = context.WithValue(reqCtx, "user_role", claims.Role)
		reqCtx = context.WithValue(reqCtx, "client_ip", c.ClientIP())
		reqCtx = context.WithValue(reqCtx, "user_agent", c.Request.UserAgent())

		permScopeMap, err := permService.GetPermissionsWithScope(claims.Role)
		if err != nil {
			plainPerms, loadErr := permService.GetPermissions(claims.Role)
			if loadErr != nil {
				errors.ErrorResponse(c, "FORBIDDEN", map[string]interface{}{"reason": "unable to load user permissions"}, nil)
				c.Abort()
				return
			}
			permScopeMap = make(map[string]string, len(plainPerms))
			for _, code := range plainPerms {
				permScopeMap[code] = "ALL"
			}
		}

		permMap := make(map[string]bool, len(permScopeMap))
		for code := range permScopeMap {
			permMap[code] = true
		}
		c.Set("user_permissions", permMap)
		c.Set("user_permissions_scope", permScopeMap)

		reqCtx = context.WithValue(reqCtx, "user_permissions", permMap)
		reqCtx = context.WithValue(reqCtx, "user_permissions_scope", permScopeMap)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}
