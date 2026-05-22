package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/analyzer"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

// RBACScanner checks permission coverage for API endpoints
type RBACScanner struct{}

func NewRBACScanner() *RBACScanner { return &RBACScanner{} }

func (s *RBACScanner) Name() string { return "RBAC Permission Scanner" }

func (s *RBACScanner) Run(cfg *analyzer.Config) []analyzer.Finding {
	var findings []analyzer.Finding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findings = append(findings, s.checkAdminHasAllPermissions(ctx)...)
	findings = append(findings, s.checkOrphanPermissions(ctx)...)
	findings = append(findings, s.checkMenuPermissionSync(ctx)...)

	return findings
}

// checkAdminHasAllPermissions ensures admin role has every permission assigned
func (s *RBACScanner) checkAdminHasAllPermissions(ctx context.Context) []analyzer.Finding {
	db := database.DB

	var totalPerms int64
	db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM permissions WHERE deleted_at IS NULL`).Scan(&totalPerms)

	var adminPerms int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		WHERE r.code = 'admin' AND r.deleted_at IS NULL
	`).Scan(&adminPerms)

	if totalPerms == 0 {
		return []analyzer.Finding{{
			Code:     "RBAC-001",
			Severity: analyzer.SeveritySkipped,
			Module:   "auth",
			Entity:   "permissions",
			Message:  "No permissions found in database — seeder may not have run",
		}}
	}

	if adminPerms < totalPerms {
		return []analyzer.Finding{{
			Code:           "RBAC-001",
			Severity:       analyzer.SeverityWarning,
			Module:         "auth",
			Entity:         "role_permissions",
			Message:        fmt.Sprintf("Admin role missing %d permissions (has %d/%d)", totalPerms-adminPerms, adminPerms, totalPerms),
			Recommendation: "Run SyncAdminPermissions or re-seed permissions.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "RBAC-001",
		Severity: analyzer.SeverityPass,
		Module:   "auth",
		Entity:   "role_permissions",
		Message:  fmt.Sprintf("Admin role has all %d permissions", totalPerms),
	}}
}

// checkOrphanPermissions finds permissions not assigned to any role
func (s *RBACScanner) checkOrphanPermissions(ctx context.Context) []analyzer.Finding {
	db := database.DB

	var count int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM permissions p
		WHERE p.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM role_permissions rp WHERE rp.permission_id = p.id
		  )
	`).Scan(&count)

	if count > 0 {
		return []analyzer.Finding{{
			Code:           "RBAC-002",
			Severity:       analyzer.SeverityWarning,
			Module:         "auth",
			Entity:         "permissions",
			Message:        fmt.Sprintf("%d permissions not assigned to any role", count),
			Recommendation: "Assign these permissions to at least one role or remove if obsolete.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "RBAC-002",
		Severity: analyzer.SeverityPass,
		Module:   "auth",
		Entity:   "permissions",
		Message:  "All permissions are assigned to at least one role",
	}}
}

// checkMenuPermissionSync verifies every menu has at least one permission
func (s *RBACScanner) checkMenuPermissionSync(ctx context.Context) []analyzer.Finding {
	db := database.DB

	// Menus with URLs that have no matching permission
	var count int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM menus m
		WHERE m.deleted_at IS NULL
		  AND m.url IS NOT NULL AND m.url != ''
		  AND m.parent_id IS NOT NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM permissions p
		    WHERE p.deleted_at IS NULL AND p.menu_id = m.id
		  )
	`).Scan(&count)

	if count > 0 {
		return []analyzer.Finding{{
			Code:           "RBAC-003",
			Severity:       analyzer.SeverityWarning,
			Module:         "auth",
			Entity:         "menus",
			Message:        fmt.Sprintf("%d child menus have no matching permissions", count),
			Evidence:       "Menus with URLs but no permission rows will be inaccessible via RBAC",
			Recommendation: "Add permission entries in the permission seeder for these menus.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "RBAC-003",
		Severity: analyzer.SeverityPass,
		Module:   "auth",
		Entity:   "menus",
		Message:  "All child menus have matching permissions",
	}}
}
