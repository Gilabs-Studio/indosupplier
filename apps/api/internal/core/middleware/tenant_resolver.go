package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// userTenantCache caches user_id -> tenant_id to avoid a DB query on every request.
// Cache entries expire after 5 minutes.
type userTenantCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	tenantID  string
	expiresAt time.Time
}

var tenantCache = &userTenantCache{
	entries: make(map[string]cacheEntry),
}

func (c *userTenantCache) get(userID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[userID]
	if !ok || time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.tenantID, true
}

func (c *userTenantCache) set(userID, tenantID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[userID] = cacheEntry{
		tenantID:  tenantID,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
}

// TenantResolverMiddleware resolves tenant_id for the authenticated user
// and injects it into both gin.Context and request context.
// Must run AFTER AuthMiddleware.
func TenantResolverMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if system admin
		if isAdmin, exists := c.Get(IsSystemAdminKey); exists {
			if b, ok := isAdmin.(bool); ok && b {
				c.Next()
				return
			}
		}

		userID, exists := c.Get("user_id")
		if !exists {
			// No auth context — skip (public endpoint or webhook)
			c.Next()
			return
		}

		uid, ok := userID.(string)
		if !ok || uid == "" {
			c.Next()
			return
		}

		// Check cache first
		if tid, found := tenantCache.get(uid); found {
			setTenantContext(c, tid)
			c.Next()
			return
		}

		// Query tenant_id from users table
		var tenantID string
		err := db.Table("users").
			Select("tenant_id").
			Where("id = ? AND deleted_at IS NULL", uid).
			Row().Scan(&tenantID)
		if err != nil || tenantID == "" {
			// User might not have tenant_id yet (migration pending)
			// Gracefully continue without tenant scope
			c.Next()
			return
		}

		tenantCache.set(uid, tenantID)
		setTenantContext(c, tenantID)
		c.Next()
	}
}

func setTenantContext(c *gin.Context, tenantID string) {
	c.Set(TenantIDKey, tenantID)

	reqCtx := c.Request.Context()
	reqCtx = context.WithValue(reqCtx, TenantIDKey, tenantID)
	c.Request = c.Request.WithContext(reqCtx)
}
