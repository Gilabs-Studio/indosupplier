package service

import (
	"fmt"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

// CalculateDepreciationForPeriod calculates the depreciation amount, accumulated, and resulting book value
// for the given asset and period (format: "YYYY-MM"). It uses the existing DepreciationEngine for calculations
// and guarantees the amount will not exceed (book_value - salvage_value).
func CalculateDepreciationForPeriod(asset *financeModels.Asset, period string) (amount float64, accumulated float64, bookValue float64, err error) {
	if asset == nil {
		return 0, 0, 0, fmt.Errorf("asset is nil")
	}

	if !asset.CanDepreciate() {
		return 0, asset.AccumulatedDepreciation, asset.BookValue, nil
	}

	// parse period
	p, err := time.Parse("2006-01", period)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid period format: %w", err)
	}
	// use period end as reference date
	periodStart := time.Date(p.Year(), p.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	method := financeModels.DepreciationMethod(asset.GetDepreciationMethod())
	useful := asset.GetUsefulLifeMonths()
	deprStart := asset.DepreciationStartDate
	if deprStart == nil {
		// fallback to acquisition date
		tmp := asset.AcquisitionDate
		deprStart = &tmp
	}

	engine, err := NewDepreciationEngine(
		method,
		asset.AcquisitionCost,
		asset.SalvageValue,
		useful,
		asset.AcquisitionDate,
		*deprStart,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed create depreciation engine: %w", err)
	}

	// generate schedules up to periodEnd
	schedules, err := engine.GenerateSchedule(periodEnd)
	if err != nil {
		return 0, 0, 0, err
	}

	if len(schedules) == 0 {
		// no depreciation
		return 0, asset.AccumulatedDepreciation, asset.BookValue, nil
	}

	// find schedule item matching period
	var item *DepreciationScheduleItem
	for i := range schedules {
		if schedules[i].Period == periodStart.Format("2006-01") {
			item = &schedules[i]
			break
		}
	}

	if item == nil {
		// if not found, use last generated (possibly earlier than period)
		last := schedules[len(schedules)-1]
		item = &last
	}

	amount = item.DepreciationAmount
	accumulated = item.AccumulatedDepreciation
	bookValue = item.BookValue

	// ensure cap: amount should not exceed remaining depreciable
	remaining := asset.BookValue - asset.SalvageValue
	if remaining < 0 {
		remaining = 0
	}
	if amount > remaining {
		amount = remaining
		accumulated = asset.AccumulatedDepreciation + amount
		bookValue = asset.BookValue - amount
	}

	return round(amount, 2), round(accumulated, 2), round(bookValue, 2), nil
}
