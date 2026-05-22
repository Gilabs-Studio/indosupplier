package router

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gin-gonic/gin"
)

// NewEngine creates a new Gin engine with global middlewares
func NewEngine(jwtManager *jwt.JWTManager) *gin.Engine {
	// Set Gin mode
	if config.AppConfig.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Reverse-proxy support: trust X-Forwarded-* ONLY when explicitly enabled.
	if config.AppConfig != nil && config.AppConfig.Security.ProxyHeadersEnabled {
		trusted := config.AppConfig.Security.TrustedProxies
		if len(trusted) > 0 {
			_ = r.SetTrustedProxies(trusted)
		}
	}

	// Multipart parsing memory limit (files spill to disk beyond this limit)
	if config.AppConfig != nil && config.AppConfig.Server.MaxMultipartMemoryBytes > 0 {
		r.MaxMultipartMemory = config.AppConfig.Server.MaxMultipartMemoryBytes
	}

	// Global middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.BodySizeLimitMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.HSTSMiddleware())
	r.Use(middleware.RateLimitMiddleware("general")) // Default rate limit
	if config.AppConfig != nil && config.AppConfig.Security.CSRFEnabled {
		r.Use(middleware.CSRF()) // Add CSRF protection
	}

	return r
}
