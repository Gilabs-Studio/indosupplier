package usecase

import (
	"context"
	"testing"

	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRunDepreciation_RollsBackWholeBatchOnFatalError(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS employees (
		id TEXT PRIMARY KEY,
		user_id TEXT,
		company_id TEXT,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS journal_entries (
		id TEXT PRIMARY KEY,
		tenant_id TEXT,
		company_id TEXT,
		fiscal_year_id TEXT,
		journal_number TEXT,
		entry_date DATE,
		reference TEXT,
		description TEXT,
		reference_type TEXT,
		reference_id TEXT,
		status TEXT,
		journal_type TEXT,
		posted_by TEXT,
		posted_at TIMESTAMP,
		reversed_by TEXT,
		reversed_at TIMESTAMP,
		debit_total NUMERIC,
		credit_total NUMERIC,
		currency_code TEXT,
		exchange_rate NUMERIC,
		original_journal_id TEXT,
		reversal_reason TEXT,
		is_reversal BOOLEAN,
		reversed_from TEXT,
		created_by TEXT,
		is_system_generated BOOLEAN,
		is_valuation BOOLEAN,
		source TEXT,
		valuation_run_id TEXT,
		source_document_url TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS journal_lines (
		id TEXT PRIMARY KEY,
		tenant_id TEXT,
		journal_entry_id TEXT,
		chart_of_account_id TEXT,
		chart_of_account_code_snapshot TEXT,
		chart_of_account_name_snapshot TEXT,
		chart_of_account_type_snapshot TEXT,
		debit NUMERIC,
		credit NUMERIC,
		memo TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	)`).Error)

	if db.Dialector.Name() == "sqlite" {
		require.NoError(t, db.Exec(`ALTER TABLE asset_categories ADD COLUMN accumulated_depreciation_account_id TEXT`).Error)
	} else {
		require.NoError(t, db.Exec(`ALTER TABLE asset_categories ADD COLUMN IF NOT EXISTS accumulated_depreciation_account_id TEXT`).Error)
	}

	tenantID := "tenant-test"
	companyID := uuid.NewString()
	locationID := uuid.NewString()
	validCategoryID := uuid.NewString()
	invalidCategoryID := uuid.NewString()
	validAssetID := uuid.NewString()
	invalidAssetID := uuid.NewString()

	require.NoError(t, db.Exec(`INSERT INTO asset_locations (id, tenant_id, name) VALUES (?, ?, ?)`, locationID, tenantID, "HQ").Error)
	require.NoError(t, db.Exec(`INSERT INTO asset_categories (id, tenant_id, name, depreciation_method, useful_life_months, is_depreciable, asset_account_id, accumulated_depreciation_account_id, depreciation_expense_account_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		validCategoryID, tenantID, "Valid Cat", "SL", 12, 1, uuid.NewString(), uuid.NewString(), uuid.NewString(), 1).Error)
	require.NoError(t, db.Exec(`INSERT INTO asset_categories (id, tenant_id, name, depreciation_method, useful_life_months, is_depreciable, asset_account_id, accumulated_depreciation_account_id, depreciation_expense_account_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		invalidCategoryID, tenantID, "Invalid Cat", "SL", 12, 1, uuid.NewString(), "", "", 1).Error)
	require.NoError(t, db.Exec(`INSERT INTO fixed_assets (id, tenant_id, company_id, code, name, category_id, location_id, acquisition_date, acquisition_cost, salvage_value, accumulated_depreciation, book_value, depreciation_method, useful_life_months, status, lifecycle_stage, is_capitalized, is_depreciable, is_fully_depreciated, is_parent) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		validAssetID, tenantID, companyID, "AST-001", "Valid Asset", validCategoryID, locationID, "2026-01-10", 1200, 0, 0, 1200, "SL", 12, "active", "in_use", 1, 1, 0, 0).Error)
	require.NoError(t, db.Exec(`INSERT INTO fixed_assets (id, tenant_id, company_id, code, name, category_id, location_id, acquisition_date, acquisition_cost, salvage_value, accumulated_depreciation, book_value, depreciation_method, useful_life_months, status, lifecycle_stage, is_capitalized, is_depreciable, is_fully_depreciated, is_parent) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		invalidAssetID, tenantID, companyID, "AST-002", "Invalid Asset", invalidCategoryID, locationID, "2026-01-10", 1200, 0, 0, 1200, "SL", 12, "active", "in_use", 1, 1, 0, 0).Error)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "company_id", companyID)
	ctx = context.WithValue(ctx, "user_id", uuid.NewString())

	res, err := uc.RunDepreciation(ctx, &dto.RunDepreciationRequest{Period: "2026-05"})
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "asset category depreciation accounts are not configured")

	var depreciationCount int64
	require.NoError(t, db.Model(&models.AssetDepreciation{}).Count(&depreciationCount).Error)
	require.Equal(t, int64(0), depreciationCount)

	var journalEntryCount int64
	require.NoError(t, db.Model(&models.JournalEntry{}).Count(&journalEntryCount).Error)
	require.Equal(t, int64(0), journalEntryCount)

	var journalLineCount int64
	require.NoError(t, db.Model(&models.JournalLine{}).Count(&journalLineCount).Error)
	require.Equal(t, int64(0), journalLineCount)

	var transactionCount int64
	require.NoError(t, db.Model(&models.AssetTransaction{}).Count(&transactionCount).Error)
	require.Equal(t, int64(0), transactionCount)

	var validAsset models.Asset
	require.NoError(t, db.First(&validAsset, "id = ?", validAssetID).Error)
	require.InDelta(t, 1200, validAsset.BookValue, 0.001)
	require.InDelta(t, 0, validAsset.AccumulatedDepreciation, 0.001)

	var invalidAsset models.Asset
	require.NoError(t, db.First(&invalidAsset, "id = ?", invalidAssetID).Error)
	require.InDelta(t, 1200, invalidAsset.BookValue, 0.001)
	require.InDelta(t, 0, invalidAsset.AccumulatedDepreciation, 0.001)
}