package security

import "testing"

func TestInvalidateCacheRemovesTenantScopedEntries(t *testing.T) {
	svc := &cachedPermissionService{}

	svc.l1Cache.Store("role-user", l1CacheItem{})
	svc.l1Cache.Store("tenant-1::role-user", l1CacheItem{})
	svc.l1Cache.Store("tenant-2::role-admin", l1CacheItem{})
	svc.l1ScopeCache.Store("role-user", l1ScopeCacheItem{})
	svc.l1ScopeCache.Store("tenant-1::role-user", l1ScopeCacheItem{})
	svc.l1ScopeCache.Store("tenant-2::role-admin", l1ScopeCacheItem{})

	if err := svc.InvalidateCache("role-user"); err != nil {
		t.Fatalf("InvalidateCache returned error: %v", err)
	}

	if _, ok := svc.l1Cache.Load("role-user"); ok {
		t.Fatal("expected global l1 cache entry to be removed")
	}
	if _, ok := svc.l1Cache.Load("tenant-1::role-user"); ok {
		t.Fatal("expected tenant-scoped l1 cache entry to be removed")
	}
	if _, ok := svc.l1ScopeCache.Load("role-user"); ok {
		t.Fatal("expected global scoped l1 cache entry to be removed")
	}
	if _, ok := svc.l1ScopeCache.Load("tenant-1::role-user"); ok {
		t.Fatal("expected tenant-scoped scoped l1 cache entry to be removed")
	}

	if _, ok := svc.l1Cache.Load("tenant-2::role-admin"); !ok {
		t.Fatal("expected unrelated l1 cache entry to remain")
	}
	if _, ok := svc.l1ScopeCache.Load("tenant-2::role-admin"); !ok {
		t.Fatal("expected unrelated scoped l1 cache entry to remain")
	}
}
