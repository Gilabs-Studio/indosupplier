package middleware

import (
	"fmt"
	"time"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	tenantRepos "github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const entitlementCacheTTL = 5 * time.Minute

// RequireModule returns a middleware that blocks the request with HTTP 403 when the
// tenant's active subscription plan does not include the given moduleSlug.
//
// Results are cached in Redis for entitlementCacheTTL to avoid per-request DB lookups.
// Cache key format: entitlement:{tenantID}:{moduleSlug}
func RequireModule(moduleSlug string, planRepo tenantRepos.SubscriptionPlanRepository, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString(TenantIDKey)
		if tenantID == "" {
			coreErrors.UnauthorizedResponse(c, "missing tenant context")
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		cacheKey := fmt.Sprintf("entitlement:%s:%s", tenantID, moduleSlug)

		// Check Redis cache first.
		if redisClient != nil {
			cached, err := redisClient.Get(ctx, cacheKey).Result()
			if err == nil {
				if cached == "1" {
					c.Next()
					return
				}
				coreErrors.ErrorResponse(c, "MODULE_NOT_ENTITLED", map[string]interface{}{
					"module": moduleSlug,
					"hint":   fmt.Sprintf("Your subscription plan does not include the '%s' module.", moduleSlug),
				}, nil)
				c.Abort()
				return
			}
		}

		// Load enabled modules for this tenant's active plan from DB.
		modules, err := planRepo.GetEnabledModulesForTenant(ctx, tenantID)
		if err != nil {
			// On DB error, allow access to avoid blocking legitimate users.
			c.Next()
			return
		}

		entitled := false
		for _, m := range modules {
			if m == moduleSlug {
				entitled = true
				break
			}
		}

		// Cache the result.
		if redisClient != nil {
			val := "0"
			if entitled {
				val = "1"
			}
			redisClient.Set(ctx, cacheKey, val, entitlementCacheTTL)
		}

		if !entitled {
			coreErrors.ErrorResponse(c, "MODULE_NOT_ENTITLED", map[string]interface{}{
				"module": moduleSlug,
				"hint":   fmt.Sprintf("Your subscription plan does not include the '%s' module.", moduleSlug),
			}, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
