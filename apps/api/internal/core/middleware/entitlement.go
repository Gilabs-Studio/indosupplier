package middleware

import "github.com/gin-gonic/gin"

// EntitlementMiddleware is kept as a no-op during project reset.
func EntitlementMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
