package usecase

import (
	"context"
	"fmt"
	"math"

	"gorm.io/gorm"
)

// ValuationStrategyValidator validates data integrity before strategy execution.
type ValuationStrategyValidator interface {
	// ValidateInventoryData checks that inventory_batches table has required columns and data.
	ValidateInventoryData(ctx context.Context) error

	// ValidateFXData checks that exchange_rate and currency columns exist for FX valuation.
	ValidateFXData(ctx context.Context) error

	// ValidateDepreciationData checks that asset_depreciations table has required columns.
	ValidateDepreciationData(ctx context.Context) error

	// ValidateNumericValue ensures a numeric value is valid (not NaN, not Inf).
	ValidateNumericValue(key string, value float64) error
}

type valuationStrategyValidator struct {
	db *gorm.DB
}

// NewValuationStrategyValidator creates a new strategy validator.
func NewValuationStrategyValidator(db *gorm.DB) ValuationStrategyValidator {
	return &valuationStrategyValidator{db: db}
}

// ValidateInventoryData checks inventory_batches structure.
func (v *valuationStrategyValidator) ValidateInventoryData(ctx context.Context) error {
	// Check table exists
	if !v.db.Migrator().HasTable("inventory_batches") {
		return fmt.Errorf("inventory_batches table not found — inventory valuation requires inventory module")
	}

	// Check required columns
	requiredCols := []string{"product_id", "current_quantity", "cost_price"}
	for _, col := range requiredCols {
		if !v.db.Migrator().HasColumn("inventory_batches", col) {
			return fmt.Errorf("inventory_batches.%s column not found — required for valuation calculation", col)
		}
	}

	return nil
}

// ValidateFXData checks exchange_rate and currency columns for FX valuation.
func (v *valuationStrategyValidator) ValidateFXData(ctx context.Context) error {
	// FX valuation requires exchange_rates table
	if !v.db.Migrator().HasTable("exchange_rates") {
		return fmt.Errorf("exchange_rates table not found — FX valuation requires exchange rate configuration")
	}

	// Check customer_invoices has required columns
	requiredARCols := []string{"remaining_amount", "currency_code"}
	for _, col := range requiredARCols {
		if !v.db.Migrator().HasColumn("customer_invoices", col) {
			return fmt.Errorf("customer_invoices.%s column missing — required for AR FX valuation", col)
		}
	}

	// Check supplier_invoices has required columns
	requiredAPCols := []string{"remaining_amount", "currency_code"}
	for _, col := range requiredAPCols {
		if !v.db.Migrator().HasColumn("supplier_invoices", col) {
			return fmt.Errorf("supplier_invoices.%s column missing — required for AP FX valuation", col)
		}
	}

	return nil
}

// ValidateDepreciationData checks asset_depreciations structure.
func (v *valuationStrategyValidator) ValidateDepreciationData(ctx context.Context) error {
	// Check table exists
	if !v.db.Migrator().HasTable("asset_depreciations") {
		return fmt.Errorf("asset_depreciations table not found — depreciation valuation requires asset module")
	}

	// Check required columns
	requiredCols := []string{"asset_id", "amount", "book_value", "depreciation_date", "status"}
	for _, col := range requiredCols {
		if !v.db.Migrator().HasColumn("asset_depreciations", col) {
			return fmt.Errorf("asset_depreciations.%s column not found — required for depreciation calculation", col)
		}
	}

	return nil
}

// ValidateNumericValue ensures a float64 is valid (not NaN or Inf).
func (v *valuationStrategyValidator) ValidateNumericValue(key string, value float64) error {
	if math.IsNaN(value) {
		return fmt.Errorf("%s is NaN (not a number) — likely calculation error", key)
	}
	if math.IsInf(value, 0) {
		return fmt.Errorf("%s is infinity — likely division by zero", key)
	}
	return nil
}
