package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type isReversalKey struct{}

// WithReversalFlag returns a context flagged as a reversal operation.
// Reversal entries are exempt from period-closed checks because they have net zero impact.
func WithReversalFlag(ctx context.Context) context.Context {
	return context.WithValue(ctx, isReversalKey{}, true)
}

func isReversalContext(ctx context.Context) bool {
	v, _ := ctx.Value(isReversalKey{}).(bool)
	return v
}

func ensureNotClosed(ctx context.Context, tx *gorm.DB, entryDate time.Time) error {
	// Reversal entries are exempt — they have net zero impact on financials
	if isReversalContext(ctx) {
		return nil
	}

	// First, check for explicit closed accounting periods.
	var period financeModels.AccountingPeriod
	err := tx.WithContext(ctx).
		Where("status = ?", financeModels.AccountingPeriodStatusClosed).
		Where("? BETWEEN start_date AND end_date", entryDate).
		First(&period).Error
	if err == nil {
		return errors.New("period is closed")
	}
	if err != nil && strings.Contains(err.Error(), `relation "accounting_periods" does not exist`) {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// No explicitly-closed accounting period matched this date, so posting may continue.
	return nil
}

