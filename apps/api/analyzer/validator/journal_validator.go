package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/analyzer"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

// JournalValidator checks journal entry integrity
type JournalValidator struct{}

func NewJournalValidator() *JournalValidator { return &JournalValidator{} }

func (v *JournalValidator) Name() string { return "Journal Integrity Validator" }

func (v *JournalValidator) Run(cfg *analyzer.Config) []analyzer.Finding {
	if !cfg.ShouldRunModule("finance") {
		return nil
	}

	var findings []analyzer.Finding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findings = append(findings, v.checkUnbalancedJournals(ctx, cfg)...)
	findings = append(findings, v.checkOrphanJournals(ctx, cfg)...)
	findings = append(findings, v.checkMissingJournalForTransactions(ctx, cfg)...)
	findings = append(findings, v.checkDraftJournals(ctx, cfg)...)

	return findings
}

// checkUnbalancedJournals finds posted journals where debit != credit
func (v *JournalValidator) checkUnbalancedJournals(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	type unbalanced struct {
		ID          string
		Description string
		DebitSum    float64
		CreditSum   float64
	}

	var results []unbalanced
	db.WithContext(ctx).Raw(`
		SELECT je.id, je.description,
		       COALESCE(SUM(jl.debit), 0) AS debit_sum,
		       COALESCE(SUM(jl.credit), 0) AS credit_sum
		FROM journal_entries je
		JOIN journal_lines jl ON jl.journal_entry_id = je.id
		WHERE je.status = 'posted' AND je.deleted_at IS NULL
		  AND je.entry_date >= ? AND je.entry_date <= ?
		GROUP BY je.id, je.description
		HAVING ABS(SUM(jl.debit) - SUM(jl.credit)) > 0.01
		LIMIT ?
	`, cfg.FromDateStr(), cfg.ToDateStr(), cfg.BatchLimit).Scan(&results)

	if len(results) == 0 {
		return []analyzer.Finding{{
			Code:     "JRN-001",
			Severity: analyzer.SeverityPass,
			Module:   "finance",
			Entity:   "journal_entry",
			Message:  "All posted journals are balanced (debit == credit)",
		}}
	}

	var findings []analyzer.Finding
	for _, r := range results {
		findings = append(findings, analyzer.Finding{
			Code:           "JRN-001",
			Severity:       analyzer.SeverityCritical,
			Module:         "finance",
			Entity:         "journal_entry",
			Message:        fmt.Sprintf("Unbalanced journal: %s (Debit=%.2f Credit=%.2f)", r.Description, r.DebitSum, r.CreditSum),
			Evidence:       fmt.Sprintf("journal_entries.id = %s", r.ID),
			Recommendation: "Correct the journal lines or reverse and re-post.",
		})
	}
	return findings
}

// checkOrphanJournals finds journals with reference_type but no matching source record
func (v *JournalValidator) checkOrphanJournals(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	// Check journals referencing supplier_invoices that don't exist
	refChecks := []struct {
		refType string
		table   string
		code    string
	}{
		{"SUPPLIER_INVOICE", "supplier_invoices", "JRN-002a"},
		{"SALES_INVOICE", "customer_invoices", "JRN-002b"},
		{"PURCHASE_PAYMENT", "purchase_payments", "JRN-002c"},
		{"SALES_PAYMENT", "sales_payments", "JRN-002d"},
		{"GOODS_RECEIPT", "goods_receipts", "JRN-002e"},
	}

	var findings []analyzer.Finding
	for _, rc := range refChecks {
		var count int64
		db.WithContext(ctx).Raw(fmt.Sprintf(`
			SELECT COUNT(*) FROM journal_entries je
			WHERE je.reference_type = ? AND je.deleted_at IS NULL
			  AND je.reference_id IS NOT NULL
			  AND NOT EXISTS (SELECT 1 FROM %s t WHERE t.id::text = je.reference_id AND t.deleted_at IS NULL)
			  AND je.entry_date >= ? AND je.entry_date <= ?
		`, rc.table), rc.refType, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&count)

		if count > 0 {
			findings = append(findings, analyzer.Finding{
				Code:           rc.code,
				Severity:       analyzer.SeverityError,
				Module:         "finance",
				Entity:         "journal_entry",
				Message:        fmt.Sprintf("%d orphan journals referencing non-existent %s records", count, rc.refType),
				Evidence:       fmt.Sprintf("reference_type='%s' with missing source in %s", rc.refType, rc.table),
				Recommendation: "Investigate deleted source records. Consider reversing orphan journals.",
			})
		} else {
			findings = append(findings, analyzer.Finding{
				Code:     rc.code,
				Severity: analyzer.SeverityPass,
				Module:   "finance",
				Entity:   "journal_entry",
				Message:  fmt.Sprintf("No orphan journals for %s", rc.refType),
			})
		}
	}
	return findings
}

// checkMissingJournalForTransactions finds posted transactions without journal entries
func (v *JournalValidator) checkMissingJournalForTransactions(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	txChecks := []struct {
		table      string
		refType    string
		statusCol  string
		statusVal  string
		dateCol    string
		code       string
		entityName string
	}{
		{"supplier_invoices", "SUPPLIER_INVOICE", "status", "approved", "invoice_date", "JRN-003a", "supplier_invoice"},
		{"customer_invoices", "SALES_INVOICE", "status", "approved", "invoice_date", "JRN-003b", "customer_invoice"},
		{"purchase_payments", "PURCHASE_PAYMENT", "status", "approved", "payment_date", "JRN-003c", "purchase_payment"},
		{"sales_payments", "SALES_PAYMENT", "status", "approved", "payment_date", "JRN-003d", "sales_payment"},
	}

	var findings []analyzer.Finding
	for _, tc := range txChecks {
		var count int64
		db.WithContext(ctx).Raw(fmt.Sprintf(`
			SELECT COUNT(*) FROM %s t
			WHERE t.%s = ? AND t.deleted_at IS NULL
			  AND t.%s >= ? AND t.%s <= ?
			  AND NOT EXISTS (
			    SELECT 1 FROM journal_entries je
			    WHERE je.reference_type = ? AND je.reference_id = t.id::text
			      AND je.deleted_at IS NULL
			  )
		`, tc.table, tc.statusCol, tc.dateCol, tc.dateCol),
			tc.statusVal, cfg.FromDateStr(), cfg.ToDateStr(), tc.refType).Scan(&count)

		if count > 0 {
			findings = append(findings, analyzer.Finding{
				Code:           tc.code,
				Severity:       analyzer.SeverityCritical,
				Module:         "finance",
				Entity:         tc.entityName,
				Message:        fmt.Sprintf("%d approved %s without journal entry", count, tc.entityName),
				Evidence:       fmt.Sprintf("table=%s status=%s with no matching journal_entries.reference_type=%s", tc.table, tc.statusVal, tc.refType),
				Recommendation: "Re-trigger journal generation for these transactions or investigate the approval flow.",
			})
		} else {
			findings = append(findings, analyzer.Finding{
				Code:     tc.code,
				Severity: analyzer.SeverityPass,
				Module:   "finance",
				Entity:   tc.entityName,
				Message:  fmt.Sprintf("All approved %s have journal entries", tc.entityName),
			})
		}
	}
	return findings
}

// checkDraftJournals counts draft journals that may be stuck
func (v *JournalValidator) checkDraftJournals(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB

	var count int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM journal_entries
		WHERE status = 'draft' AND deleted_at IS NULL
		  AND entry_date >= ? AND entry_date <= ?
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&count)

	if count > 10 {
		return []analyzer.Finding{{
			Code:           "JRN-004",
			Severity:       analyzer.SeverityWarning,
			Module:         "finance",
			Entity:         "journal_entry",
			Message:        fmt.Sprintf("%d draft journals found — may be stuck/unprocessed", count),
			Recommendation: "Review draft journals and post or delete as appropriate.",
		}}
	}

	return []analyzer.Finding{{
		Code:     "JRN-004",
		Severity: analyzer.SeverityPass,
		Module:   "finance",
		Entity:   "journal_entry",
		Message:  fmt.Sprintf("%d draft journals (within normal range)", count),
	}}
}
