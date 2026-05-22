package service

import (
	"context"
	"testing"

	notificationModels "github.com/gilabs/gims/api/internal/notification/data/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupApprovalNotificationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed opening sqlite db: %v", err)
	}

	schemaSQL := []string{
		`CREATE TABLE notifications (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			tenant_id TEXT,
			user_id TEXT NOT NULL,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			message TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			is_read BOOLEAN NOT NULL DEFAULT 0,
			read_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE TABLE users (id TEXT PRIMARY KEY, role_id TEXT, email TEXT, status TEXT, deleted_at DATETIME);`,
		`CREATE TABLE permissions (id TEXT PRIMARY KEY, code TEXT, action TEXT, resource TEXT);`,
		`CREATE TABLE role_permissions (role_id TEXT, permission_id TEXT, scope TEXT);`,
		`CREATE TABLE employees (id TEXT PRIMARY KEY, user_id TEXT, email TEXT, division_id TEXT, deleted_at DATETIME);`,
		`CREATE TABLE employee_areas (employee_id TEXT, area_id TEXT);`,
		`CREATE TABLE purchase_requisitions (id TEXT PRIMARY KEY, employee_id TEXT, created_by TEXT, deleted_at DATETIME);`,
		`CREATE TABLE purchase_orders (id TEXT PRIMARY KEY, created_by TEXT, deleted_at DATETIME);`,
	}

	for _, stmt := range schemaSQL {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed creating schema with %q: %v", stmt, err)
		}
	}

	return db
}

func TestCreateApprovalNotification_FallbackWhenScopeFilteredOut(t *testing.T) {
	db := setupApprovalNotificationTestDB(t)
	ctx := context.Background()

	seedSQL := []string{
		`INSERT INTO permissions (id, code, action, resource) VALUES ('perm_pr_approve', 'purchase_requisition.approve', 'APPROVE', 'purchase_requisition');`,
		`INSERT INTO users (id, role_id, email, status, deleted_at) VALUES ('u_manager', 'r_manager', 'manager@example.com', 'active', NULL);`,
		`INSERT INTO users (id, role_id, email, status, deleted_at) VALUES ('u_staff', 'r_staff', 'staff@example.com', 'active', NULL);`,
		`INSERT INTO role_permissions (role_id, permission_id, scope) VALUES ('r_manager', 'perm_pr_approve', 'DIVISION');`,
		`INSERT INTO employees (id, user_id, email, division_id, deleted_at) VALUES ('e_manager', 'u_manager', 'manager@example.com', 'division_b', NULL);`,
		`INSERT INTO employees (id, user_id, email, division_id, deleted_at) VALUES ('e_staff', 'u_staff', 'staff@example.com', 'division_a', NULL);`,
		`INSERT INTO purchase_requisitions (id, employee_id, created_by, deleted_at) VALUES ('pr_1', 'e_staff', NULL, NULL);`,
	}

	for _, stmt := range seedSQL {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed seeding test data with %q: %v", stmt, err)
		}
	}

	err := CreateApprovalNotification(ctx, db, ApprovalNotificationParams{
		PermissionCode: "purchase_requisition.approve",
		EntityType:     "purchase_requisition",
		EntityID:       "pr_1",
		Title:          "Purchase Requisition Approval",
		Message:        "A purchase requisition has been submitted and requires your approval.",
		ActorUserID:    "u_staff",
	})
	if err != nil {
		t.Fatalf("CreateApprovalNotification returned error: %v", err)
	}

	var rows []notificationModels.Notification
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("failed reading notifications: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(rows))
	}
	if rows[0].UserID != "u_manager" {
		t.Fatalf("expected notification recipient u_manager, got %s", rows[0].UserID)
	}
	if rows[0].EntityType != "purchase_requisition" {
		t.Fatalf("expected entity type purchase_requisition, got %s", rows[0].EntityType)
	}
	if rows[0].EntityID != "pr_1" {
		t.Fatalf("expected entity id pr_1, got %s", rows[0].EntityID)
	}
}

func TestCreateApprovalNotification_FallbackByResourceApproveAction(t *testing.T) {
	db := setupApprovalNotificationTestDB(t)
	ctx := context.Background()

	seedSQL := []string{
		`INSERT INTO permissions (id, code, action, resource) VALUES ('perm_po_confirm', 'purchase_order.confirm', 'APPROVE', 'purchase_order');`,
		`INSERT INTO users (id, role_id, email, status, deleted_at) VALUES ('u_manager', 'r_manager', 'manager@example.com', 'active', NULL);`,
		`INSERT INTO users (id, role_id, email, status, deleted_at) VALUES ('u_staff', 'r_staff', 'staff@example.com', 'active', NULL);`,
		`INSERT INTO role_permissions (role_id, permission_id, scope) VALUES ('r_manager', 'perm_po_confirm', 'ALL');`,
		`INSERT INTO purchase_orders (id, created_by, deleted_at) VALUES ('po_1', 'u_staff', NULL);`,
	}

	for _, stmt := range seedSQL {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed seeding test data with %q: %v", stmt, err)
		}
	}

	err := CreateApprovalNotification(ctx, db, ApprovalNotificationParams{
		PermissionCode: "purchase_order.approve",
		EntityType:     "purchase_order",
		EntityID:       "po_1",
		Title:          "Purchase Order Approval",
		Message:        "A purchase order has been submitted and requires your approval.",
		ActorUserID:    "u_staff",
	})
	if err != nil {
		t.Fatalf("CreateApprovalNotification returned error: %v", err)
	}

	var rows []notificationModels.Notification
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("failed reading notifications: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(rows))
	}
	if rows[0].UserID != "u_manager" {
		t.Fatalf("expected notification recipient u_manager, got %s", rows[0].UserID)
	}
	if rows[0].EntityType != "purchase_order" {
		t.Fatalf("expected entity type purchase_order, got %s", rows[0].EntityType)
	}
	if rows[0].EntityID != "po_1" {
		t.Fatalf("expected entity id po_1, got %s", rows[0].EntityID)
	}
}

func TestCreateApprovalNotification_IncludesActorWhenActorHasApprovePermission(t *testing.T) {
	db := setupApprovalNotificationTestDB(t)
	ctx := context.Background()

	seedSQL := []string{
		`INSERT INTO permissions (id, code, action, resource) VALUES ('perm_pr_approve', 'purchase_requisition.approve', 'APPROVE', 'purchase_requisition');`,
		`INSERT INTO users (id, role_id, email, status, deleted_at) VALUES ('u_admin', 'r_admin', 'admin@example.com', 'active', NULL);`,
		`INSERT INTO role_permissions (role_id, permission_id, scope) VALUES ('r_admin', 'perm_pr_approve', 'ALL');`,
		`INSERT INTO employees (id, user_id, email, division_id, deleted_at) VALUES ('e_admin', 'u_admin', 'admin@example.com', 'division_a', NULL);`,
		`INSERT INTO purchase_requisitions (id, employee_id, created_by, deleted_at) VALUES ('pr_2', 'e_admin', 'u_admin', NULL);`,
	}

	for _, stmt := range seedSQL {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed seeding test data with %q: %v", stmt, err)
		}
	}

	err := CreateApprovalNotification(ctx, db, ApprovalNotificationParams{
		PermissionCode: "purchase_requisition.approve",
		EntityType:     "purchase_requisition",
		EntityID:       "pr_2",
		Title:          "Purchase Requisition Approval",
		Message:        "A purchase requisition has been submitted and requires your approval.",
		ActorUserID:    "u_admin",
	})
	if err != nil {
		t.Fatalf("CreateApprovalNotification returned error: %v", err)
	}

	var rows []notificationModels.Notification
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("failed reading notifications: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(rows))
	}
	if rows[0].UserID != "u_admin" {
		t.Fatalf("expected notification recipient u_admin, got %s", rows[0].UserID)
	}
}
