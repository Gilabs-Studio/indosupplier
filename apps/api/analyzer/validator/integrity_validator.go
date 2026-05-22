package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/analyzer"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

// IntegrityValidator checks data integrity: FK orphans, status anomalies
type IntegrityValidator struct{}

func NewIntegrityValidator() *IntegrityValidator { return &IntegrityValidator{} }

func (v *IntegrityValidator) Name() string { return "Data Integrity Validator" }

func (v *IntegrityValidator) Run(cfg *analyzer.Config) []analyzer.Finding {
	var findings []analyzer.Finding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findings = append(findings, v.checkOrphanPayments(ctx, cfg)...)
	findings = append(findings, v.checkStatusAnomalies(ctx, cfg)...)
	findings = append(findings, v.checkDuplicateReferences(ctx, cfg)...)

	return findings
}

// checkOrphanPayments finds payments referencing deleted/non-existent invoices
func (v *IntegrityValidator) checkOrphanPayments(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB
	var findings []analyzer.Finding

	checks := []struct {
		payTable   string
		invTable   string
		fkCol      string
		code       string
		moduleName string
	}{
		{"purchase_payments", "supplier_invoices", "supplier_invoice_id", "INT-001a", "purchase"},
		{"sales_payments", "customer_invoices", "customer_invoice_id", "INT-001b", "sales"},
	}

	for _, c := range checks {
		var count int64
		db.WithContext(ctx).Raw(fmt.Sprintf(`
			SELECT COUNT(*) FROM %s p
			WHERE p.deleted_at IS NULL
			  AND p.%s IS NOT NULL
			  AND NOT EXISTS (
			    SELECT 1 FROM %s i WHERE i.id = p.%s AND i.deleted_at IS NULL
			  )
		`, c.payTable, c.fkCol, c.invTable, c.fkCol)).Scan(&count)

		if count > 0 {
			findings = append(findings, analyzer.Finding{
				Code:           c.code,
				Severity:       analyzer.SeverityError,
				Module:         c.moduleName,
				Entity:         c.payTable,
				Message:        fmt.Sprintf("%d payments reference non-existent invoices in %s", count, c.invTable),
				Recommendation: "Investigate soft-deleted invoices linked to active payments.",
			})
		} else {
			findings = append(findings, analyzer.Finding{
				Code:     c.code,
				Severity: analyzer.SeverityPass,
				Module:   c.moduleName,
				Entity:   c.payTable,
				Message:  fmt.Sprintf("All %s have valid invoice references", c.payTable),
			})
		}
	}
	return findings
}

// checkStatusAnomalies finds records with impossible status combinations
func (v *IntegrityValidator) checkStatusAnomalies(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB
	var findings []analyzer.Finding

	// Supplier invoices marked approved but with zero amount
	var zeroAmountApproved int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM supplier_invoices
		WHERE status = 'approved' AND deleted_at IS NULL AND amount <= 0
	`).Scan(&zeroAmountApproved)

	if zeroAmountApproved > 0 {
		findings = append(findings, analyzer.Finding{
			Code:           "INT-002a",
			Severity:       analyzer.SeverityWarning,
			Module:         "purchase",
			Entity:         "supplier_invoice",
			Message:        fmt.Sprintf("%d approved supplier invoices with zero/negative amount", zeroAmountApproved),
			Recommendation: "Review zero-amount invoices — they may represent credit notes or data errors.",
		})
	}

	// Customer invoices marked approved but with zero amount
	var zeroAmountSales int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM customer_invoices
		WHERE status = 'approved' AND deleted_at IS NULL AND amount <= 0
	`).Scan(&zeroAmountSales)

	if zeroAmountSales > 0 {
		findings = append(findings, analyzer.Finding{
			Code:           "INT-002b",
			Severity:       analyzer.SeverityWarning,
			Module:         "sales",
			Entity:         "customer_invoice",
			Message:        fmt.Sprintf("%d approved customer invoices with zero/negative amount", zeroAmountSales),
			Recommendation: "Review zero-amount invoices.",
		})
	}

	if zeroAmountApproved == 0 && zeroAmountSales == 0 {
		findings = append(findings, analyzer.Finding{
			Code:     "INT-002",
			Severity: analyzer.SeverityPass,
			Module:   "finance",
			Entity:   "invoices",
			Message:  "No status anomalies detected",
		})
	}

	return findings
}

// checkDuplicateReferences finds journals with same reference_type+reference_id (potential double-posting)
func (v *IntegrityValidator) checkDuplicateReferences(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	type dupRef struct {
		ReferenceType string
		ReferenceID   string
		Count         int64
	}

	var results []dupRef
	db.WithContext(ctx).Raw(`
		SELECT reference_type, reference_id, COUNT(*) AS count
		FROM journal_entries
		WHERE reference_type IS NOT NULL AND reference_id IS NOT NULL
		  AND status = 'posted' AND deleted_at IS NULL
		  AND entry_date >= ? AND entry_date <= ?
		GROUP BY reference_type, reference_id
		HAVING COUNT(*) > 1
		LIMIT ?
	`, cfg.FromDateStr(), cfg.ToDateStr(), cfg.BatchLimit).Scan(&results)

	if len(results) == 0 {
		return []analyzer.Finding{{
			Code:     "INT-003",
			Severity: analyzer.SeverityPass,
			Module:   "finance",
			Entity:   "journal_entry",
			Message:  "No duplicate journal references found",
		}}
	}

	var findings []analyzer.Finding
	for _, r := range results {
		findings = append(findings, analyzer.Finding{
			Code:           "INT-003",
			Severity:       analyzer.SeverityWarning,
			Module:         "finance",
			Entity:         "journal_entry",
			Message:        fmt.Sprintf("Duplicate journal for %s/%s (count=%d)", r.ReferenceType, r.ReferenceID, r.Count),
			Evidence:       fmt.Sprintf("reference_type=%s reference_id=%s", r.ReferenceType, r.ReferenceID),
			Recommendation: "May indicate double-posting. Verify and reverse duplicates if needed.",
		})
	}
	return findings
}
