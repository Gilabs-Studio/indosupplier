package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	dbtest "github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"github.com/google/uuid"
)

func financeCreateAssetTestContext() context.Context {
	ctx := financeTestContext()
	ctx = context.WithValue(ctx, "tenant_id", financeTestCompanyID)
	ctx = context.WithValue(ctx, "permission_scope", "ALL")
	return ctx
}

func newCreateAssetTestUsecase(t *testing.T, db *gorm.DB) *assetUsecase {
	t.Helper()
	catRepo := repositories.NewAssetCategoryRepository(db)
	locRepo := repositories.NewAssetLocationRepository(db)
	assetRepo := repositories.NewAssetRepository(db)
	assetMapper := mapper.NewAssetMapper(mapper.NewAssetCategoryMapper(), mapper.NewAssetLocationMapper())
	return NewAssetUsecase(db, nil, catRepo, locRepo, nil, assetRepo, assetMapper, nil, nil, nil).(*assetUsecase)
}

func seedCreateAssetDependencies(t *testing.T, db *gorm.DB) (string, string) {
	t.Helper()

	categoryID := uuid.NewString()
	locationID := uuid.NewString()

	require.NoError(t, db.Create(&financeModels.AssetCategory{
		ID:                             categoryID,
		TenantID:                       financeTestCompanyID,
		Name:                           "Equipment",
		Type:                           financeModels.AssetCategoryTypeFixed,
		DepreciationMethod:             financeModels.DepreciationMethodStraightLine,
		UsefulLifeMonths:               12,
		IsDepreciable:                  true,
		AssetAccountID:                 uuid.NewString(),
		AccumulatedDepreciationAccountID: uuid.NewString(),
		DepreciationExpenseAccountID:    uuid.NewString(),
		IsActive:                       true,
	}).Error)

	require.NoError(t, db.Create(&financeModels.AssetLocation{
		ID:          locationID,
		TenantID:    financeTestCompanyID,
		Name:        "Head Office",
		Description: "Main location",
	}).Error)

	return categoryID, locationID
}

func seedCreateAssetPolicy(t *testing.T, db *gorm.DB, threshold string, approvalRequired string) {
	t.Helper()
	require.NoError(t, db.Create(&financeModels.FinanceSetting{
		TenantID:    financeTestCompanyID,
		SettingKey:  "fixed_assets.capitalization_threshold",
		Value:       threshold,
		Description: "Capitalization threshold",
		Category:    "fixed_assets",
	}).Error)
	require.NoError(t, db.Create(&financeModels.FinanceSetting{
		TenantID:    financeTestCompanyID,
		SettingKey:  "fixed_assets.approval_required",
		Value:       approvalRequired,
		Description: "Approval required",
		Category:    "fixed_assets",
	}).Error)
}

func prepareAssetCreateSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`CREATE TABLE IF NOT EXISTS finance_settings (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			setting_key TEXT NOT NULL,
			value TEXT NOT NULL,
			description TEXT,
			category TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_categories (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			depreciation_method TEXT NOT NULL,
			useful_life_months INTEGER NOT NULL,
			depreciation_rate NUMERIC,
			is_depreciable BOOLEAN,
			asset_account_id TEXT NOT NULL,
			accumulated_depreciation_account_id TEXT NOT NULL,
			depreciation_expense_account_id TEXT NOT NULL,
			disposal_gain_account_id TEXT,
			disposal_loss_account_id TEXT,
			is_active BOOLEAN,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_locations (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			name TEXT NOT NULL,
			description TEXT,
			address TEXT,
			latitude NUMERIC,
			longitude NUMERIC,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS fixed_assets (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			code TEXT,
			name TEXT,
			description TEXT,
			asset_type_id TEXT,
			serial_number TEXT,
			barcode TEXT,
			qr_code TEXT,
			asset_tag TEXT,
			category_id TEXT,
			location_id TEXT,
			company_id TEXT,
			business_unit_id TEXT,
			department_id TEXT,
			assigned_to_employee_id TEXT,
			assignment_date TIMESTAMP,
			acquisition_date DATE,
			acquisition_cost NUMERIC,
			salvage_value NUMERIC,
			supplier_id TEXT,
			purchase_order_id TEXT,
			supplier_invoice_id TEXT,
			custodian_user_id TEXT,
			shipping_cost NUMERIC,
			installation_cost NUMERIC,
			tax_amount NUMERIC,
			other_costs NUMERIC,
			accumulated_depreciation NUMERIC,
			book_value NUMERIC,
			depreciation_method TEXT,
			useful_life_months INTEGER,
			depreciation_start_date DATE,
			status TEXT,
			lifecycle_stage TEXT,
			disposed_at TIMESTAMP,
			is_capitalized BOOLEAN,
			is_depreciable BOOLEAN,
			is_fully_depreciated BOOLEAN,
			parent_asset_id TEXT,
			is_parent BOOLEAN,
			warranty_start DATE,
			warranty_end DATE,
			warranty_provider TEXT,
			warranty_terms TEXT,
			insurance_policy_number TEXT,
			insurance_provider TEXT,
			insurance_start DATE,
			insurance_end DATE,
			insurance_value NUMERIC,
			created_by TEXT,
			approved_by TEXT,
			approved_at TIMESTAMP,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_transactions (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			asset_id TEXT,
			type TEXT,
			transaction_date DATE,
			amount NUMERIC,
			description TEXT,
			status TEXT,
			reference_type TEXT,
			reference_id TEXT,
			proceeds_amount NUMERIC,
			bank_account_id TEXT,
			book_value_at_transaction NUMERIC,
			gain_loss_amount NUMERIC,
			gain_loss_account_id TEXT,
			created_by TEXT,
			created_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_depreciations (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			asset_id TEXT,
			period TEXT,
			depreciation_date DATE,
			method TEXT,
			amount NUMERIC,
			accumulated NUMERIC,
			book_value NUMERIC,
			status TEXT,
			journal_entry_id TEXT,
			created_by TEXT,
			created_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_audit_logs (
			id TEXT PRIMARY KEY,
			asset_id TEXT,
			action TEXT,
			changes TEXT,
			performed_by TEXT,
			performed_at TIMESTAMP,
			ip_address TEXT,
			user_agent TEXT,
			metadata TEXT,
			created_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_assignment_histories (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			asset_id TEXT,
			employee_id TEXT,
			department_id TEXT,
			location_id TEXT,
			assigned_at TIMESTAMP,
			assigned_by TEXT,
			returned_at TIMESTAMP,
			return_reason TEXT,
			notes TEXT,
			created_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_depreciation_schedules (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			asset_id TEXT,
			period_start_date DATE,
			period_end_date DATE,
			period_month INTEGER,
			depreciation_amount NUMERIC,
			accumulated_depreciation NUMERIC,
			book_value NUMERIC,
			journal_entry_id TEXT,
			is_posted BOOLEAN,
			posted_at TIMESTAMP,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)`,
	}

	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
}

func TestAssetCreate_BelowThresholdCreatesNonCapitalizedAsset(t *testing.T) {
	db, cleanup := dbtest.OpenTestDB(t)
	defer cleanup()

	prepareAssetCreateSchema(t, db)

	categoryID, locationID := seedCreateAssetDependencies(t, db)
	seedCreateAssetPolicy(t, db, "1000", "true")

	uc := newCreateAssetTestUsecase(t, db)
	ctx := financeCreateAssetTestContext()
	ctx = context.WithValue(ctx, "user_id", financeTestCompanyID)

	threshold, approvalRequired, err := uc.loadCreatePolicy(ctx, financeTestCompanyID)
	require.NoError(t, err)
	require.Equal(t, 1000.0, threshold)
	require.True(t, approvalRequired)

	created, err := uc.Create(ctx, &dto.CreateAssetRequest{
		Name:            "Laptop",
		AssetTypeID:     "FIXED",
		CategoryID:      categoryID,
		LocationID:      locationID,
		AcquisitionDate: "2026-05-01",
		PurchasePrice:   500,
		AcquisitionCost: 500,
		UsefulLifeMonths: func() *int { v := 12; return &v }(),
		DepreciationMethod: func() *string { v := "SL"; return &v }(),
	})
	require.NoError(t, err)
	require.False(t, created.IsCapitalized)
	require.False(t, created.IsDepreciable)
	require.Equal(t, financeModels.AssetStatusPendingCapitalization, created.Status)

	var scheduleCount int64
	require.NoError(t, db.Model(&financeModels.AssetDepreciationSchedule{}).Where("asset_id = ?", created.ID).Count(&scheduleCount).Error)
	require.Zero(t, scheduleCount)
}

func TestAssetCreate_DuplicateCodeIsRejectedPerTenant(t *testing.T) {
	db, cleanup := dbtest.OpenTestDB(t)
	defer cleanup()

	prepareAssetCreateSchema(t, db)

	categoryID, locationID := seedCreateAssetDependencies(t, db)
	seedCreateAssetPolicy(t, db, "1000", "true")

	uc := newCreateAssetTestUsecase(t, db)
	ctx := financeCreateAssetTestContext()
	ctx = context.WithValue(ctx, "user_id", financeTestCompanyID)

	first, err := uc.Create(ctx, &dto.CreateAssetRequest{
		Code:            "AST-0001",
		Name:            "Monitor",
		AssetTypeID:     "FIXED",
		CategoryID:      categoryID,
		LocationID:      locationID,
		AcquisitionDate: "2026-05-01",
		PurchasePrice:   500,
		AcquisitionCost: 500,
		UsefulLifeMonths: func() *int { v := 12; return &v }(),
		DepreciationMethod: func() *string { v := "SL"; return &v }(),
	})
	require.NoError(t, err)
	require.Equal(t, "AST-0001", first.Code)

	_, err = uc.Create(ctx, &dto.CreateAssetRequest{
		Code:            "AST-0001",
		Name:            "Monitor 2",
		AssetTypeID:     "FIXED",
		CategoryID:      categoryID,
		LocationID:      locationID,
		AcquisitionDate: "2026-05-01",
		PurchasePrice:   500,
		AcquisitionCost: 500,
		UsefulLifeMonths: func() *int { v := 12; return &v }(),
		DepreciationMethod: func() *string { v := "SL"; return &v }(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "asset code already exists")
}

func TestAssetAssignEmployee_IsRolledBackAtomically(t *testing.T) {
	db, cleanup := dbtest.OpenTestDB(t)
	defer cleanup()

	prepareAssetCreateSchema(t, db)

	categoryID, locationID := seedCreateAssetDependencies(t, db)
	asset := financeModels.Asset{
		TenantID:               financeTestCompanyID,
		Code:                   "AST-ROLLBACK",
		Name:                   "Printer",
		AssetTypeID:            func() *string { v := "FIXED"; return &v }(),
		CategoryID:             categoryID,
		LocationID:             locationID,
		AcquisitionDate:        time.Now(),
		AcquisitionCost:        1000,
		BookValue:              1000,
		Status:                 financeModels.AssetStatusActive,
		LifecycleStage:         financeModels.AssetLifecycleActive,
		IsCapitalized:          true,
		IsDepreciable:          true,
	}
	require.NoError(t, db.Create(&asset).Error)

	uc := newCreateAssetTestUsecase(t, db)
	ctx := financeCreateAssetTestContext()
	ctx = context.WithValue(ctx, "user_id", financeTestCompanyID)

	assignErr := db.Transaction(func(tx *gorm.DB) error {
		if err := uc.assignEmployeeTx(tx, &asset, "00000000-0000-0000-0000-000000000777", financeTestCompanyID, nil); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	require.Error(t, assignErr)

	refetched, err := uc.repo.FindByID(ctx, asset.ID, false)
	require.NoError(t, err)
	require.Nil(t, refetched.AssignedToEmployeeID)

	var historyCount int64
	require.NoError(t, db.Model(&financeModels.AssetAssignmentHistory{}).Where("asset_id = ?", asset.ID).Count(&historyCount).Error)
	require.Zero(t, historyCount)
}
