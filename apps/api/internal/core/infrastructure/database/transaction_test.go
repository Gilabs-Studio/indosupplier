package database

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB opens an in-memory SQLite database for use in unit tests.
// SQLite does not enforce tenant_id as a real column constraint, but it is
// sufficient to verify that GetDB appends the correct WHERE clause.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	return db
}

// stmtSQL extracts the SQL statement string from a prepared GORM statement
// without executing it, by using gorm.DB.ToSQL.
func toSQL(db *gorm.DB) string {
	return db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(nil)
	})
}

// TestGetDB_WithTenantID verifies that GetDB appends WHERE tenant_id = ?
// when the context carries a tenant_id value (i.e., authenticated tenant request).
func TestGetDB_WithTenantID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.WithValue(context.Background(), tenantIDKey, "tenant-abc-123")

	scoped := GetDB(ctx, db)

	// Build the SQL without executing it
	sql := scoped.ToSQL(func(tx *gorm.DB) *gorm.DB {
		type dummy struct{ ID string }
		return tx.Model(&dummy{}).Where("1=1").Find(&[]dummy{})
	})

	if sql == "" {
		t.Fatal("expected SQL statement but got empty string")
	}

	// The WHERE clause must include the tenant_id binding
	const want = "tenant_id"
	if !containsStr(sql, want) {
		t.Errorf("GetDB with tenant context: expected SQL to contain %q, got:\n%s", want, sql)
	}
}

// TestGetDB_WithSystemAdmin verifies that GetDB returns an unscoped DB
// when the context carries is_system_admin = true (platform-level access).
func TestGetDB_WithSystemAdmin(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.WithValue(context.Background(), isSystemAdminKey, true)
	// Also add a tenant_id to ensure system_admin flag wins
	ctx = context.WithValue(ctx, tenantIDKey, "tenant-should-be-ignored")

	scoped := GetDB(ctx, db)

	sql := scoped.ToSQL(func(tx *gorm.DB) *gorm.DB {
		type dummy struct{ ID string }
		return tx.Model(&dummy{}).Where("1=1").Find(&[]dummy{})
	})

	const forbidden = "tenant_id"
	if containsStr(sql, forbidden) {
		t.Errorf("GetDB with system_admin: expected SQL to NOT contain %q, got:\n%s", forbidden, sql)
	}
}

// TestGetDB_Unauthenticated verifies that GetDB returns an unscoped DB
// when the context has neither tenant_id nor is_system_admin (e.g. login endpoint).
func TestGetDB_Unauthenticated(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	scoped := GetDB(ctx, db)

	sql := scoped.ToSQL(func(tx *gorm.DB) *gorm.DB {
		type dummy struct{ ID string }
		return tx.Model(&dummy{}).Where("1=1").Find(&[]dummy{})
	})

	const forbidden = "tenant_id"
	if containsStr(sql, forbidden) {
		t.Errorf("GetDB unauthenticated: expected SQL to NOT contain %q, got:\n%s", forbidden, sql)
	}
}

// TestGetDB_TenantIsolation verifies that two different tenant contexts
// produce two different WHERE clauses (no data leak between tenants).
func TestGetDB_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)

	tenant1Ctx := context.WithValue(context.Background(), tenantIDKey, "tenant-111")
	tenant2Ctx := context.WithValue(context.Background(), tenantIDKey, "tenant-222")

	sql1 := GetDB(tenant1Ctx, db).ToSQL(func(tx *gorm.DB) *gorm.DB {
		type dummy struct{ ID string }
		return tx.Model(&dummy{}).Find(&[]dummy{})
	})
	sql2 := GetDB(tenant2Ctx, db).ToSQL(func(tx *gorm.DB) *gorm.DB {
		type dummy struct{ ID string }
		return tx.Model(&dummy{}).Find(&[]dummy{})
	})

	if sql1 == sql2 {
		t.Errorf("GetDB isolation: tenant1 and tenant2 queries must differ, both got:\n%s", sql1)
	}
	if !containsStr(sql1, "tenant-111") {
		t.Errorf("tenant1 SQL missing expected tenant-111:\n%s", sql1)
	}
	if !containsStr(sql2, "tenant-222") {
		t.Errorf("tenant2 SQL missing expected tenant-222:\n%s", sql2)
	}
}

// containsStr is a lightweight substring check used in SQL assertions.
func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
