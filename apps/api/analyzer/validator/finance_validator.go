package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/analyzer"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

// FinanceValidator checks trial balance, P&L, balance sheet consistency
type FinanceValidator struct{}

func NewFinanceValidator() *FinanceValidator { return &FinanceValidator{} }

func (v *FinanceValidator) Name() string { return "Finance Validator (TB / P&L / BS)" }

func (v *FinanceValidator) Run(cfg *analyzer.Config) []analyzer.Finding {
	if !cfg.ShouldRunModule("finance") {
		return nil
	}

	var findings []analyzer.Finding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findings = append(findings, v.checkTrialBalance(ctx, cfg)...)
	findings = append(findings, v.checkProfitAndLoss(ctx, cfg)...)
	findings = append(findings, v.checkBalanceSheet(ctx, cfg)...)

	return findings
}

func (v *FinanceValidator) checkTrialBalance(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB
	var tb struct {
		DebitTotal  float64
		CreditTotal float64
	}

	err := db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(jl.debit),0) AS debit_total,
		       COALESCE(SUM(jl.credit),0) AS credit_total
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		WHERE je.status = 'posted'
		  AND je.deleted_at IS NULL
		  AND je.entry_date >= ? AND je.entry_date <= ?
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&tb).Error

	if err != nil {
		return []analyzer.Finding{{
			Code:     "FIN-001",
			Severity: analyzer.SeverityError,
			Module:   "finance",
			Entity:   "trial_balance",
			Message:  fmt.Sprintf("Failed to query trial balance: %v", err),
		}}
	}

	debitStr := fmt.Sprintf("%.2f", tb.DebitTotal)
	creditStr := fmt.Sprintf("%.2f", tb.CreditTotal)

	if debitStr != creditStr {
		return []analyzer.Finding{{
			Code:           "FIN-001",
			Severity:       analyzer.SeverityCritical,
			Module:         "finance",
			Entity:         "trial_balance",
			Message:        fmt.Sprintf("Trial Balance UNBALANCED: Debit=%s Credit=%s", debitStr, creditStr),
			Evidence:       fmt.Sprintf("Period %s to %s", cfg.FromDate.Format("2006-01-02"), cfg.ToDate.Format("2006-01-02")),
			Recommendation: "Check journal entries with unequal debit/credit lines. Run journal reconciliation.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "FIN-001",
		Severity: analyzer.SeverityPass,
		Module:   "finance",
		Entity:   "trial_balance",
		Message:  fmt.Sprintf("Trial Balance OK (Debit=Credit=%s)", debitStr),
	}}
}

func (v *FinanceValidator) checkProfitAndLoss(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	// Revenue (account type 'revenue' / codes starting with 4)
	var revenueTotal float64
	db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(jl.credit - jl.debit), 0)
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN chart_of_accounts coa ON coa.id = jl.chart_of_account_id
		WHERE je.status = 'posted' AND je.deleted_at IS NULL
		  AND coa.type = 'REVENUE'
		  AND je.entry_date >= ? AND je.entry_date <= ?
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&revenueTotal)

	// Expense (account types 'expense', 'cogs')
	var expenseTotal float64
	db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(jl.debit - jl.credit), 0)
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN chart_of_accounts coa ON coa.id = jl.chart_of_account_id
		WHERE je.status = 'posted' AND je.deleted_at IS NULL
		  AND coa.type IN ('EXPENSE', 'COST_OF_GOODS_SOLD', 'OPERATIONAL', 'SALARY_WAGES')
		  AND je.entry_date >= ? AND je.entry_date <= ?
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&expenseTotal)

	netProfit := revenueTotal - expenseTotal

	return []analyzer.Finding{{
		Code:     "FIN-002",
		Severity: analyzer.SeverityPass,
		Module:   "finance",
		Entity:   "profit_and_loss",
		Message:  fmt.Sprintf("P&L: Revenue=%.2f Expense=%.2f NetProfit=%.2f", revenueTotal, expenseTotal, netProfit),
		Evidence: fmt.Sprintf("Period %s to %s", cfg.FromDate.Format("2006-01-02"), cfg.ToDate.Format("2006-01-02")),
	}}
}

func (v *FinanceValidator) checkBalanceSheet(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	// Assets
	var assetTotal float64
	db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(jl.debit - jl.credit), 0)
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN chart_of_accounts coa ON coa.id = jl.chart_of_account_id
		WHERE je.status = 'posted' AND je.deleted_at IS NULL
		  AND coa.type IN ('ASSET', 'CASH_BANK', 'CURRENT_ASSET', 'FIXED_ASSET')
		  AND je.entry_date <= ?
	`, cfg.ToDateStr()).Scan(&assetTotal)

	// Liabilities
	var liabilityTotal float64
	db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(jl.credit - jl.debit), 0)
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN chart_of_accounts coa ON coa.id = jl.chart_of_account_id
		WHERE je.status = 'posted' AND je.deleted_at IS NULL
		  AND coa.type IN ('LIABILITY', 'TRADE_PAYABLE')
		  AND je.entry_date <= ?
	`, cfg.ToDateStr()).Scan(&liabilityTotal)

	// Equity
	var equityTotal float64
	db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(jl.credit - jl.debit), 0)
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN chart_of_accounts coa ON coa.id = jl.chart_of_account_id
		WHERE je.status = 'posted' AND je.deleted_at IS NULL
		  AND coa.type = 'EQUITY'
		  AND je.entry_date <= ?
	`, cfg.ToDateStr()).Scan(&equityTotal)

	liabilityEquity := liabilityTotal + equityTotal
	diff := assetTotal - liabilityEquity

	assetStr := fmt.Sprintf("%.2f", assetTotal)
	leStr := fmt.Sprintf("%.2f", liabilityEquity)

	if assetStr != leStr {
		return []analyzer.Finding{{
			Code:           "FIN-003",
			Severity:       analyzer.SeverityWarning,
			Module:         "finance",
			Entity:         "balance_sheet",
			Message:        fmt.Sprintf("Balance Sheet unbalanced: Assets=%s vs L+E=%s (diff=%.2f)", assetStr, leStr, diff),
			Evidence:       fmt.Sprintf("Assets=%.2f Liabilities=%.2f Equity=%.2f", assetTotal, liabilityTotal, equityTotal),
			Recommendation: "Retained earnings / net profit may not be reflected in equity yet. Run financial closing.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "FIN-003",
		Severity: analyzer.SeverityPass,
		Module:   "finance",
		Entity:   "balance_sheet",
		Message:  fmt.Sprintf("Balance Sheet OK (Assets=L+E=%s)", assetStr),
	}}
}
