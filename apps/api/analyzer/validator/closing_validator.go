package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/analyzer"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

// ClosingValidator checks financial closing coverage and consistency
type ClosingValidator struct{}

func NewClosingValidator() *ClosingValidator { return &ClosingValidator{} }

func (v *ClosingValidator) Name() string { return "Financial Closing Validator" }

func (v *ClosingValidator) Run(cfg *analyzer.Config) []analyzer.Finding {
	if !cfg.ShouldRunModule("finance") {
		return nil
	}

	var findings []analyzer.Finding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findings = append(findings, v.checkClosingExists(ctx, cfg)...)
	findings = append(findings, v.checkClosingSnapshots(ctx, cfg)...)
	findings = append(findings, v.checkPeriodLock(ctx, cfg)...)

	return findings
}

func (v *ClosingValidator) checkClosingExists(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	var count int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM financial_closings
		WHERE deleted_at IS NULL AND status = 'approved'
	`).Scan(&count)

	if count == 0 {
		return []analyzer.Finding{{
			Code:           "CLS-001",
			Severity:       analyzer.SeverityWarning,
			Module:         "finance",
			Entity:         "financial_closing",
			Message:        "No approved financial closings found",
			Recommendation: "Run financial closing for completed accounting periods.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "CLS-001",
		Severity: analyzer.SeverityPass,
		Module:   "finance",
		Entity:   "financial_closing",
		Message:  fmt.Sprintf("%d approved financial closings found", count),
	}}
}

func (v *ClosingValidator) checkClosingSnapshots(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	// Find closings without snapshots
	var noSnapshot int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM financial_closings fc
		WHERE fc.deleted_at IS NULL AND fc.status = 'approved'
		  AND NOT EXISTS (
		    SELECT 1 FROM financial_closing_snapshots fcs
		    WHERE fcs.period_end_date = fc.period_end_date
		  )
	`).Scan(&noSnapshot)

	if noSnapshot > 0 {
		return []analyzer.Finding{{
			Code:           "CLS-002",
			Severity:       analyzer.SeverityError,
			Module:         "finance",
			Entity:         "financial_closing",
			Message:        fmt.Sprintf("%d approved closings without snapshots", noSnapshot),
			Recommendation: "Re-run closing snapshot generation. Snapshots are required for audit trail.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "CLS-002",
		Severity: analyzer.SeverityPass,
		Module:   "finance",
		Entity:   "financial_closing",
		Message:  "All approved closings have snapshots",
	}}
}

func (v *ClosingValidator) checkPeriodLock(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	// Find journals posted in a closed period
	var countViolation int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM journal_entries je
		WHERE je.status = 'posted' AND je.deleted_at IS NULL
		  AND EXISTS (
		    SELECT 1 FROM accounting_periods ap
		    WHERE ap.status = 'closed'
		      AND je.entry_date >= ap.start_date
		      AND je.entry_date <= ap.end_date
		  )
		  AND je.created_at > (
		    SELECT MAX(fc.approved_at) FROM financial_closings fc
		    WHERE fc.status = 'approved' AND fc.deleted_at IS NULL
		  )
	`).Scan(&countViolation)

	if countViolation > 0 {
		return []analyzer.Finding{{
			Code:           "CLS-003",
			Severity:       analyzer.SeverityCritical,
			Module:         "finance",
			Entity:         "journal_entry",
			Message:        fmt.Sprintf("%d journals posted in closed periods AFTER closing approval", countViolation),
			Recommendation: "Period lock enforcement may be bypassed. Check closing_guard middleware.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "CLS-003",
		Severity: analyzer.SeverityPass,
		Module:   "finance",
		Entity:   "journal_entry",
		Message:  "No period lock violations detected",
	}}
}
