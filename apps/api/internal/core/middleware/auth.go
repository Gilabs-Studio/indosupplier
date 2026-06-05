package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
)

// AuthMiddleware validates JWT token and sets user context.
func AuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
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

		if claims.UserID == "" || claims.Email == "" {
			errors.ErrorResponse(c, "TOKEN_INVALID", nil, nil)
			c.Abort()
			return
		}
		if claims.SubjectType != jwt.TokenSubjectUser {
			errors.ErrorResponse(c, "TOKEN_INVALID", nil, nil)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "user_id", claims.UserID)
		reqCtx = context.WithValue(reqCtx, "user_email", claims.Email)
		reqCtx = context.WithValue(reqCtx, "client_ip", c.ClientIP())
		reqCtx = context.WithValue(reqCtx, "user_agent", c.Request.UserAgent())
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}
