package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/redis"
	"gorm.io/gorm"
)

type PermissionService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
	InvalidateCache(roleCode string) error
}

type cachedPermissionService struct {
	db           *gorm.DB
	l1Cache      sync.Map
	l1ScopeCache sync.Map // Separate L1 cache for scope-aware permissions
	l1TTL        time.Duration
	l2TTL        time.Duration
	redisTimeout time.Duration
}

type l1CacheItem struct {
	permissions []string
	expiresAt   time.Time
}

type l1ScopeCacheItem struct {
	permissions map[string]string // code -> scope
	expiresAt   time.Time
}

func NewPermissionService(db *gorm.DB) PermissionService {
	return &cachedPermissionService{
		db:           db,
		l1TTL:        1 * time.Minute,
		l2TTL:        1 * time.Hour,
		redisTimeout: 2 * time.Second,
	}
}

func (s *cachedPermissionService) redisContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.redisTimeout)
}

func (s *cachedPermissionService) GetPermissions(roleCode string) ([]string, error) {
	return s.GetPermissionsForTenant(roleCode, "")
}

// GetPermissionsForTenant loads permission codes for a role code within a tenant boundary.
// When tenantID is empty, it falls back to role-code-only behavior for system/global contexts.
func (s *cachedPermissionService) GetPermissionsForTenant(roleCode string, tenantID string) ([]string, error) {
	scopeKey := roleCode
	cacheKey := fmt.Sprintf("permissions:%s", roleCode)
	if tenantID != "" {
		scopeKey = fmt.Sprintf("%s::%s", tenantID, roleCode)
		cacheKey = fmt.Sprintf("permissions:%s:%s", tenantID, roleCode)
	}

	// 1. Check L1 Cache (Memory)
	if item, ok := s.l1Cache.Load(scopeKey); ok {
		cached := item.(l1CacheItem)
		if apptime.Now().Before(cached.expiresAt) {
			return cached.permissions, nil
		}
		s.l1Cache.Delete(scopeKey)
	}

	// 2. Check L2 Cache (Redis)
	redisClient := redis.GetClient()
	if redisClient != nil {
		ctx, cancel := s.redisContext()
		val, err := redisClient.Get(ctx, cacheKey).Result()
		cancel()
		if err == nil {
			var perms []string
			if err := json.Unmarshal([]byte(val), &perms); err == nil {
				s.l1Cache.Store(scopeKey, l1CacheItem{
					permissions: perms,
					expiresAt:   apptime.Now().Add(s.l1TTL),
				})
				return perms, nil
			}
		}
	}

	// 3. Fetch from DB
	var perms []string
	query := `
		SELECT p.code
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN roles r ON r.id = rp.role_id
		WHERE r.code = ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL
	`
	args := []interface{}{roleCode}
	if tenantID != "" {
		query += " AND r.tenant_id = ?"
		args = append(args, tenantID)
	}
	if err := s.db.Raw(query, args...).Scan(&perms).Error; err != nil {
		return nil, err
	}

	// 4. Update Caches
	if redisClient != nil {
		data, _ := json.Marshal(perms)
		ctx, cancel := s.redisContext()
		redisClient.Set(ctx, cacheKey, data, s.l2TTL)
		cancel()
	}

	s.l1Cache.Store(scopeKey, l1CacheItem{
		permissions: perms,
		expiresAt:   apptime.Now().Add(s.l1TTL),
	})

	return perms, nil
}

// GetPermissionsWithScope returns permission codes mapped to their scope for a role
func (s *cachedPermissionService) GetPermissionsWithScope(roleCode string) (map[string]string, error) {
	return s.GetPermissionsWithScopeForTenant(roleCode, "")
}

// GetPermissionsWithScopeForTenant loads scoped permissions for a role code within tenant boundary.
func (s *cachedPermissionService) GetPermissionsWithScopeForTenant(roleCode string, tenantID string) (map[string]string, error) {
	scopeKey := roleCode
	scopeCacheKey := fmt.Sprintf("permissions_scope:%s", roleCode)
	if tenantID != "" {
		scopeKey = fmt.Sprintf("%s::%s", tenantID, roleCode)
		scopeCacheKey = fmt.Sprintf("permissions_scope:%s:%s", tenantID, roleCode)
	}

	// 1. Check L1 Scope Cache (Memory)
	if item, ok := s.l1ScopeCache.Load(scopeKey); ok {
		cached := item.(l1ScopeCacheItem)
		if apptime.Now().Before(cached.expiresAt) {
			return cached.permissions, nil
		}
		s.l1ScopeCache.Delete(scopeKey)
	}

	// 2. Check L2 Cache (Redis)
	redisClient := redis.GetClient()
	if redisClient != nil {
		ctx, cancel := s.redisContext()
		val, err := redisClient.Get(ctx, scopeCacheKey).Result()
		cancel()
		if err == nil {
			var perms map[string]string
			if err := json.Unmarshal([]byte(val), &perms); err == nil {
				s.l1ScopeCache.Store(scopeKey, l1ScopeCacheItem{
					permissions: perms,
					expiresAt:   apptime.Now().Add(s.l1TTL),
				})
				return perms, nil
			}
		}
	}

	// 3. Fetch from DB with scope
	type permRow struct {
		Code  string
		Scope string
	}
	var rows []permRow
	query := `
		SELECT p.code, COALESCE(rp.scope, 'ALL') as scope
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN roles r ON r.id = rp.role_id
		WHERE r.code = ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL
	`
	args := []interface{}{roleCode}
	if tenantID != "" {
		query += " AND r.tenant_id = ?"
		args = append(args, tenantID)
	}
	if err := s.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		log.Printf("[PermissionService] GetPermissionsWithScope DB error for role '%s' tenant '%s': %v", roleCode, tenantID, err)

		fallbackPerms, fallbackErr := s.GetPermissionsForTenant(roleCode, tenantID)
		if fallbackErr != nil {
			log.Printf("[PermissionService] Fallback GetPermissions also failed for role '%s' tenant '%s': %v", roleCode, tenantID, fallbackErr)
			return nil, err
		}

		perms := make(map[string]string, len(fallbackPerms))
		for _, code := range fallbackPerms {
			perms[code] = "ALL"
		}
		return perms, nil
	}

	perms := make(map[string]string, len(rows))
	for _, row := range rows {
		scope := row.Scope
		if scope == "" {
			scope = "ALL"
		}
		perms[row.Code] = scope
	}

	// 4. Update Caches
	if redisClient != nil {
		data, _ := json.Marshal(perms)
		ctx, cancel := s.redisContext()
		redisClient.Set(ctx, scopeCacheKey, data, s.l2TTL)
		cancel()
	}

	s.l1ScopeCache.Store(scopeKey, l1ScopeCacheItem{
		permissions: perms,
		expiresAt:   apptime.Now().Add(s.l1TTL),
	})

	return perms, nil
}

func (s *cachedPermissionService) InvalidateCache(roleCode string) error {
	// Clear L1 (both caches), including tenant-scoped entries stored as
	// "tenant-id::roleCode". This keeps permission changes visible immediately
	// after role updates in multi-tenant sessions.
	s.l1Cache.Range(func(key, _ any) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}
		if keyStr == roleCode || strings.HasSuffix(keyStr, "::"+roleCode) {
			s.l1Cache.Delete(keyStr)
		}
		return true
	})
	s.l1ScopeCache.Range(func(key, _ any) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}
		if keyStr == roleCode || strings.HasSuffix(keyStr, "::"+roleCode) {
			s.l1ScopeCache.Delete(keyStr)
		}
		return true
	})

	// Clear L2
	redisClient := redis.GetClient()
	if redisClient != nil {
		cacheKey := fmt.Sprintf("permissions:%s", roleCode)
		scopeCacheKey := fmt.Sprintf("permissions_scope:%s", roleCode)
		ctx, cancel := s.redisContext()
		defer cancel()
		if err := redisClient.Del(ctx, cacheKey, scopeCacheKey).Err(); err != nil {
			return err
		}

		for _, pattern := range []string{
			fmt.Sprintf("permissions:*:%s", roleCode),
			fmt.Sprintf("permissions_scope:*:%s", roleCode),
		} {
			iter := redisClient.Scan(ctx, 0, pattern, 0).Iterator()
			for iter.Next(ctx) {
				if err := redisClient.Del(ctx, iter.Val()).Err(); err != nil {
					return err
				}
			}
			if err := iter.Err(); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}
