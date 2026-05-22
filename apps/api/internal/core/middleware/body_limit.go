package middleware

import (
	"net/http"
	"strings"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gin-gonic/gin"
)

// BodySizeLimitMiddleware enforces maximum request body size.
//
// This protects the server from unbounded memory usage and oversized payload DoS.
// Limits are configured via config.AppConfig.Server.
func BodySizeLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.AppConfig == nil {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		isMultipart := strings.HasPrefix(strings.ToLower(contentType), "multipart/")

		maxBytes := config.AppConfig.Server.MaxBodyBytes
		if isMultipart {
			maxBytes = config.AppConfig.Server.MaxMultipartBodyBytes
		}

		if maxBytes > 0 && c.Request.ContentLength > maxBytes {
			coreErrors.ErrorResponse(c, "REQUEST_BODY_TOO_LARGE", nil, nil)
			c.Abort()
			return
		}

		if maxBytes > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}

		c.Next()
	}
}
