package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/middleware"
	"gorm.io/gorm"
)

func applyTenantFilter(ctx context.Context, query *gorm.DB, qualifiedColumns ...string) (*gorm.DB, error) {
	if query == nil {
		return nil, fmt.Errorf("query is nil")
	}

	if middleware.IsSystemAdmin(ctx) {
		return query, nil
	}

	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	if tenantID == "" {
		return nil, fmt.Errorf("tenant context is required")
	}

	for _, col := range qualifiedColumns {
		if strings.TrimSpace(col) == "" {
			continue
		}
		query = query.Where(col+" = ?", tenantID)
	}

	return query, nil
}
