package usecase

import (
	"context"
	"fmt"
	"math"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	inventoryModels "github.com/gilabs/gims/api/internal/inventory/data/models"
	"gorm.io/gorm"
)

// ReconciliationStatus indicates if GL posting matches subledger.
type ReconciliationStatus string

const (
	ReconciliationStatusMatched    ReconciliationStatus = "matched"
	ReconciliationStatusMismatched ReconciliationStatus = "mismatched"
	ReconciliationStatusNoJournal  ReconciliationStatus = "no_journal"
)

// ReconciliationDetail is a single line in reconciliation showing GL vs subledger.
type ReconciliationDetail struct {
	Account         string  `json:"account"`           // COA account code + name
	GLBalance       float64 `json:"gl_balance"`        // Balance from journal_lines
	SubledgerTotal  float64 `json:"subledger_total"`   // Balance from source table
	Delta           float64 `json:"delta"`             // Difference (GL - Subledger)
	Status          string  `json:"status"`            // "match" or "mismatch"
	Tolerance       float64 `json:"tolerance"`         // Allowed difference
	WithinTolerance bool    `json:"within_tolerance"` // Is delta < tolerance?
}

// ReconciliationReport is the complete reconciliation for a valuation run.
type ReconciliationReport struct {
	ValuationRunID   string                 `json:"valuation_run_id"`
	ValuationType    string                 `json:"valuation_type"`
	Status           ReconciliationStatus   `json:"status"` // matched, mismatched, no_journal
	OverallDelta     float64                `json:"overall_delta"`
	AllLinesMatched  bool                   `json:"all_lines_matched"`
	Details          []ReconciliationDetail `json:"details,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	CheckedAt        string                 `json:"checked_at"`
}

// ValuationReconciliationService validates GL posting accuracy for valuation runs.
type ValuationReconciliationService interface {
	// GenerateReconciliationReport compares GL balance against subledger for a valuation run.
	// Returns detailed reconciliation status for audit trail.
	GenerateReconciliationReport(ctx context.Context, valuationRunID string) (*ReconciliationReport, error)
}

type valuationReconciliationService struct {
	db               *gorm.DB
	repo             repositories.ValuationRunRepository
	coaRepo          repositories.ChartOfAccountRepository
	accountingEngine interface {
		ResolveCOAID(ctx context.Context, settingKey string) (string, error)
	}
	settings interface {
		GetValue(ctx context.Context, settingKey string) (string, error)
	}
}

// NewValuationReconciliationService creates a new reconciliation service.
func NewValuationReconciliationService(
	db *gorm.DB,
	repo repositories.ValuationRunRepository,
	coaRepo repositories.ChartOfAccountRepository,
	accountingEngine interface {
		ResolveCOAID(ctx context.Context, settingKey string) (string, error)
	},
	settings interface {
		GetValue(ctx context.Context, settingKey string) (string, error)
	},
) ValuationReconciliationService {
	return &valuationReconciliationService{
		db:               db,
		repo:             repo,
		coaRepo:          coaRepo,
		accountingEngine: accountingEngine,
		settings:         settings,
	}
}

// GenerateReconciliationReport validates GL posting for a valuation run.
func (s *valuationReconciliationService) GenerateReconciliationReport(ctx context.Context, valuationRunID string) (*ReconciliationReport, error) {
	// 1. Fetch the valuation run
	run, err := s.repo.FindByID(ctx, valuationRunID)
	if err != nil {
		return nil, fmt.Errorf("valuation run not found: %w", err)
	}

	report := &ReconciliationReport{
		ValuationRunID: run.ID,
		ValuationType:  string(run.ValuationType),
		CheckedAt:      run.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 2. If no journal posted yet, return early
	if run.JournalEntryID == nil {
		report.Status = ReconciliationStatusNoJournal
		report.ErrorMessage = "No journal entry posted yet"
		return report, nil
	}

	// 3. Get reconciliation tolerance
	tolerance, err := s.getReconciliationTolerance(ctx)
	if err != nil {
		tolerance = 0.01 // Default
	}

	// 4. Run valuation type-specific reconciliation
	switch run.ValuationType {
	case financeModels.ValuationTypeInventory:
		return s.reconcileInventory(ctx, report, run, tolerance)
	case financeModels.ValuationTypeFX:
		return s.reconcileFX(ctx, report, run, tolerance)
	case financeModels.ValuationTypeDepreciation:
		return s.reconcileDepreciation(ctx, report, run, tolerance)
	default:
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("Unknown valuation type: %s", run.ValuationType)
		return report, nil
	}
}

// reconcileInventory validates inventory GL posting vs subledger.
func (s *valuationReconciliationService) reconcileInventory(
	ctx context.Context,
	report *ReconciliationReport,
	run *financeModels.ValuationRun,
	tolerance float64,
) (*ReconciliationReport, error) {
	// Get the inventory asset COA account
	coaID, err := s.accountingEngine.ResolveCOAID(ctx, financeModels.SettingCOAInventoryAsset)
	if err != nil {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("Failed to resolve COA: %v", err)
		return report, nil
	}

	// Query GL balance for inventory asset account
	var glBalance float64
	if err := s.db.WithContext(ctx).
		Table("journal_lines jl").
		Select("COALESCE(SUM(jl.debit - jl.credit), 0)").
		Joins("JOIN journal_entries je ON je.id = jl.journal_entry_id").
		Where(
			"je.status = ? AND jl.chart_of_account_id = ? AND je.entry_date <= ? AND je.id != ?",
			financeModels.JournalStatusPosted,
			coaID,
			run.PeriodEnd,
			*run.JournalEntryID, // Exclude the valuation journal itself
		).
		Scan(&glBalance).Error; err != nil {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("GL query failed: %v", err)
		return report, nil
	}

	// Query subledger (inventory batches) balance
	var subledgerBalance float64
	if err := s.db.WithContext(ctx).
		Model(&inventoryModels.InventoryBatch{}).
		Select("COALESCE(SUM(current_quantity * cost_price), 0)").
		Where("is_active = ? AND DATE(created_at) <= ?", true, run.PeriodEnd.Format("2006-01-02")).
		Scan(&subledgerBalance).Error; err != nil {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("Subledger query failed: %v", err)
		return report, nil
	}

	delta := glBalance - subledgerBalance
	withinTolerance := math.Abs(delta) <= tolerance

	detail := ReconciliationDetail{
		Account:         "Inventory Asset (1120)",
		GLBalance:       glBalance,
		SubledgerTotal:  subledgerBalance,
		Delta:           delta,
		Status:          "inventory_asset",
		Tolerance:       tolerance,
		WithinTolerance: withinTolerance,
	}

	report.Details = append(report.Details, detail)
	report.OverallDelta = delta
	report.AllLinesMatched = withinTolerance

	if withinTolerance {
		report.Status = ReconciliationStatusMatched
	} else {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf(
			"Inventory reconciliation failed: GL=%.2f vs Subledger=%.2f (delta=%.2f, tolerance=%.2f)",
			glBalance, subledgerBalance, delta, tolerance,
		)
	}

	return report, nil
}

// reconcileFX validates FX GL posting vs subledger.
func (s *valuationReconciliationService) reconcileFX(
	ctx context.Context,
	report *ReconciliationReport,
	run *financeModels.ValuationRun,
	tolerance float64,
) (*ReconciliationReport, error) {
	// For FX, reconcile against the remeasurement account
	coaID, err := s.accountingEngine.ResolveCOAID(ctx, financeModels.SettingCOAFXRemeasurement)
	if err != nil {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("Failed to resolve FX COA: %v", err)
		return report, nil
	}

	// Query GL balance
	var glBalance float64
	if err := s.db.WithContext(ctx).
		Table("journal_lines jl").
		Select("COALESCE(SUM(jl.debit - jl.credit), 0)").
		Joins("JOIN journal_entries je ON je.id = jl.journal_entry_id").
		Where(
			"je.status = ? AND jl.chart_of_account_id = ? AND je.entry_date <= ? AND je.id != ?",
			financeModels.JournalStatusPosted,
			coaID,
			run.PeriodEnd,
			*run.JournalEntryID,
		).
		Scan(&glBalance).Error; err != nil {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("GL FX query failed: %v", err)
		return report, nil
	}

	// For FX, subledger is sum of outstanding AR/AP in foreign currency
	// This is a simplified check; in production you'd sum all FX-marked invoices
	var subledgerBalance float64 = run.TotalDelta // Use the delta as proxy for now

	delta := glBalance - subledgerBalance
	withinTolerance := math.Abs(delta) <= tolerance

	detail := ReconciliationDetail{
		Account:         "FX Remeasurement (1160)",
		GLBalance:       glBalance,
		SubledgerTotal:  subledgerBalance,
		Delta:           delta,
		Status:          "fx_remeasurement",
		Tolerance:       tolerance,
		WithinTolerance: withinTolerance,
	}

	report.Details = append(report.Details, detail)
	report.OverallDelta = delta
	report.AllLinesMatched = withinTolerance

	if withinTolerance {
		report.Status = ReconciliationStatusMatched
	} else {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf(
			"FX reconciliation mismatch: GL=%.2f vs Expected=%.2f (delta=%.2f)",
			glBalance, subledgerBalance, delta,
		)
	}

	return report, nil
}

// reconcileDepreciation validates depreciation GL posting.
func (s *valuationReconciliationService) reconcileDepreciation(
	ctx context.Context,
	report *ReconciliationReport,
	run *financeModels.ValuationRun,
	tolerance float64,
) (*ReconciliationReport, error) {
	// For depreciation, check accumulated depreciation account
	coaID, err := s.accountingEngine.ResolveCOAID(ctx, financeModels.SettingCOADepreciationAccumulated)
	if err != nil {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("Failed to resolve depreciation COA: %v", err)
		return report, nil
	}

	// Query GL balance
	var glBalance float64
	if err := s.db.WithContext(ctx).
		Table("journal_lines jl").
		Select("COALESCE(SUM(jl.credit - jl.debit), 0)"). // Accumulated depreciation is typically a credit
		Joins("JOIN journal_entries je ON je.id = jl.journal_entry_id").
		Where(
			"je.status = ? AND jl.chart_of_account_id = ? AND je.entry_date <= ? AND je.id != ?",
			financeModels.JournalStatusPosted,
			coaID,
			run.PeriodEnd,
			*run.JournalEntryID,
		).
		Scan(&glBalance).Error; err != nil {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf("GL depreciation query failed: %v", err)
		return report, nil
	}

	// For depreciation subledger, use the run's total delta as proxy
	subledgerBalance := math.Abs(run.TotalDelta)

	delta := glBalance - subledgerBalance
	withinTolerance := math.Abs(delta) <= tolerance

	detail := ReconciliationDetail{
		Account:         "Accumulated Depreciation (1140)",
		GLBalance:       glBalance,
		SubledgerTotal:  subledgerBalance,
		Delta:           delta,
		Status:          "depreciation_accumulated",
		Tolerance:       tolerance,
		WithinTolerance: withinTolerance,
	}

	report.Details = append(report.Details, detail)
	report.OverallDelta = delta
	report.AllLinesMatched = withinTolerance

	if withinTolerance {
		report.Status = ReconciliationStatusMatched
	} else {
		report.Status = ReconciliationStatusMismatched
		report.ErrorMessage = fmt.Sprintf(
			"Depreciation reconciliation mismatch: GL=%.2f vs Expected=%.2f (delta=%.2f)",
			glBalance, subledgerBalance, delta,
		)
	}

	return report, nil
}

// getReconciliationTolerance fetches the configured tolerance from settings.
func (s *valuationReconciliationService) getReconciliationTolerance(ctx context.Context) (float64, error) {
	value, err := s.settings.GetValue(ctx, "valuation.reconciliation_tolerance")
	if err != nil {
		return 0.01, nil // Default
	}

	var tolerance float64
	if _, err := fmt.Sscanf(value, "%f", &tolerance); err != nil {
		return 0.01, nil
	}

	if tolerance < 0 {
		return 0.01, nil
	}

	return tolerance, nil
}
