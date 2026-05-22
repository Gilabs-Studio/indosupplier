package service

import (
	"fmt"
	"math"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

// DepreciationScheduleItem represents a single period's depreciation
type DepreciationScheduleItem struct {
	Period                  string    `json:"period"`
	PeriodMonth             int       `json:"period_month"`
	PeriodStartDate         time.Time `json:"period_start_date"`
	PeriodEndDate           time.Time `json:"period_end_date"`
	DepreciationAmount      float64   `json:"depreciation_amount"`
	AccumulatedDepreciation float64   `json:"accumulated_depreciation"`
	BookValue               float64   `json:"book_value"`
}

// DepreciationEngine handles all depreciation calculations
type DepreciationEngine struct {
	method           financeModels.DepreciationMethod
	acquisitionCost  float64
	salvageValue     float64
	usefulLifeMonths int
	acquisitionDate  time.Time
	depreciationStartDate time.Time
}

// NewDepreciationEngine creates a new depreciation engine instance
func NewDepreciationEngine(
	method financeModels.DepreciationMethod,
	acquisitionCost float64,
	salvageValue float64,
	usefulLifeMonths int,
	acquisitionDate time.Time,
	depreciationStartDate time.Time,
) (*DepreciationEngine, error) {
	// Validation
	if acquisitionCost < 0 {
		return nil, fmt.Errorf("acquisition cost cannot be negative: %f", acquisitionCost)
	}
	if salvageValue < 0 || salvageValue > acquisitionCost {
		return nil, fmt.Errorf("salvage value must be between 0 and acquisition cost")
	}
	if usefulLifeMonths <= 0 && method != financeModels.DepreciationMethodNone {
		return nil, fmt.Errorf("useful life must be positive: %d", usefulLifeMonths)
	}

	if depreciationStartDate.Before(acquisitionDate) {
		return nil, fmt.Errorf("depreciation start date cannot be before acquisition date")
	}

	return &DepreciationEngine{
		method:                method,
		acquisitionCost:       acquisitionCost,
		salvageValue:          salvageValue,
		usefulLifeMonths:      usefulLifeMonths,
		acquisitionDate:       acquisitionDate,
		depreciationStartDate: depreciationStartDate,
	}, nil
}

// GenerateSchedule generates a complete depreciation schedule
func (de *DepreciationEngine) GenerateSchedule(endDate time.Time) ([]DepreciationScheduleItem, error) {
	if de.method == financeModels.DepreciationMethodNone {
		return []DepreciationScheduleItem{}, nil
	}

	// Calculate number of months to depreciate
	monthsToDepreciate := int(endDate.Sub(de.depreciationStartDate).Hours() / (24 * 30))
	if monthsToDepreciate > de.usefulLifeMonths {
		monthsToDepreciate = de.usefulLifeMonths
	}

	schedules := make([]DepreciationScheduleItem, 0)
	accumulated := 0.0

	for month := 1; month <= monthsToDepreciate; month++ {
		periodStart := de.depreciationStartDate.AddDate(0, month-1, 0)
		periodEnd := de.depreciationStartDate.AddDate(0, month, 0).AddDate(0, 0, -1)

		monthlyAmount := de.calculateMonthlyDepreciation(month, accumulated)
		accumulated += monthlyAmount
		bookValue := de.acquisitionCost - accumulated

		// Ensure book value doesn't go below salvage value
		if bookValue < de.salvageValue {
			monthlyAmount -= (de.salvageValue - bookValue)
			accumulated = de.acquisitionCost - de.salvageValue
			bookValue = de.salvageValue
		}

		schedules = append(schedules, DepreciationScheduleItem{
			Period:                  periodStart.Format("2006-01"),
			PeriodMonth:             month,
			PeriodStartDate:         periodStart,
			PeriodEndDate:           periodEnd,
			DepreciationAmount:      round(monthlyAmount, 2),
			AccumulatedDepreciation: round(accumulated, 2),
			BookValue:               round(bookValue, 2),
		})

		// Stop if we've reached salvage value
		if bookValue <= de.salvageValue {
			break
		}
	}

	return schedules, nil
}

// calculateMonthlyDepreciation calculates depreciation for a specific month based on method
func (de *DepreciationEngine) calculateMonthlyDepreciation(month int, accumulatedSoFar float64) float64 {
	switch de.method {
	case financeModels.DepreciationMethodStraightLine:
		return de.calculateStraightLine()

	case financeModels.DepreciationMethodDecliningBalance:
		return de.calculateDecliningBalance(accumulatedSoFar)

	case financeModels.DepreciationMethodSumOfYearsDigits:
		return de.calculateSumOfYearsDigits(month)

	case financeModels.DepreciationMethodUnitsOfProduction:
		// UOP requires additional data (units produced per month)
		// For now, distribute equally
		return de.calculateUnitsOfProduction()

	case financeModels.DepreciationMethodNone:
		return 0

	default:
		return 0
	}
}

// calculateStraightLine: (Cost - Salvage Value) / Useful Life
// Monthly = Annual / 12
func (de *DepreciationEngine) calculateStraightLine() float64 {
	depreciableAmount := de.acquisitionCost - de.salvageValue
	return depreciableAmount / float64(de.usefulLifeMonths)
}

// calculateDecliningBalance: Book Value × (1 / Useful Life)
// Uses double declining balance rate (2 / Useful Life)
func (de *DepreciationEngine) calculateDecliningBalance(accumulatedDepreciation float64) float64 {
	bookValue := de.acquisitionCost - accumulatedDepreciation
	rate := 2.0 / float64(de.usefulLifeMonths) // Double declining balance

	monthlyDepreciation := bookValue * (rate / 12.0) // Monthly rate
	
	// Ensure we don't depreciate below salvage value
	remainingToDepreciate := bookValue - de.salvageValue
	if monthlyDepreciation > remainingToDepreciate {
		return remainingToDepreciate
	}

	return monthlyDepreciation
}

// calculateSumOfYearsDigits: (Remaining Months / Sum of All Months) × (Cost - Salvage Value)
// Sum = n × (n + 1) / 2, where n = useful life in years
// For monthly: n = useful life in months
func (de *DepreciationEngine) calculateSumOfYearsDigits(currentMonth int) float64 {
	n := de.usefulLifeMonths
	sum := float64(n * (n + 1) / 2)
	
	remainingMonths := float64(n - currentMonth + 1)
	depreciableAmount := de.acquisitionCost - de.salvageValue
	
	monthlyDepreciation := (remainingMonths / sum) * depreciableAmount
	
	return monthlyDepreciation
}

// calculateUnitsOfProduction: (Cost - Salvage Value) / Total Units × Units Produced This Period
// Without actual unit data, we distribute equally
func (de *DepreciationEngine) calculateUnitsOfProduction() float64 {
	depreciableAmount := de.acquisitionCost - de.salvageValue
	return depreciableAmount / float64(de.usefulLifeMonths)
}

// CalculateBookValue returns current book value given a date
func (de *DepreciationEngine) CalculateBookValue(date time.Time) (float64, float64) {
	if date.Before(de.depreciationStartDate) {
		return de.acquisitionCost, 0
	}

	// Generate schedule up to the given date
	schedules, _ := de.GenerateSchedule(date)
	
	if len(schedules) == 0 {
		return de.acquisitionCost, 0
	}

	lastSchedule := schedules[len(schedules)-1]
	return lastSchedule.BookValue, lastSchedule.AccumulatedDepreciation
}

// GetMonthlyDepreciationAmount returns the monthly depreciation amount
func (de *DepreciationEngine) GetMonthlyDepreciationAmount() float64 {
	if de.usefulLifeMonths == 0 {
		return 0
	}

	switch de.method {
	case financeModels.DepreciationMethodStraightLine:
		return de.calculateStraightLine()

	case financeModels.DepreciationMethodDecliningBalance:
		// For DB, return first month's depreciation as "average"
		return de.calculateDecliningBalance(0)

	case financeModels.DepreciationMethodSumOfYearsDigits:
		// Return first month's depreciation
		return de.calculateSumOfYearsDigits(1)

	case financeModels.DepreciationMethodUnitsOfProduction:
		return de.calculateUnitsOfProduction()

	default:
		return 0
	}
}

// GetAnnualDepreciationAmount returns annual depreciation (12 months)
func (de *DepreciationEngine) GetAnnualDepreciationAmount() float64 {
	monthly := de.GetMonthlyDepreciationAmount()
	return monthly * 12
}

// GetTotalDepreciation returns total depreciation over useful life
func (de *DepreciationEngine) GetTotalDepreciation() float64 {
	return de.acquisitionCost - de.salvageValue
}

// CalculateDepreciationPercentage returns depreciation as percentage of cost
func (de *DepreciationEngine) CalculateDepreciationPercentage(accumulatedDepreciation float64) float64 {
	if de.acquisitionCost == 0 {
		return 0
	}
	return (accumulatedDepreciation / de.acquisitionCost) * 100
}

// ValidateConfiguration checks if depreciation configuration is valid
func (de *DepreciationEngine) ValidateConfiguration() error {
	if de.acquisitionCost <= 0 {
		return fmt.Errorf("acquisition cost must be positive")
	}

	if de.salvageValue < 0 || de.salvageValue > de.acquisitionCost {
		return fmt.Errorf("salvage value must be between 0 and acquisition cost")
	}

	if de.usefulLifeMonths <= 0 && de.method != financeModels.DepreciationMethodNone {
		return fmt.Errorf("useful life must be positive for depreciable assets")
	}

	if de.depreciationStartDate.Before(de.acquisitionDate) {
		return fmt.Errorf("depreciation start date must be after or equal to acquisition date")
	}

	return nil
}

// ===== UTILITY FUNCTIONS =====

// round rounds a float64 to the specified number of decimal places
func round(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// ===== DEPRECIATION CALCULATION EXAMPLES =====

// Example usage in documentation:
/*

// Example 1: Straight Line Depreciation
// Asset: Laptop
// Cost: 10,000,000 IDR
// Salvage Value: 1,000,000 IDR
// Useful Life: 60 months (5 years)
// Depreciation = (10,000,000 - 1,000,000) / 60 = 150,000 IDR per month

engine, _ := NewDepreciationEngine(
	financeModels.DepreciationMethodStraightLine,
	10000000,
	1000000,
	60,
	time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
)

schedules, _ := engine.GenerateSchedule(time.Date(2029, 1, 15, 0, 0, 0, 0, time.UTC))
// Output: 60 schedule items, each with 150,000 monthly depreciation

// Example 2: Declining Balance Depreciation
// Rate = 2 / 60 = 3.33% monthly
// Month 1: 10,000,000 × 3.33% = 333,000
// Month 2: (10,000,000 - 333,000) × 3.33% = 321,889
// ... and so on

// Example 3: Sum of Years Digits
// Sum = 60 × 61 / 2 = 1,830
// Month 1: (60/1830) × 9,000,000 = 295,082
// Month 2: (59/1830) × 9,000,000 = 290,164
// ... decreasing each month

*/
