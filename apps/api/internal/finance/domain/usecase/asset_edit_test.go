package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedEditTestAsset(t *testing.T, db *gorm.DB) (tenantID string, categoryID string, locationID string, assetID string) {
	t.Helper()

	tenantID = "tenant-test"
	categoryID = uuid.NewString()
	locationID = uuid.NewString()
	assetID = uuid.NewString()
	acquisitionDate := "2026-01-10"

	require.NoError(t, db.Exec(`INSERT INTO asset_categories (id, tenant_id, name, depreciation_method, useful_life_months, is_depreciable, asset_account_id, depreciation_expense_account_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, categoryID, tenantID, "TestCat", "SL", 12, 1, "1000", "2000", 1).Error)
	require.NoError(t, db.Exec(`INSERT INTO asset_locations (id, tenant_id, name) VALUES (?, ?, ?)`, locationID, tenantID, "HQ").Error)
	require.NoError(t, db.Exec(`INSERT INTO fixed_assets (id, tenant_id, code, name, description, asset_type_id, category_id, location_id, acquisition_date, acquisition_cost, salvage_value, shipping_cost, installation_cost, tax_amount, other_costs, accumulated_depreciation, book_value, depreciation_method, useful_life_months, status, lifecycle_stage, is_capitalized, is_depreciable, is_fully_depreciated, is_parent, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assetID, tenantID, "AST-001", "Original Asset", "Original description", "FIXED", categoryID, locationID, acquisitionDate, 1000, 0, 100, 50, 25, 10, 0, 1000, "SL", 12, "active", "in_use", 1, 1, 0, 0, nil).Error)

	return tenantID, categoryID, locationID, assetID
}

func baseEditRequest(categoryID, locationID string) *dto.EditAssetRequest {
	life := 12
	method := "SL"
	status := financeModels.AssetStatusActive
	acquisitionCost := 1000.0
	salvageValue := 0.0
	return &dto.EditAssetRequest{
		Code:             "AST-001",
		Name:             "Edited Asset",
		Description:      "Edited description",
		AssetTypeID:      "FIXED",
		CategoryID:       categoryID,
		LocationID:       locationID,
		AcquisitionDate:  "2026-01-10",
		AcquisitionCost:  &acquisitionCost,
		SalvageValue:     &salvageValue,
		Status:           status,
		UsefulLifeMonths: &life,
		DepreciationMethod: &method,
	}
}

func TestEditAsset_GroupA_NoScheduleRecalc(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	tenantID, categoryID, locationID, assetID := seedEditTestAsset(t, db)
	oldScheduleID := uuid.NewString()
	require.NoError(t, db.Exec(`INSERT INTO asset_depreciation_schedules (id, tenant_id, asset_id, period_start_date, period_end_date, period_month, depreciation_amount, accumulated_depreciation, book_value, is_posted, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		oldScheduleID, tenantID, assetID, "2026-01-10", "2026-02-09", 1, 83.33, 83.33, 916.67, 0, time.Now(), time.Now()).Error)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", uuid.NewString())

	req := baseEditRequest(categoryID, locationID)
	req.Name = "Edited Asset A"
	req.Description = "Updated description"

	_, err := uc.EditAsset(ctx, assetID, req)
	require.NoError(t, err)

	var asset financeModels.Asset
	require.NoError(t, db.First(&asset, "id = ?", assetID).Error)
	require.Equal(t, "Edited Asset A", asset.Name)
	require.Equal(t, "Updated description", asset.Description)

	var scheduleCount int64
	require.NoError(t, db.Model(&financeModels.AssetDepreciationSchedule{}).Where("asset_id = ? AND deleted_at IS NULL", assetID).Count(&scheduleCount).Error)
	require.Equal(t, int64(1), scheduleCount)
}

func TestEditAsset_UsefulLife_RecalcScheduleProspective(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	tenantID, categoryID, locationID, assetID := seedEditTestAsset(t, db)
	oldScheduleID := uuid.NewString()
	require.NoError(t, db.Exec(`INSERT INTO asset_depreciation_schedules (id, tenant_id, asset_id, period_start_date, period_end_date, period_month, depreciation_amount, accumulated_depreciation, book_value, is_posted, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		oldScheduleID, tenantID, assetID, "2026-01-10", "2026-02-09", 1, 83.33, 83.33, 916.67, 0, time.Now(), time.Now()).Error)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", uuid.NewString())

	req := baseEditRequest(categoryID, locationID)
	life := 1
	req.UsefulLifeMonths = &life
	req.Name = "Edited Asset B"

	_, err := uc.EditAsset(ctx, assetID, req)
	require.NoError(t, err)

	var deletedAt sql.NullTime
	require.NoError(t, db.Raw(`SELECT deleted_at FROM asset_depreciation_schedules WHERE id = ?`, oldScheduleID).Scan(&deletedAt).Error)
	require.True(t, deletedAt.Valid)

	var activeCount int64
	require.NoError(t, db.Model(&financeModels.AssetDepreciationSchedule{}).Where("asset_id = ? AND tenant_id = ? AND deleted_at IS NULL", assetID, tenantID).Count(&activeCount).Error)
	require.Equal(t, int64(1), activeCount)
}

func TestEditAsset_AssignedEmployee_UpdatesAssignmentHistory(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	tenantID, categoryID, locationID, assetID := seedEditTestAsset(t, db)
	oldEmployeeID := uuid.New()
	newEmployeeID := uuid.New()
	assignmentID := uuid.NewString()
	now := time.Now()
	require.NoError(t, db.Exec(`INSERT INTO asset_assignment_histories (id, tenant_id, asset_id, employee_id, assigned_at, assigned_by, returned_at, return_reason, notes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assignmentID, tenantID, assetID, oldEmployeeID.String(), now.AddDate(0, 0, -1).Format(time.RFC3339), uuid.NewString(), nil, nil, nil, now).Error)
	require.NoError(t, db.Exec(`UPDATE fixed_assets SET assigned_to_employee_id = ? WHERE id = ?`, oldEmployeeID.String(), assetID).Error)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", uuid.NewString())

	req := baseEditRequest(categoryID, locationID)
	req.AssignedToEmployeeID = func() *string { value := newEmployeeID.String(); return &value }()
	req.Name = "Edited Asset C"

	_, err := uc.EditAsset(ctx, assetID, req)
	require.NoError(t, err)

	var asset financeModels.Asset
	require.NoError(t, db.First(&asset, "id = ?", assetID).Error)
	require.NotNil(t, asset.AssignedToEmployeeID)
	require.Equal(t, newEmployeeID.String(), *asset.AssignedToEmployeeID)

	var previousReturnedAt sql.NullTime
	require.NoError(t, db.Raw(`SELECT returned_at FROM asset_assignment_histories WHERE id = ?`, assignmentID).Scan(&previousReturnedAt).Error)
	require.True(t, previousReturnedAt.Valid)

	var historyCount int64
	require.NoError(t, db.Model(&financeModels.AssetAssignmentHistory{}).Where("asset_id = ?", assetID).Count(&historyCount).Error)
	require.Equal(t, int64(2), historyCount)
}

func TestEditAsset_BookValueRecalculatesFromUpdatedCosts(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	tenantID, categoryID, locationID, assetID := seedEditTestAsset(t, db)
	require.NoError(t, db.Exec(`UPDATE fixed_assets SET shipping_cost = 0, installation_cost = 0, tax_amount = 0, other_costs = 0, acquisition_cost = 1000, book_value = 1000 WHERE id = ?`, assetID).Error)
	oldScheduleID := uuid.NewString()
	require.NoError(t, db.Exec(`INSERT INTO asset_depreciation_schedules (id, tenant_id, asset_id, period_start_date, period_end_date, period_month, depreciation_amount, accumulated_depreciation, book_value, is_posted, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		oldScheduleID, tenantID, assetID, "2026-01-10", "2026-02-09", 1, 0, 0, 1000, 0, time.Now(), time.Now()).Error)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", uuid.NewString())

	req := baseEditRequest(categoryID, locationID)
	shippingCost := 100.0
	installationCost := 50.0
	taxAmount := 25.0
	otherCosts := 10.0
	req.ShippingCost = &shippingCost
	req.InstallationCost = &installationCost
	req.TaxAmount = &taxAmount
	req.OtherCosts = &otherCosts

	_, err := uc.EditAsset(ctx, assetID, req)
	require.NoError(t, err)

	var asset financeModels.Asset
	require.NoError(t, db.First(&asset, "id = ?", assetID).Error)
	require.InDelta(t, 1185.0, asset.AcquisitionCost, 0.001)
	require.InDelta(t, 1185.0, asset.BookValue, 0.001)
}

func TestEditAsset_CodeChangeAfterActiveRejected(t *testing.T) {
	db, uc, cleanup := buildAssetUsecaseForTest(t)
	defer cleanup()

	tenantID, categoryID, locationID, assetID := seedEditTestAsset(t, db)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", uuid.NewString())

	req := baseEditRequest(categoryID, locationID)
	req.Code = "AST-999"
	req.Name = "Edited Asset D"

	_, err := uc.EditAsset(ctx, assetID, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "immutable")
}
