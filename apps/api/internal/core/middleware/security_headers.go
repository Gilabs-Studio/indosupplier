package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const strictAPICSP = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'"
const posReceiptCSP = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; img-src data: https://api.qrserver.com"

func isPOSReceiptPath(path string) bool {
	trimmed := strings.TrimSuffix(strings.TrimSpace(path), "/")
	return strings.HasPrefix(trimmed, "/api/v1/pos/orders/") && strings.HasSuffix(trimmed, "/receipt")
}

// SecurityHeadersMiddleware adds baseline security headers.
// Safe for APIs and helps prevent common browser-based attacks.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		csp := strictAPICSP
		if isPOSReceiptPath(c.Request.URL.Path) {
			csp = posReceiptCSP
		}
		c.Header("Content-Security-Policy", csp)
		c.Next()
	}
}
