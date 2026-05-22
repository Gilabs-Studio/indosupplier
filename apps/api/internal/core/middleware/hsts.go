package middleware

import (
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gin-gonic/gin"
)

// HSTSMiddleware sets HTTP Strict Transport Security headers
// Only applies to HTTPS connections to prevent MITM attacks
func HSTSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isHTTPS := c.Request.TLS != nil
		if !isHTTPS && config.AppConfig != nil && config.AppConfig.Security.ProxyHeadersEnabled {
			xfp := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")))
			isHTTPS = xfp == "https"
		}

		// Only set HSTS header if request is HTTPS (direct or trusted proxy)
		if isHTTPS {
			hstsConfig := config.AppConfig.HSTS

			// Build HSTS header value
			headerValue := fmt.Sprintf("max-age=%d", hstsConfig.MaxAge)

			if hstsConfig.IncludeSubDomains {
				headerValue += "; includeSubDomains"
			}

			if hstsConfig.Preload {
				headerValue += "; preload"
			}

			c.Header("Strict-Transport-Security", headerValue)
		}

		c.Next()
	}
}
