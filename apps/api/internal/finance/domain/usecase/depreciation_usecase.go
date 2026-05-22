package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"gorm.io/gorm"
)

func parseYearMonthPeriod(period string) (time.Time, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		return time.Time{}, errors.New("PERIOD_REQUIRED")
	}
	parsed, err := time.ParseInLocation("2006-01", period, apptime.Location())
	if err != nil {
		return time.Time{}, errors.New("INVALID_PERIOD_FORMAT")
	}
	return parsed, nil
}

func calculateDepreciation(asset *financeModels.Asset, method financeModels.DepreciationMethod) (float64, error) {
	if asset == nil {
		return 0, errors.New("asset is required")
	}
	if asset.GetUsefulLifeMonths() <= 0 {
		return 0, errors.New("useful_life_months must be greater than zero")
	}
	if !asset.CanDepreciate() {
		return 0, errors.New("asset is fully depreciated")
	}

	bookValue := round2(asset.BookValue)
	if bookValue <= asset.SalvageValue {
		return 0, errors.New("asset is fully depreciated")
	}

	var amount float64
	switch method {
	case financeModels.DepreciationMethodStraightLine:
		amount = round2(asset.AcquisitionCost / float64(asset.GetUsefulLifeMonths()))
	case financeModels.DepreciationMethodDecliningBalance:
		amount = round2((2.0 / float64(asset.GetUsefulLifeMonths())) * bookValue)
	case financeModels.DepreciationMethodNone:
		return 0, errors.New("asset is not depreciable")
	default:
		return 0, fmt.Errorf("unsupported depreciation method: %s", method)
	}

	remainingFloor := math.Max(0, bookValue-asset.SalvageValue)
	if remainingFloor <= 0 {
		return 0, errors.New("asset is fully depreciated")
	}
	if amount > remainingFloor {
		amount = remainingFloor
	}
	if amount <= 0 {
		return 0, errors.New("depreciation amount must be greater than zero")
	}
	return round2(amount), nil
}

func (uc *assetUsecase) GetDepreciationSchedule(ctx context.Context, req *dto.RunDepreciationRequest) (*dto.DepreciationScheduleResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	periodStart, err := parseYearMonthPeriod(req.Period)
	if err != nil {
		return nil, err
	}
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	periodKey := ymPeriod(periodStart)

	q := database.GetDB(ctx, uc.db).
		Model(&financeModels.Asset{}).
		Preload("Category").
		Preload("Location").
		Where("status = ?", financeModels.AssetStatusActive).
		Where("acquisition_date <= ?", periodEnd)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if req.AssetID != nil && strings.TrimSpace(*req.AssetID) != "" {
		q = q.Where("id = ?", strings.TrimSpace(*req.AssetID))
	}
	if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		q = q.Where("category_id = ?", strings.TrimSpace(*req.CategoryID))
	}
	if req.LocationID != nil && strings.TrimSpace(*req.LocationID) != "" {
		q = q.Where("location_id = ?", strings.TrimSpace(*req.LocationID))
	}

	var assets []financeModels.Asset
	if err := q.Order("code asc").Find(&assets).Error; err != nil {
		return nil, err
	}

	items := make([]dto.DepreciationScheduleItem, 0, len(assets))
	posted := 0
	pending := 0
	totalAmount := 0.0

	for i := range assets {
		asset := &assets[i]
		item := dto.DepreciationScheduleItem{
			AssetID:                 asset.ID,
			AssetCode:               asset.Code,
			AssetName:               asset.Name,
			CategoryID:              asset.CategoryID,
			LocationID:              asset.LocationID,
			Method:                  asset.GetDepreciationMethod(),
			Period:                  periodKey,
			AcquisitionCost:         asset.AcquisitionCost,
			AccumulatedDepreciation: round2(asset.AccumulatedDepreciation),
			NetBookValue:            round2(asset.BookValue),
			Status:                  "skipped",
		}
		if asset.Category != nil {
			item.CategoryName = asset.Category.Name
		}
		if asset.Location != nil {
			item.LocationName = asset.Location.Name
		}

		var dep financeModels.AssetDepreciation
		err := database.GetDB(ctx, uc.db).
			Where("asset_id = ? AND period = ?", asset.ID, periodKey).
			First(&dep).Error
		if err == nil {
			item.DepreciationAmount = round2(dep.Amount)
			item.Posted = dep.Status == financeModels.AssetDepreciationStatusApproved || dep.JournalEntryID != nil
			item.Highlighted = !item.Posted
			item.JournalEntryID = dep.JournalEntryID
			if item.Posted {
				item.Status = "posted"
				posted++
				totalAmount += dep.Amount
			} else {
				item.Status = "pending"
				pending++
			}
		} else if err == gorm.ErrRecordNotFound {
			method := financeModels.DepreciationMethod(asset.GetDepreciationMethod())
			amount, calcErr := calculateDepreciation(asset, method)
			if calcErr != nil {
				item.SkipReason = calcErr.Error()
				item.Highlighted = true
			} else {
				item.DepreciationAmount = amount
				item.Posted = false
				item.Highlighted = true
				item.Status = "pending"
				pending++
				totalAmount += amount
			}
		} else {
			item.SkipReason = "failed to check depreciation status"
			item.Highlighted = true
		}

		items = append(items, item)
	}

	return &dto.DepreciationScheduleResponse{
		Period:      periodKey,
		TotalAssets: len(items),
		Posted:      posted,
		Pending:     pending,
		TotalAmount: round2(totalAmount),
		Items:       items,
	}, nil
}

func (uc *assetUsecase) GetDepreciationHistory(ctx context.Context, period string) (*dto.DepreciationHistoryResponse, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		return nil, errors.New("period is required")
	}

	var entries []financeModels.AssetDepreciation
	q := database.GetDB(ctx, uc.db).
		Preload("Asset").
		Where("period = ?", period)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if err := q.Order("created_at desc").Find(&entries).Error; err != nil {
		return nil, err
	}

	items := make([]dto.DepreciationHistoryItem, 0, len(entries))
	posted := 0
	pending := 0
	for i := range entries {
		entry := &entries[i]
		item := dto.DepreciationHistoryItem{
			DepreciationID: entry.ID,
			AssetID:        entry.AssetID,
			Period:         entry.Period,
			Amount:         round2(entry.Amount),
			Posted:         entry.Status == financeModels.AssetDepreciationStatusApproved || entry.JournalEntryID != nil,
			JournalEntryID: entry.JournalEntryID,
			CreatedAt:      entry.CreatedAt,
		}
		if entry.Asset != nil {
			item.AssetCode = entry.Asset.Code
			item.AssetName = entry.Asset.Name
		}
		if item.Posted {
			posted++
		} else {
			pending++
		}
		items = append(items, item)
	}

	return &dto.DepreciationHistoryResponse{
		Period:  period,
		Total:   len(items),
		Posted:  posted,
		Pending: pending,
		Items:   items,
	}, nil
}

func (uc *assetUsecase) RunDepreciation(ctx context.Context, req *dto.RunDepreciationRequest) (*dto.BatchDepreciationRunResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	periodStart, err := parseYearMonthPeriod(req.Period)
	if err != nil {
		return nil, err
	}
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	periodKey := ymPeriod(periodStart)

	tenantID := tenantIDFromContext(ctx)
	if tenantID != "" && uc.apRepo != nil {
		closed, err := uc.apRepo.IsClosed(ctx, tenantID, periodKey)
		if err != nil {
			return nil, err
		}
		if closed {
			return nil, errors.New("ACCOUNTING_PERIOD_CLOSED")
		}
	}

	q := database.GetDB(ctx, uc.db).
		Model(&financeModels.Asset{}).
		Preload("Category").
		Preload("Location").
		Where("status = ?", financeModels.AssetStatusActive).
		Where("acquisition_date <= ?", periodEnd)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if req.AssetID != nil && strings.TrimSpace(*req.AssetID) != "" {
		q = q.Where("id = ?", strings.TrimSpace(*req.AssetID))
	}
	if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		q = q.Where("category_id = ?", strings.TrimSpace(*req.CategoryID))
	}
	if req.LocationID != nil && strings.TrimSpace(*req.LocationID) != "" {
		q = q.Where("location_id = ?", strings.TrimSpace(*req.LocationID))
	}

	var assets []financeModels.Asset
	if err := q.Order("code asc").Find(&assets).Error; err != nil {
		return nil, err
	}

	currencyCode := currencyCodeFromContext(ctx)
	items := make([]dto.BatchDepreciationRunItem, 0, len(assets))
	skipped := 0
	failed := 0

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range assets {
			asset := &assets[i]
			runItem := dto.BatchDepreciationRunItem{
				AssetID:      asset.ID,
				AssetCode:    asset.Code,
				AssetName:    asset.Name,
				CurrencyCode: currencyCode,
			}

			if asset.Category == nil {
				runItem.Status = "skipped"
				runItem.Reason = "asset category not found"
				skipped++
				items = append(items, runItem)
				continue
			}

			if !asset.CanDepreciate() {
				runItem.Status = "skipped"
				runItem.Reason = "asset fully depreciated or not depreciable"
				skipped++
				items = append(items, runItem)
				continue
			}

			method := financeModels.DepreciationMethod(asset.GetDepreciationMethod())
			amount, calcErr := calculateDepreciation(asset, method)
			if calcErr != nil {
				runItem.Status = "failed"
				runItem.Reason = calcErr.Error()
				failed++
				items = append(items, runItem)
				continue
			}

			if amount <= 0 {
				runItem.Status = "skipped"
				runItem.Reason = "depreciation amount is zero"
				skipped++
				items = append(items, runItem)
				continue
			}

			bookValue := round2(asset.BookValue - amount)
			if bookValue < asset.SalvageValue {
				amount = round2(asset.BookValue - asset.SalvageValue)
				bookValue = round2(asset.SalvageValue)
			}
			if amount <= 0 {
				runItem.Status = "skipped"
				runItem.Reason = "asset already at salvage value"
				skipped++
				items = append(items, runItem)
				continue
			}

			if existingErr := tx.Where("asset_id = ? AND period = ?", asset.ID, periodKey).First(&financeModels.AssetDepreciation{}).Error; existingErr == nil {
				runItem.Status = "skipped"
				runItem.Reason = "period already processed for asset"
				skipped++
				items = append(items, runItem)
				continue
			} else if existingErr != gorm.ErrRecordNotFound {
				runItem.Status = "failed"
				runItem.Reason = existingErr.Error()
				failed++
				items = append(items, runItem)
				continue
			}

			dep := &financeModels.AssetDepreciation{
				TenantID:         asset.TenantID,
				AssetID:          asset.ID,
				Period:           periodKey,
				DepreciationDate: periodStart,
				Method:           method,
				Amount:           amount,
				Accumulated:      round2(asset.AccumulatedDepreciation + amount),
				BookValue:        bookValue,
				Status:           financeModels.AssetDepreciationStatusPending,
				CreatedBy:        &actorID,
			}
			if err := tx.Create(dep).Error; err != nil {
				runItem.Status = "failed"
				runItem.Reason = err.Error()
				failed++
				items = append(items, runItem)
				continue
			}

			runItem.Amount = amount
			runItem.Status = "pending"
			items = append(items, runItem)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.BatchDepreciationRunResponse{
		CurrencyCode: currencyCode,
		Processed:    len(items),
		Posted:       0,
		Skipped:      skipped,
		Failed:       failed,
		Items:        items,
	}, nil
}

func (uc *assetUsecase) ApproveDepreciationRun(ctx context.Context, req *dto.RunDepreciationRequest) (*dto.BatchDepreciationRunResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	periodStart, err := parseYearMonthPeriod(req.Period)
	if err != nil {
		return nil, err
	}
	periodKey := ymPeriod(periodStart)

	tx := database.GetDB(ctx, uc.db.WithContext(ctx))
	tx = security.ApplyScopeFilter(tx, ctx, security.FinanceScopeQueryOptions())

	tenantID := strings.TrimSpace(tenantIDFromContext(ctx))
	if tenantID == "" {
		return nil, errors.New("tenant not found in context")
	}

	query := tx.Model(&financeModels.AssetDepreciation{}).
		Select("asset_depreciations.id, asset_depreciations.asset_id, asset_depreciations.amount, assets.code AS asset_code, assets.name AS asset_name").
		Joins("JOIN assets ON assets.id = asset_depreciations.asset_id").
		Where("asset_depreciations.period = ?", periodKey).
		Where("asset_depreciations.status = ?", financeModels.AssetDepreciationStatusPending).
		Where("asset_depreciations.tenant_id = ?", tenantID)

	if req.AssetID != nil && strings.TrimSpace(*req.AssetID) != "" {
		query = query.Where("asset_depreciations.asset_id = ?", strings.TrimSpace(*req.AssetID))
	}
	if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		query = query.Where("assets.category_id = ?", strings.TrimSpace(*req.CategoryID))
	}
	if req.LocationID != nil && strings.TrimSpace(*req.LocationID) != "" {
		query = query.Where("assets.location_id = ?", strings.TrimSpace(*req.LocationID))
	}

	type pendingDepreciation struct {
		ID        string  `gorm:"column:id"`
		AssetID   string  `gorm:"column:asset_id"`
		Amount    float64 `gorm:"column:amount"`
		AssetCode string  `gorm:"column:asset_code"`
		AssetName string  `gorm:"column:asset_name"`
	}

	var pending []pendingDepreciation
	if err := query.Order("asset_depreciations.created_at asc").Find(&pending).Error; err != nil {
		return nil, err
	}

	currencyCode := currencyCodeFromContext(ctx)
	items := make([]dto.BatchDepreciationRunItem, 0, len(pending))
	posted := 0
	failed := 0

	for _, dep := range pending {
		item := dto.BatchDepreciationRunItem{
			AssetID:      dep.AssetID,
			AssetCode:    dep.AssetCode,
			AssetName:    dep.AssetName,
			CurrencyCode: currencyCode,
			Amount:       dep.Amount,
		}

		if _, err := uc.ApproveDepreciation(ctx, dep.ID); err != nil {
			item.Status = "failed"
			item.Reason = err.Error()
			failed++
			items = append(items, item)
			continue
		}

		var approvedDep financeModels.AssetDepreciation
		_ = uc.db.WithContext(ctx).
			Select("id", "journal_entry_id").
			Where("id = ?", dep.ID).
			First(&approvedDep).Error

		item.Status = "posted"
		item.JournalEntryID = approvedDep.JournalEntryID
		posted++
		items = append(items, item)
	}

	return &dto.BatchDepreciationRunResponse{
		CurrencyCode: currencyCode,
		Processed:    len(items),
		Posted:       posted,
		Skipped:      0,
		Failed:       failed,
		Items:        items,
	}, nil
}
