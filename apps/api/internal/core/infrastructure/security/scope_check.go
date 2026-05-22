package security

import (
	"context"

	"gorm.io/gorm"
)

// CheckRecordScopeAccess validates whether the current user (from context) has
// scope-based access to a specific record. It issues a lightweight COUNT query
// against the given table with the same ApplyScopeFilter logic used by List
// endpoints. Returns true when access is allowed.
//
// Usage from usecase GetByID:
//
//	if !security.CheckRecordScopeAccess(u.db, ctx, &models.DeliveryOrder{}, id, security.DefaultScopeQueryOptions()) {
//	    return nil, ErrDeliveryOrderNotFound
//	}
func CheckRecordScopeAccess(db *gorm.DB, ctx context.Context, model interface{}, recordID string, opts ScopeQueryOptions) bool {
	scope, _ := ctx.Value("permission_scope").(string)
	if scope == "" || scope == "ALL" {
		return true
	}

	query := db.Model(model).Where("id = ?", recordID)
	query = ApplyScopeFilter(query, ctx, opts)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}
