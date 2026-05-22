package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
)

func main() {
	log.Println("Starting finance validation...")

	if err := config.Load(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	ctx := context.Background()
	if err := runValidation(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Finance validation completed")
}

func runValidation(ctx context.Context) error {
	// Determine validation date range (defaults to current month)
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := now
	if raw := os.Getenv("VALIDATE_FINANCE_START"); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			start = t
		}
	}
	if raw := os.Getenv("VALIDATE_FINANCE_END"); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			end = t
		}
	}

	// Build usecases
	db := database.DB
	coaRepo := repositories.NewChartOfAccountRepository(db)
	reportRepo := repositories.NewFinanceReportRepository(db)
	reportUC := usecase.NewFinanceReportUsecase(db, coaRepo, reportRepo)

	journalRepo := repositories.NewJournalEntryRepository(db)
	journalMapper := mapper.NewJournalEntryMapper(mapper.NewChartOfAccountMapper())
	auditService := audit.NewAuditService(db)
	journalUC := usecase.NewJournalEntryUsecase(db, coaRepo, journalRepo, journalMapper, auditService)

	financialClosingRepo := repositories.NewFinancialClosingRepository(db)
	accountingPeriodRepo := repositories.NewAccountingPeriodRepository(db)
	financialClosingSnapshotRepo := repositories.NewFinancialClosingSnapshotRepository(db)
	financialClosingLogRepo := repositories.NewFinancialClosingLogRepository(db)
	financialClosingMapper := mapper.NewFinancialClosingMapper()
	financialClosingUC := usecase.NewFinancialClosingUsecase(
		db,
		coaRepo,
		financialClosingRepo,
		accountingPeriodRepo,
		financialClosingSnapshotRepo,
		financialClosingLogRepo,
		journalUC,
		financialClosingMapper,
	)

	fmt.Println("\n--- Finance Validation Report ---")
	fmt.Printf("Period: %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))

	// 1) Journal / Trial Balance
	if err := validateTrialBalance(ctx, start, end); err != nil {
		fmt.Println("❌ Trial Balance check failed:", err)
	} else {
		fmt.Println("✅ Trial Balance OK")
	}

	// 2) Profit & Loss
	if err := validateProfitAndLoss(ctx, reportUC, start, end); err != nil {
		fmt.Println("❌ P&L validation failed:", err)
	} else {
		fmt.Println("✅ P&L OK")
	}

	// 3) Cross-module reconciliation (Invoice totals vs journal totals)
	if err := validateCrossModule(ctx, reportUC, start, end); err != nil {
		fmt.Println("❌ Cross-module reconciliation failed:", err)
	} else {
		fmt.Println("✅ Cross-module reconciliation OK")
	}

	// 4) Balance Sheet
	if err := validateBalanceSheet(ctx, reportUC, start, end); err != nil {
		fmt.Println("❌ Balance Sheet validation failed:", err)
	} else {
		fmt.Println("✅ Balance Sheet OK")
	}

	// 5) Inventory vs GL
	if err := validateInventoryVsGL(ctx, start, end); err != nil {
		fmt.Println("❌ Inventory vs GL validation failed:", err)
	} else {
		fmt.Println("✅ Inventory vs GL OK")
	}

	// 6) Financial Closing
	if err := validateFinancialClosing(ctx, financialClosingRepo, financialClosingUC); err != nil {
		fmt.Println("❌ Financial Closing validation failed:", err)
	} else {
		fmt.Println("✅ Financial Closing validation OK")
	}

	return nil
}

func validateTrialBalance(ctx context.Context, start, end time.Time) error {
	var tb struct {
		DebitTotal  float64
		CreditTotal float64
	}
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(jl.debit),0) AS debit_total, COALESCE(SUM(jl.credit),0) AS credit_total
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		WHERE je.status = ?
		  AND je.entry_date >= ?
		  AND je.entry_date <= ?
	`, "posted", start, end).Scan(&tb).Error; err != nil {
		return err
	}
	if fmt.Sprintf("%.2f", tb.DebitTotal) != fmt.Sprintf("%.2f", tb.CreditTotal) {
		return fmt.Errorf("debit %.2f != credit %.2f", tb.DebitTotal, tb.CreditTotal)
	}
	return nil
}

func validateProfitAndLoss(ctx context.Context, uc usecase.FinanceReportUsecase, start, end time.Time) error {
	pl, err := uc.GetProfitAndLoss(ctx, start, end, nil, nil)
	if err != nil {
		return err
	}
	if fmt.Sprintf("%.2f", pl.NetProfit) != fmt.Sprintf("%.2f", pl.RevenueTotal-pl.ExpenseTotal) {
		return fmt.Errorf("net profit %.2f does not equal revenue (%.2f) - expense (%.2f)", pl.NetProfit, pl.RevenueTotal, pl.ExpenseTotal)
	}
	return nil
}

func validateCrossModule(ctx context.Context, uc usecase.FinanceReportUsecase, start, end time.Time) error {
	// Sales invoices vs revenue
	var salesInvTotal float64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(amount),0) FROM customer_invoices
		WHERE invoice_date >= ? AND invoice_date <= ?
	`, start, end).Scan(&salesInvTotal).Error; err != nil {
		return err
	}
	pl, err := uc.GetProfitAndLoss(ctx, start, end, nil, nil)
	if err != nil {
		return err
	}
	if fmt.Sprintf("%.2f", salesInvTotal) != fmt.Sprintf("%.2f", pl.RevenueTotal) {
		return fmt.Errorf("sales invoices total %.2f != revenue total %.2f", salesInvTotal, pl.RevenueTotal)
	}

	// Purchase invoices vs expenses+COGS
	var purchaseInvTotal float64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(amount),0) FROM supplier_invoices
		WHERE invoice_date >= ? AND invoice_date <= ?
	`, start, end).Scan(&purchaseInvTotal).Error; err != nil {
		return err
	}
	expensePlusCOGS := pl.ExpenseTotal + pl.COGSTotal
	if fmt.Sprintf("%.2f", purchaseInvTotal) != fmt.Sprintf("%.2f", expensePlusCOGS) {
		return fmt.Errorf("purchase invoices total %.2f != expense+COGS total %.2f", purchaseInvTotal, expensePlusCOGS)
	}

	return nil
}

func validateBalanceSheet(ctx context.Context, uc usecase.FinanceReportUsecase, start, end time.Time) error {
	bs, err := uc.GetBalanceSheet(ctx, start, end, nil, nil, true)
	if err != nil {
		return err
	}
	if !bs.IsBalanced {
		return fmt.Errorf("assets %.2f != liabilities+equity %.2f", bs.AssetTotal, bs.LiabilityEquity)
	}
	return nil
}

func validateInventoryVsGL(ctx context.Context, start, end time.Time) error {
	// Calculate inventory valuation directly from batches
	var inventoryValuation float64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(current_quantity * cost_price), 0)
		FROM inventory_batches
		WHERE deleted_at IS NULL
	`).Scan(&inventoryValuation).Error; err != nil {
		return err
	}

	// Calculate inventory account balance from GL
	// We sum balance of all accounts where type = 'CURRENT_ASSET' and system_reserved or name like inventory.
	// For safer generic check, let's query the specific account if it's set in posting profile,
	// or we sum all accounts that are inventory accounts.
	// Let's assume accounts with code prefix '14' or name containing 'Inventory' or configured as default inventory.
	var glInventoryBalance float64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(jl.debit - jl.credit), 0)
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN chart_of_accounts coa ON coa.id = jl.chart_of_account_id
		WHERE je.status = 'posted' 
		  AND je.entry_date <= ?
		  AND (coa.name ILIKE '%inventory%' OR coa.name ILIKE '%persediaan%')
	`, end).Scan(&glInventoryBalance).Error; err != nil {
		return err
	}

	if fmt.Sprintf("%.2f", inventoryValuation) != fmt.Sprintf("%.2f", glInventoryBalance) {
		return fmt.Errorf("inventory valuation %.2f != gl inventory balance %.2f", inventoryValuation, glInventoryBalance)
	}

	return nil
}

func validateFinancialClosing(ctx context.Context, repo repositories.FinancialClosingRepository, uc usecase.FinancialClosingUsecase) error {
	closing, err := repo.LatestApproved(ctx)
	if err != nil {
		return fmt.Errorf("no approved financial closing found: %w", err)
	}

	analysis, err := uc.GetAnalysis(ctx, closing.ID)
	if err != nil {
		return err
	}

	for _, v := range analysis.Validations {
		status := "✅"
		if !v.Passed {
			status = "❌"
		}
		fmt.Printf("%s %s: %s\n", status, v.Name, v.Message)
	}

	return nil
}
