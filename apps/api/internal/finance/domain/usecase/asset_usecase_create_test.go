package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	fm "github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func buildAssetUsecaseForTest(t *testing.T) (*gorm.DB, AssetUsecase, func()) {
	db, cleanup := database.OpenTestDB(t)

	// Ensure minimal schema for tests when using in-memory sqlite by creating simple tables
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS asset_categories (id TEXT PRIMARY KEY, tenant_id TEXT, name TEXT, depreciation_method TEXT, useful_life_months INTEGER, is_depreciable INTEGER, asset_account_id TEXT, depreciation_expense_account_id TEXT, is_active INTEGER DEFAULT 1, deleted_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS asset_locations (id TEXT PRIMARY KEY, tenant_id TEXT, name TEXT, description TEXT, deleted_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS fixed_assets (id TEXT PRIMARY KEY, tenant_id TEXT, code TEXT, name TEXT, description TEXT, asset_type_id TEXT, category_id TEXT, location_id TEXT, company_id TEXT, business_unit_id TEXT, department_id TEXT, supplier_id TEXT, supplier_invoice_id TEXT, assigned_to_employee_id TEXT, custodian_user_id TEXT, assignment_date DATETIME, acquisition_date DATE, acquisition_cost NUMERIC, salvage_value NUMERIC, shipping_cost NUMERIC, installation_cost NUMERIC, tax_amount NUMERIC, other_costs NUMERIC, accumulated_depreciation NUMERIC, book_value NUMERIC, depreciation_method TEXT, useful_life_months INTEGER, status TEXT, lifecycle_stage TEXT, is_capitalized INTEGER, is_depreciable INTEGER, is_fully_depreciated INTEGER, parent_asset_id TEXT, is_parent INTEGER, serial_number TEXT, barcode TEXT, asset_tag TEXT, warranty_start DATE, warranty_end DATE, warranty_provider TEXT, warranty_terms TEXT, insurance_policy_number TEXT, insurance_provider TEXT, insurance_start DATE, insurance_end DATE, insurance_value NUMERIC, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS asset_assignment_histories (id TEXT PRIMARY KEY, tenant_id TEXT, asset_id TEXT, employee_id TEXT, department_id TEXT, location_id TEXT, assigned_at DATETIME, returned_at DATETIME, return_reason TEXT, notes TEXT, assigned_by TEXT, created_at DATETIME, deleted_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS asset_depreciations (id TEXT PRIMARY KEY, tenant_id TEXT, asset_id TEXT, period TEXT, depreciation_date DATE, method TEXT, amount NUMERIC, accumulated NUMERIC, book_value NUMERIC, status TEXT, journal_entry_id TEXT, created_by TEXT, created_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS asset_depreciation_schedules (id TEXT PRIMARY KEY, tenant_id TEXT, asset_id TEXT, period_start_date DATE, period_end_date DATE, period_month INTEGER, depreciation_amount NUMERIC, accumulated_depreciation NUMERIC, book_value NUMERIC, journal_entry_id TEXT, is_posted INTEGER, posted_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS asset_audit_logs (id TEXT PRIMARY KEY, asset_id TEXT, action TEXT, changes TEXT, performed_by TEXT, performed_at DATETIME, ip_address TEXT, user_agent TEXT, metadata TEXT, created_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS asset_transactions (id TEXT PRIMARY KEY, tenant_id TEXT, asset_id TEXT, type TEXT, transaction_date DATE, amount NUMERIC, description TEXT, status TEXT, reference_type TEXT, reference_id TEXT, proceeds_amount NUMERIC, bank_account_id TEXT, book_value_at_transaction NUMERIC, gain_loss_amount NUMERIC, gain_loss_account_id TEXT, created_by TEXT, created_at DATETIME);`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS finance_settings (id TEXT PRIMARY KEY, tenant_id TEXT, setting_key TEXT, value TEXT, deleted_at DATETIME);`)

	// create repos and mapper
	catRepo := repositories.NewAssetCategoryRepository(db)
	locRepo := repositories.NewAssetLocationRepository(db)
	assetRepo := repositories.NewAssetRepository(db)
	attachmentRepo := repositories.NewAssetAttachmentRepository(db)
	auditRepo := repositories.NewAssetAuditLogRepository(db)
	assignmentRepo := repositories.NewAssetAssignmentRepository(db)
	mapper := fm.NewAssetMapper(fm.NewAssetCategoryMapper(), fm.NewAssetLocationMapper())

	uc := NewAssetUsecase(db, nil, catRepo, locRepo, nil, assetRepo, mapper, attachmentRepo, auditRepo, assignmentRepo)
	return db, uc, cleanup
}

func TestCreateAsset_SetsWarrantyAndInsurance(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	// seed category and location
	if err := db.Exec(`INSERT INTO asset_categories (id, tenant_id, name, depreciation_method, useful_life_months, is_depreciable, asset_account_id, depreciation_expense_account_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "cat-1", "tenant-test", "TestCat", "SL", 12, 1, "1000", "2000", 1).Error; err != nil {
		t.Fatalf("seed category failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO asset_locations (id, tenant_id, name) VALUES (?, ?, ?)`, "loc-1", "tenant-test", "HQ").Error; err != nil {
		t.Fatalf("seed location failed: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", "tenant-test")
	ctx = context.WithValue(ctx, "user_id", "11111111-1111-1111-1111-111111111111")

	req := &dto.CreateAssetRequest{
		Name:               "Asset A",
		AssetTypeID:        "FIXED",
		CategoryID:         "cat-1",
		LocationID:         "loc-1",
		AcquisitionDate:    time.Now().Format("2006-01-02"),
		PurchasePrice:      1000,
		UsefulLifeMonths:   func(i int) *int { return &i }(12),
		DepreciationMethod: func(s string) *string { return &s }("SL"),
		WarrantyStart:      func(s string) *string { return &s }(time.Now().AddDate(0, 0, -1).Format("2006-01-02")),
		WarrantyEnd:        func(s string) *string { return &s }(time.Now().AddDate(0, 6, 0).Format("2006-01-02")),
		InsuranceStart:     func(s string) *string { return &s }(time.Now().AddDate(0, 0, -1).Format("2006-01-02")),
		InsuranceEnd:       func(s string) *string { return &s }(time.Now().AddDate(1, 0, 0).Format("2006-01-02")),
	}

	res, err := uc.Create(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.WarrantyStart)
	require.NotNil(t, res.WarrantyEnd)
	require.NotNil(t, res.InsuranceStart)
	require.NotNil(t, res.InsuranceEnd)
}

func TestAssign_ClosesPreviousAssignment(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	// seed category and location and asset
	if err := db.Exec(`INSERT INTO asset_categories (id, tenant_id, name, depreciation_method, useful_life_months, is_depreciable, asset_account_id, depreciation_expense_account_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "cat-2", "tenant-test", "TestCat2", "SL", 12, 0, "1000", "2000", 1).Error; err != nil {
		t.Fatalf("seed category2 failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO asset_locations (id, tenant_id, name) VALUES (?, ?, ?)`, "loc-2", "", "Branch").Error; err != nil {
		t.Fatalf("seed location2 failed: %v", err)
	}

	if err := db.Exec(`INSERT INTO fixed_assets (id, tenant_id, code, name, category_id, location_id, acquisition_date, acquisition_cost) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "asset-1", "tenant-test", "A1", "Asset1", "cat-2", "loc-2", time.Now().Format("2006-01-02"), 100).Error; err != nil {
		t.Fatalf("seed asset failed: %v", err)
	}

	// create existing assignment
	prevUUID := uuid.New()
	assetUUID, _ := uuid.Parse("asset-1")
	prev := &financeModels.AssetAssignmentHistory{ID: prevUUID, AssetID: assetUUID, AssignedAt: time.Now(), AssignedBy: func() *uuid.UUID { u, _ := uuid.Parse("11111111-1111-1111-1111-111111111111"); return &u }()}
	// set tenant
	tid, _ := uuid.Parse("tenant-test")
	prev.TenantID = &tid
	if err := db.Exec(`INSERT INTO asset_assignment_histories (id, tenant_id, asset_id, assigned_at, assigned_by) VALUES (?, ?, ?, ?, ?)`, prev.ID.String(), prev.TenantID.String(), prev.AssetID.String(), prev.AssignedAt.Format(time.RFC3339), prev.AssignedBy.String()).Error; err != nil {
		t.Fatalf("seed prev assignment failed: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", "tenant-test")
	ctx = context.WithValue(ctx, "user_id", "11111111-1111-1111-1111-111111111111")

	// perform assign via usecase
	_, err := uc.Assign(ctx, "asset-1", &dto.AssignAssetRequest{EmployeeID: "22222222-2222-2222-2222-222222222222"})
	require.NoError(t, err)

	// prev should have returned_at set
	var p financeModels.AssetAssignmentHistory
	row := db.Raw(`SELECT id, returned_at FROM asset_assignment_histories WHERE id = ?`, prev.ID.String()).Scan(&p)
	require.NoError(t, row.Error)
	require.NotNil(t, p.ReturnedAt)
}

func TestCreateAsset_BelowThreshold_NotCapitalized_NoSchedule(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	if err := db.Exec(`INSERT INTO asset_categories (id, tenant_id, name, depreciation_method, useful_life_months, is_depreciable, asset_account_id, depreciation_expense_account_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "cat-3", "tenant-test", "CapCat", "SL", 12, 1, "1000", "2000", 1).Error; err != nil {
		t.Fatalf("seed category3 failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO asset_locations (id, tenant_id, name) VALUES (?, ?, ?)`, "loc-3", "", "HQ").Error; err != nil {
		t.Fatalf("seed location3 failed: %v", err)
	}

	// set capitalization threshold high
	if err := db.Exec(`INSERT INTO finance_settings (id, tenant_id, setting_key, value) VALUES (?, ?, ?, ?)`, "fs-1", "tenant-test", "fixed_assets.capitalization_threshold", "10000").Error; err != nil {
		t.Fatalf("seed finance setting failed: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", "tenant-test")
	ctx = context.WithValue(ctx, "user_id", "11111111-1111-1111-1111-111111111111")

	req := &dto.CreateAssetRequest{
		Name:               "Small Asset",
		AssetTypeID:        "FIXED",
		CategoryID:         "cat-3",
		LocationID:         "loc-3",
		AcquisitionDate:    time.Now().Format("2006-01-02"),
		PurchasePrice:      100,
		UsefulLifeMonths:   func(i int) *int { return &i }(12),
		DepreciationMethod: func(s string) *string { return &s }("SL"),
	}

	res, err := uc.Create(ctx, req)
	require.NoError(t, err)
	require.False(t, res.IsCapitalized)

	// schedules should be empty
	var count int64
	require.NoError(t, db.Model(&financeModels.AssetDepreciationSchedule{}).Where("asset_id = ?", res.ID).Count(&count).Error)
	require.Equal(t, int64(0), count)
}
