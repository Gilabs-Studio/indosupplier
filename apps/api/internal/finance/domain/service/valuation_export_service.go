package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/finance/data/repositories"
)

// ValuationExportFormat defines supported export formats
type ValuationExportFormat string

const (
	ExportFormatCSV ValuationExportFormat = "csv"
	ExportFormatPDF ValuationExportFormat = "pdf"
)

// ExportedFile represents an exported document
type ExportedFile struct {
	FileName    string
	ContentType string
	Content     []byte
}

// ValuationExportService generates audit-ready export files for valuation runs.
type ValuationExportService interface {
	// ExportAsCSV generates a CSV file with valuation details for spreadsheet analysis.
	ExportAsCSV(ctx context.Context, valuationRunID string) (*ExportedFile, error)

	// ExportAsPDF generates a PDF report with full valuation details for auditors.
	ExportAsPDF(ctx context.Context, valuationRunID string) (*ExportedFile, error)
}

type valuationExportService struct {
	repo repositories.ValuationRunRepository
}

// NewValuationExportService creates a new export service.
func NewValuationExportService(repo repositories.ValuationRunRepository) ValuationExportService {
	return &valuationExportService{repo: repo}
}

// ExportAsCSV generates a CSV file for the valuation run.
func (s *valuationExportService) ExportAsCSV(ctx context.Context, valuationRunID string) (*ExportedFile, error) {
	// 1. Fetch the valuation run with details
	run, err := s.repo.FindByID(ctx, valuationRunID)
	if err != nil {
		return nil, fmt.Errorf("valuation run not found: %w", err)
	}

	var buf bytes.Buffer

	// 2. Write header section
	buf.WriteString("VALUATION RUN EXPORT\n")
	buf.WriteString(fmt.Sprintf("Export Date,%s\n", time.Now().Format("2006-01-02 15:04:05")))
	buf.WriteString("\n")

	// 3. Write run details
	buf.WriteString("RUN DETAILS\n")
	buf.WriteString("Field,Value\n")
	buf.WriteString(fmt.Sprintf("Run ID,%s\n", run.ID))
	buf.WriteString(fmt.Sprintf("Valuation Type,%s\n", run.ValuationType))
	buf.WriteString(fmt.Sprintf("Period Start,%s\n", run.PeriodStart.Format("2006-01-02")))
	buf.WriteString(fmt.Sprintf("Period End,%s\n", run.PeriodEnd.Format("2006-01-02")))
	buf.WriteString(fmt.Sprintf("Status,%s\n", run.Status))
	buf.WriteString(fmt.Sprintf("Total Delta,%.2f\n", run.TotalDelta))
	buf.WriteString(fmt.Sprintf("Total Debit,%.2f\n", run.TotalDebit))
	buf.WriteString(fmt.Sprintf("Total Credit,%.2f\n", run.TotalCredit))

	if run.IsLocked {
		buf.WriteString(fmt.Sprintf("Locked At,%s\n", run.LockedAt.Format("2006-01-02 15:04:05")))
	}

	if run.ApprovedBy != nil {
		buf.WriteString(fmt.Sprintf("Approved By,%s\n", *run.ApprovedBy))
		if run.ApprovedAt != nil {
			buf.WriteString(fmt.Sprintf("Approved At,%s\n", run.ApprovedAt.Format("2006-01-02 15:04:05")))
		}
	}

	buf.WriteString("\n")

	// 4. Write valuation items (details)
	if len(run.Details) > 0 {
		buf.WriteString("VALUATION ITEMS\n")
		buf.WriteString("Reference ID,Product ID,Book Value,Actual Value,Delta,Direction,Cost Price,Currency,Exchange Rate\n")

		for _, detail := range run.Details {
			productID := ""
			if detail.ProductID != nil {
				productID = *detail.ProductID
			}

			costPrice := ""
			if detail.CostPriceSnapshot != nil {
				costPrice = fmt.Sprintf("%.2f", *detail.CostPriceSnapshot)
			}

			currency := ""
			if detail.CurrencyCodeSnapshot != nil {
				currency = *detail.CurrencyCodeSnapshot
			}

			exchangeRate := ""
			if detail.ExchangeRateSnapshot != nil {
				exchangeRate = fmt.Sprintf("%.4f", *detail.ExchangeRateSnapshot)
			}

			buf.WriteString(fmt.Sprintf(
				"\"%s\",\"%s\",%.2f,%.2f,%.2f,%s,%s,%s,%s\n",
				detail.ReferenceID,
				productID,
				detail.BookValue,
				detail.ActualValue,
				detail.Delta,
				detail.Direction,
				costPrice,
				currency,
				exchangeRate,
			))
		}
		buf.WriteString("\n")
	}

	// 5. Write audit trail
	buf.WriteString("AUDIT TRAIL\n")
	buf.WriteString("Field,Value\n")
	buf.WriteString(fmt.Sprintf("Created By,%v\n", run.CreatedBy))
	buf.WriteString(fmt.Sprintf("Created At,%s\n", run.CreatedAt.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("Updated At,%s\n", run.UpdatedAt.Format("2006-01-02 15:04:05")))

	if run.ApprovalNotes != "" {
		buf.WriteString(fmt.Sprintf("Approval Notes,\"%s\"\n", escapeCSV(run.ApprovalNotes)))
	}

	fileName := fmt.Sprintf("valuation_%s_%s.csv", run.ValuationType, time.Now().Format("20060102_150405"))

	return &ExportedFile{
		FileName:    fileName,
		ContentType: "text/csv",
		Content:     buf.Bytes(),
	}, nil
}

// ExportAsPDF generates a PDF report for the valuation run.
// NOTE: For MVP, this generates a detailed text report that can be printed as PDF.
// In production, integrate `github.com/johnfercher/maroto` or `github.com/go-pdf/fpdf` for native PDF generation.
func (s *valuationExportService) ExportAsPDF(ctx context.Context, valuationRunID string) (*ExportedFile, error) {
	// 1. Fetch the valuation run
	run, err := s.repo.FindByID(ctx, valuationRunID)
	if err != nil {
		return nil, fmt.Errorf("valuation run not found: %w", err)
	}

	var buf bytes.Buffer

	// 2. Write PDF-compatible text report (can be printed to PDF)
	buf.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                    VALUATION RUN REPORT                        ║\n")
	buf.WriteString("║                     AUDIT READY EXPORT                         ║\n")
	buf.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	buf.WriteString(fmt.Sprintf("Export Date: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	buf.WriteString("────────────────────────────────────────────────────────────────\n\n")

	// 3. Run details section
	buf.WriteString("RUN DETAILS\n")
	buf.WriteString("────────────────────────────────────────────────────────────────\n")
	buf.WriteString(fmt.Sprintf("Run ID:           %s\n", run.ID))
	buf.WriteString(fmt.Sprintf("Valuation Type:   %s\n", run.ValuationType))
	buf.WriteString(fmt.Sprintf("Status:           %s\n", run.Status))
	buf.WriteString(fmt.Sprintf("Period:           %s to %s\n",
		run.PeriodStart.Format("2006-01-02"),
		run.PeriodEnd.Format("2006-01-02"),
	))
	buf.WriteString("\n")

	// 4. Financial summary
	buf.WriteString("FINANCIAL SUMMARY\n")
	buf.WriteString("────────────────────────────────────────────────────────────────\n")
	buf.WriteString(fmt.Sprintf("Total Delta:      %.2f\n", run.TotalDelta))
	buf.WriteString(fmt.Sprintf("Total Debit:      %.2f\n", run.TotalDebit))
	buf.WriteString(fmt.Sprintf("Total Credit:     %.2f\n", run.TotalCredit))
	buf.WriteString(fmt.Sprintf("Balanced:         %v\n", run.TotalDebit == run.TotalCredit))
	buf.WriteString("\n")

	// 5. Posting status
	buf.WriteString("POSTING STATUS\n")
	buf.WriteString("────────────────────────────────────────────────────────────────\n")
	buf.WriteString(fmt.Sprintf("Posted:           %v\n", run.JournalEntryID != nil))
	if run.JournalEntryID != nil {
		buf.WriteString(fmt.Sprintf("Journal Entry ID:  %s\n", *run.JournalEntryID))
	}
	buf.WriteString(fmt.Sprintf("Period Locked:    %v\n", run.IsLocked))
	if run.IsLocked && run.LockedAt != nil {
		buf.WriteString(fmt.Sprintf("Locked At:        %s\n", run.LockedAt.Format("2006-01-02 15:04:05")))
	}
	buf.WriteString("\n")

	// 6. Approval information
	buf.WriteString("APPROVAL INFORMATION\n")
	buf.WriteString("────────────────────────────────────────────────────────────────\n")
	if run.ApprovedBy != nil {
		buf.WriteString(fmt.Sprintf("Approved By:      %s\n", *run.ApprovedBy))
		if run.ApprovedAt != nil {
			buf.WriteString(fmt.Sprintf("Approved At:      %s\n", run.ApprovedAt.Format("2006-01-02 15:04:05")))
		}
	} else {
		buf.WriteString("Approved By:      (pending)\n")
	}
	if run.ApprovalNotes != "" {
		buf.WriteString(fmt.Sprintf("Notes:\n%s\n", run.ApprovalNotes))
	}
	buf.WriteString("\n")

	// 7. Valuation items
	if len(run.Details) > 0 {
		buf.WriteString("VALUATION ITEMS DETAIL\n")
		buf.WriteString("────────────────────────────────────────────────────────────────\n")
		buf.WriteString(fmt.Sprintf("%-40s | %-12s | %-10s | %s\n", "Reference", "Book Value", "Actual Value", "Delta"))
		buf.WriteString(strings.Repeat("─", 80) + "\n")

		for _, detail := range run.Details {
			buf.WriteString(fmt.Sprintf("%-40s | %.2f %10s | %.2f %8s | %.2f\n",
				truncate(detail.ReferenceID, 39),
				detail.BookValue, "",
				detail.ActualValue, "",
				detail.Delta,
			))
		}
		buf.WriteString("\n")
	}

	// 8. Audit trail
	buf.WriteString("AUDIT TRAIL\n")
	buf.WriteString("────────────────────────────────────────────────────────────────\n")
	buf.WriteString(fmt.Sprintf("Created By:       %v\n", run.CreatedBy))
	buf.WriteString(fmt.Sprintf("Created At:       %s\n", run.CreatedAt.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("Updated At:       %s\n", run.UpdatedAt.Format("2006-01-02 15:04:05")))
	if run.CompletedAt != nil {
		buf.WriteString(fmt.Sprintf("Completed At:     %s\n", run.CompletedAt.Format("2006-01-02 15:04:05")))
	}

	buf.WriteString("\n────────────────────────────────────────────────────────────────\n")
	buf.WriteString("This is an audit-ready export generated from the GIMS Platform.\n")
	buf.WriteString("For compliance records, this document should be retained.\n")

	fileName := fmt.Sprintf("valuation_%s_%s.txt", run.ValuationType, time.Now().Format("20060102_150405"))

	return &ExportedFile{
		FileName:    fileName,
		ContentType: "text/plain",
		Content:     buf.Bytes(),
	}, nil
}

// escapeCSV escapes CSV fields that contain commas or quotes
func escapeCSV(s string) string {
	if strings.Contains(s, ",") || strings.Contains(s, "\"") || strings.Contains(s, "\n") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}

// truncate truncates a string to max length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
