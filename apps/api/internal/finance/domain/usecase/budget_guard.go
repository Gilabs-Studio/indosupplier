package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

var ErrBudgetAccountNotConfiguredForPeriod = errors.New("account is not budgeted for period")

func IsBudgetAccountNotConfiguredForPeriod(err error) bool {
	return errors.Is(err, ErrBudgetAccountNotConfiguredForPeriod)
}

// EnsureWithinBudget checks if the proposed amount for a specific COA and date is within the approved budget.
func EnsureWithinBudget(ctx context.Context, tx *gorm.DB, coaID string, entryDate time.Time, amount float64) error {
	if amount <= 0 {
		return nil
	}

	// 1. Find an approved budget item for this COA that covers the entry date.
	// This avoids false negatives when multiple approved budgets overlap the same period.
	type budgetWindowItem struct {
		BudgetID  string    `gorm:"column:budget_id"`
		StartDate time.Time `gorm:"column:start_date"`
		EndDate   time.Time `gorm:"column:end_date"`
		Amount    float64   `gorm:"column:amount"`
	}

	var selected budgetWindowItem
	err := tx.WithContext(ctx).
		Table("budget_items AS bi").
		Select("bi.budget_id, b.start_date, b.end_date, bi.amount").
		Joins("JOIN budgets AS b ON b.id = bi.budget_id").
		Where("b.status = ?", financeModels.BudgetStatusApproved).
		Where("b.start_date <= ? AND b.end_date >= ?", entryDate, entryDate).
		Where("bi.chart_of_account_id = ?", coaID).
		Order("b.start_date DESC").
		Take(&selected).Error

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}

		// If there is no budget at all for the period, allow transaction.
		var periodBudget financeModels.Budget
		periodErr := tx.WithContext(ctx).
			Where("status = ?", financeModels.BudgetStatusApproved).
			Where("start_date <= ? AND end_date >= ?", entryDate, entryDate).
			Order("start_date DESC").
			Take(&periodBudget).Error

		if periodErr == gorm.ErrRecordNotFound {
			return nil
		}
		if periodErr != nil {
			return periodErr
		}

		// Budget exists for the period, but this account is not budgeted.
		return fmt.Errorf("%w %s - %s",
			ErrBudgetAccountNotConfiguredForPeriod,
			periodBudget.StartDate.Format("2006-01-02"), periodBudget.EndDate.Format("2006-01-02"))
	}

	// 3. Calculate actual spent from Journal Entries (Posted)
	type sumResult struct{ Total float64 }
	var actual sumResult
	err = tx.Table("journal_lines").
		Select("SUM(journal_lines.debit - journal_lines.credit) as total").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_lines.chart_of_account_id = ?", coaID).
		Where("journal_entries.status = ?", financeModels.JournalStatusPosted).
		Where("journal_entries.entry_date BETWEEN ? AND ?", selected.StartDate, selected.EndDate).
		Scan(&actual).Error

	if err != nil {
		return err
	}

	// 4. Validate: Actual + Proposed <= Budgeted
	if actual.Total+amount > selected.Amount+0.0001 {
		return fmt.Errorf("budget exceeded for account. Budget: %.2f, Already Spent: %.2f, Requested: %.2f",
			selected.Amount, actual.Total, amount)
	}

	return nil
}
