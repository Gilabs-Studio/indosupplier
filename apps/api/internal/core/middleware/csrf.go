package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
	"github.com/gin-gonic/gin"
)

// generateToken creates a random 32-byte hex string
func generateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func shouldUseCrossSiteCSRFCookie(c *gin.Context) bool {
	if config.AppConfig != nil {
		env := strings.ToLower(strings.TrimSpace(config.AppConfig.Server.Env))
		if env == "production" || env == "prod" {
			return true
		}
	}

	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return false
	}

	requestHost := c.Request.Host
	if host, _, err := net.SplitHostPort(requestHost); err == nil {
		requestHost = host
	}

	originHost := originURL.Hostname()
	requestHostname := requestHost
	if requestHostname == "" {
		requestHostname = c.Request.URL.Hostname()
	}

	return !strings.EqualFold(originHost, requestHostname)
}

func abortCSRFFailedToGenerate(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "Failed to generate CSRF token",
		},
	})
}

func ensureCSRFCookie(c *gin.Context) (string, bool) {
	token, err := c.Cookie("indosupplier_csrf_token")
	if err == nil && token != "" {
		return token, true
	}

	token = generateToken()
	if token == "" {
		abortCSRFFailedToGenerate(c)
		return "", false
	}

	setCSRFCookie(c, token)
	return token, true
}

func isPublicFeedbackSubmit(method, path string) bool {
	return method == http.MethodPost && strings.HasPrefix(path, "/api/v1/public/feedback/") && strings.HasSuffix(path, "/submit")
}

func isPublicAuthEndpoint(path string) bool {
	// Public auth endpoints don't require CSRF validation because they're accessed
	// without an established browser session. Client-side CSRF isn't applicable until
	// after successful login when browser cookies are set.
	publicAuthPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/csrf",
		"/api/v1/auth/refresh-token",
	}
	for _, p := range publicAuthPaths {
		if path == p {
			return true
		}
	}
	return false
}

func isPublicPOSSelfOrderEndpoint(path string) bool {
	// Public POS self-order routes are authenticated by a one-time table token in the URL,
	// not by a browser session cookie, so CSRF validation does not add protection here.
	return strings.HasPrefix(path, "/api/v1/public/pos/tables/")
}

func shouldSkipCSRFMiddleware(c *gin.Context, path string) bool {
	if isCRMLeadUpsertWebhookRequest(c.Request.Method, path) {
		return true
	}

	// System admin login is pre-session and does not require CSRF.
	if strings.HasPrefix(path, "/internal/sys-login") {
		return true
	}

	// Public auth endpoints (registration, plans, csrf token fetch) bypass CSRF
	// since they're accessed before a session is established.
	if isPublicAuthEndpoint(path) {
		return true
	}

	if isPublicPOSSelfOrderEndpoint(path) {
		return true
	}

	// Public feedback submission uses one-time URL token, not browser session CSRF.
	return isPublicFeedbackSubmit(c.Request.Method, path)
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func isBearerAuthRequest(c *gin.Context) bool {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return false
	}

	return strings.TrimSpace(parts[1]) != ""
}

func validateCSRFTokens(c *gin.Context, token string) bool {
	requestToken := c.GetHeader("X-CSRF-Token")
	if requestToken != "" && requestToken == token {
		return true
	}

	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "CSRF_INVALID",
			"message": "Invalid or missing CSRF token",
		},
	})
	return false
}

// CSRF middleware implements the Double-Submit Cookie pattern
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := ensureCSRFCookie(c)
		if !ok {
			return
		}

		// ALWAYS expose the current token in the header so frontend can read it (cross-origin support)
		c.Header("X-CSRF-Token", token)

		path := c.Request.URL.Path
		if shouldSkipCSRFMiddleware(c, path) {
			c.Next()
			return
		}

		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// CSRF is only relevant for cookie-auth browser flows.
		// POS/mobile sends bearer token manually via Authorization header.
		if isBearerAuthRequest(c) {
			c.Next()
			return
		}

		if !validateCSRFTokens(c, token) {
			return
		}

		c.Next()
	}
}

func isHTTPS(c *gin.Context) bool {
	if config.AppConfig == nil || config.AppConfig.Server.Env != "production" {
		return false
	}
	if c.Request.TLS != nil {
		return true
	}
	if config.AppConfig.Security.ProxyHeadersEnabled {
		xfp := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")))
		return xfp == "https"
	}
	return false
}

// setCSRFCookie sets the non-HttpOnly cookie so frontend can read it
func setCSRFCookie(c *gin.Context, token string) {
	sameSite := http.SameSiteLaxMode
	domain := ""

	if shouldUseCrossSiteCSRFCookie(c) {
		sameSite = http.SameSiteNoneMode
		// In production, set root domain so cookies work across all subdomains (e.g., .gilabs.id, .indosupplier.id)
		// This allows sharing cookies between indosupplier.id, indosupplier.gilabs.id, api.gilabs.id, etc.
		if config.AppConfig != nil && strings.ToLower(strings.TrimSpace(config.AppConfig.Server.Env)) == "production" {
			domain = config.AppConfig.Server.RootDomain
		}
	}

	isSecure := isHTTPS(c)
	if config.AppConfig != nil && strings.ToLower(strings.TrimSpace(config.AppConfig.Server.Env)) == "production" {
		isSecure = true
	}
	if sameSite == http.SameSiteNoneMode {
		isSecure = true
	}

	c.SetSameSite(sameSite)

	// Note: HttpOnly is FALSE so JavaScript can read it and send in header (Double-Submit Cookie pattern)
	c.SetCookie("indosupplier_csrf_token", token, 3600*24, "/", domain, isSecure, false)
}
