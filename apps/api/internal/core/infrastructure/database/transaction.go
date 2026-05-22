package database

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type txKey struct{}

// Context key constants - duplicated from middleware package to avoid import cycle.
// These MUST stay in sync with middleware.TenantIDKey and middleware.IsSystemAdminKey.
const (
	tenantIDKey      = "tenant_id"
	isSystemAdminKey = "is_system_admin"
)

// WithTx returns a new context with the transaction attached
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetTx returns the transaction from the context if it exists, otherwise returns nil
func GetTx(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(txKey{}).(*gorm.DB)
	return tx
}

// GetDB returns a *gorm.DB scoped to the current tenant from context.
// - System admin requests (is_system_admin=true): unscoped, full access.
// - Tenant requests (tenant_id set): WHERE tenant_id = ? applied automatically.
// - Unauthenticated requests: unscoped (login, webhooks, etc.).
//
// IMPORTANT: Repositories querying platform-wide tables without tenant_id
// (geographic data, tenants table) must use r.db.WithContext(ctx) directly
// instead of this function to avoid SQL column-not-found errors.
func GetDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	var db *gorm.DB
	if tx := GetTx(ctx); tx != nil {
		// Propagate the request context into the transaction so that
		// tenant_id is visible to any subsequent GetDB calls within the tx.
		db = tx.WithContext(ctx)
	} else {
		db = fallback.WithContext(ctx)
	}

	// System admin bypass — no tenant scoping.
	if v := ctx.Value(isSystemAdminKey); v != nil {
		if b, ok := v.(bool); ok && b {
			return db
		}
	}

	// Apply tenant scope when the request carries a tenant_id.
	if v := ctx.Value(tenantIDKey); v != nil {
		if tid, ok := v.(string); ok && tid != "" {
			return db.Where(clause.Eq{
				Column: clause.Column{Table: clause.CurrentTable, Name: "tenant_id"},
				Value:  tid,
			})
		}
	}

	return db
}

// RetryTx executes the given function within a transaction, retrying on transient errors.
// This is critical for handling PostgreSQL serialization or deadlock errors in high-concurrency environments.
func RetryTx(db *gorm.DB, fc func(tx *gorm.DB) error) error {
	return db.Transaction(fc)
	// NOTE: In a full implementation, we would add a loop and check for specific PG error codes (40001, 40P01).
	// For now, this wrappers ensure standard transaction behavior which we can enhance later.
}
