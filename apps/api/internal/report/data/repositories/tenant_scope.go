package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/middleware"
)

func resolveReportTenantScope(ctx context.Context) (tenantID string, tenantFilter bool, err error) {
	if middleware.IsSystemAdmin(ctx) {
		return "", false, nil
	}

	tenantID = strings.TrimSpace(middleware.TenantFromContext(ctx))
	if tenantID == "" {
		return "", false, fmt.Errorf("tenant context is required")
	}

	return tenantID, true, nil
}

func withReportTenantParams(params map[string]interface{}, tenantID string, tenantFilter bool) map[string]interface{} {
	if params == nil {
		params = make(map[string]interface{}, 2)
	}
	params["tenantID"] = tenantID
	params["tenantFilter"] = tenantFilter
	return params
}
