package tenant

import (
	"context"

	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ScopedDB returns a GORM session scoped to the current tenant.
// Deprecated: prefer using database.GetDB which applies tenant scoping automatically.
// Keep for any legacy call sites that explicitly need a scoped DB without going through GetDB.
func ScopedDB(db *gorm.DB, ctx context.Context) *gorm.DB {
	tenantID := middleware.TenantFromContext(ctx)
	if tenantID == "" {
		return db
	}

	return db.Where("tenant_id = ?", tenantID)
}

// WithTenantID sets tenant_id on a model map before creation.
// Useful for repositories that build attribute maps.
func WithTenantID(ctx context.Context, attrs map[string]interface{}) map[string]interface{} {
	tenantID := middleware.TenantFromContext(ctx)
	if tenantID != "" {
		attrs["tenant_id"] = tenantID
	}
	return attrs
}

func tenantFromStatement(tx *gorm.DB) (string, bool) {
	if tx.Statement == nil || tx.Statement.Context == nil {
		return "", false
	}

	ctx := tx.Statement.Context
	tenantID := middleware.TenantFromContext(ctx)
	if tenantID == "" {
		return "", false
	}

	return tenantID, true
}

func statementHasTenantColumn(tx *gorm.DB) bool {
	if tx.Statement == nil || tx.Statement.Schema == nil {
		return false
	}
	return tx.Statement.Schema.LookUpField("tenant_id") != nil || tx.Statement.Schema.LookUpField("TenantID") != nil
}

func tenantScopedColumn(tx *gorm.DB) string {
	if tx.Statement == nil {
		return "tenant_id"
	}

	tableName := tx.Statement.Table
	if tableName == "" && tx.Statement.Schema != nil {
		tableName = tx.Statement.Schema.Table
	}

	if tableName == "" {
		return "tenant_id"
	}

	return tableName + ".tenant_id"
}

func setTenantOnCreate(tx *gorm.DB) {
	tenantID, ok := tenantFromStatement(tx)
	if !ok {
		return
	}

	if !statementHasTenantColumn(tx) {
		return
	}

	tx.Statement.SetColumn("tenant_id", tenantID, true)
}

func scopeTenantStatement(tx *gorm.DB) {
	tenantID, ok := tenantFromStatement(tx)
	if !ok {
		return
	}
	if !statementHasTenantColumn(tx) {
		return
	}

	tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: tenantScopedColumn(tx), Value: tenantID},
	}})
}

// SetTenantCallback registers a GORM Create callback that automatically injects
// tenant_id on new records. Query/Update/Delete scoping is handled by database.GetDB
// which applies WHERE tenant_id = ? at the session level — this is more reliable than
// GORM callbacks because the table name is always known at GetDB call time.
func SetTenantCallback(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("saas:set_tenant_id", setTenantOnCreate)
	db.Callback().Query().Before("gorm:query").Register("saas:scope_tenant_query", scopeTenantStatement)
	db.Callback().Update().Before("gorm:update").Register("saas:scope_tenant_update", scopeTenantStatement)
	db.Callback().Delete().Before("gorm:delete").Register("saas:scope_tenant_delete", scopeTenantStatement)
	db.Callback().Row().Before("gorm:row").Register("saas:scope_tenant_row", scopeTenantStatement)
}
