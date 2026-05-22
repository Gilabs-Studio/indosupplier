package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/middleware"
	"gorm.io/gorm"
)

func applyTenantJoinScope(ctx context.Context, query *gorm.DB, qualifiedColumns ...string) (*gorm.DB, error) {
	if query == nil {
		return nil, errors.New("query is nil")
	}

	if middleware.IsSystemAdmin(ctx) {
		return query, nil
	}

	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	for _, column := range qualifiedColumns {
		if strings.TrimSpace(column) == "" {
			continue
		}
		query = query.Where(column+" = ?", tenantID)
	}

	return query, nil
}
