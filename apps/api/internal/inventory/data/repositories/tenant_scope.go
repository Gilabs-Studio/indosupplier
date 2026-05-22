package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/middleware"
	"gorm.io/gorm"
)

func tenantContext(ctx context.Context) (string, bool, error) {
	if middleware.IsSystemAdmin(ctx) {
		return "", false, nil
	}

	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	if tenantID == "" {
		return "", false, errors.New("tenant_id not found in context")
	}

	return tenantID, true, nil
}

// applyTenantFilter adds tenant_id filtering to the query using specified table-qualified columns.
// This is used to resolve ambiguous column errors when joining multiple tables that all have tenant_id.
func applyTenantFilter(ctx context.Context, query *gorm.DB, qualifiedColumns ...string) (*gorm.DB, error) {
	tenantID, scoped, err := tenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if !scoped {
		return query, nil
	}

	for _, col := range qualifiedColumns {
		query = query.Where(fmt.Sprintf("%s = ?", col), tenantID)
	}

	return query, nil
}
