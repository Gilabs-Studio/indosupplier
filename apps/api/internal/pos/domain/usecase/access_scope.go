package usecase

import (
	"context"
	"log"
	"strings"

	"github.com/gilabs/gims/api/internal/core/middleware"
	orgRepo "github.com/gilabs/gims/api/internal/organization/data/repositories"
)

func scopeString(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(key).(string)
	return strings.TrimSpace(value)
}

func currentPermissionScope(ctx context.Context) string {
	return strings.ToUpper(strings.TrimSpace(scopeString(ctx, "permission_scope")))
}

func hasPOSScopedFullAccess(ctx context.Context) bool {
	if middleware.IsSystemAdmin(ctx) {
		return true
	}

	return currentPermissionScope(ctx) == "ALL"
}

func resolveScopedPOSOutletIDs(ctx context.Context, outletRepo orgRepo.OutletRepository) ([]string, error) {
	if hasPOSScopedFullAccess(ctx) {
		log.Printf("[pos][scope] full-access role=%s scope=%s", scopeString(ctx, "user_role"), currentPermissionScope(ctx))
		return nil, nil
	}

	outletIDs, _ := ctx.Value("scope_outlet_ids").([]string)
	if len(outletIDs) > 0 {
		log.Printf("[pos][scope] using explicit employee outlets count=%d user_id=%s", len(outletIDs), scopeString(ctx, "user_id"))
		return outletIDs, nil
	}

	warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string)
	if len(warehouseIDs) > 0 {
		outlets, err := outletRepo.FindByWarehouseIDs(ctx, warehouseIDs)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(outlets))
		for _, outlet := range outlets {
			if outlet == nil || strings.TrimSpace(outlet.ID) == "" {
				continue
			}
			ids = append(ids, outlet.ID)
		}
		log.Printf("[pos][scope] resolved via warehouses warehouses=%d outlets=%d user_id=%s", len(warehouseIDs), len(ids), scopeString(ctx, "user_id"))
		return ids, nil
	}

	log.Printf("[pos][scope] no assignments user_id=%s role=%s scope=%s", scopeString(ctx, "user_id"), scopeString(ctx, "user_role"), currentPermissionScope(ctx))
	return []string{}, nil
}

func isOutletAllowed(outletIDs []string, outletID string) bool {
	for _, candidate := range outletIDs {
		if strings.TrimSpace(candidate) == strings.TrimSpace(outletID) {
			return true
		}
	}
	return false
}

// ResolveScopedPOSOutletIDs exposes POS outlet scope resolution for presentation
// code that must authorize long-lived connections before handing them to a hub.
func ResolveScopedPOSOutletIDs(ctx context.Context, outletRepo orgRepo.OutletRepository) ([]string, error) {
	return resolveScopedPOSOutletIDs(ctx, outletRepo)
}

// IsOutletAllowed reports whether outletID is included in a resolved POS scope.
func IsOutletAllowed(outletIDs []string, outletID string) bool {
	return isOutletAllowed(outletIDs, outletID)
}
