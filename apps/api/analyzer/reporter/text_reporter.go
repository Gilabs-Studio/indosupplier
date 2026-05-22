package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gilabs/gims/api/analyzer"
)

// TextReporter writes findings to a plain text file
type TextReporter struct{}

func NewTextReporter() *TextReporter { return &TextReporter{} }

func (r *TextReporter) Report(findings []analyzer.Finding, cfg *analyzer.Config) error {
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	outPath := filepath.Join(cfg.OutputDir, "analysis-report.txt")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create txt report: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	sb.WriteString("=========================================================\n")
	sb.WriteString("               GIMS ERP Analyzer Report                  \n")
	sb.WriteString("=========================================================\n\n")

	sb.WriteString(fmt.Sprintf("Period: %s to %s\n", cfg.FromDate.Format("2006-01-02"), cfg.ToDate.Format("2006-01-02")))
	if len(cfg.Modules) > 0 {
		sb.WriteString(fmt.Sprintf("Modules: %v\n", cfg.Modules))
	} else {
		sb.WriteString("Modules: ALL\n")
	}
	sb.WriteString("\nFindings by Module:\n")
	sb.WriteString("---------------------------------------------------------\n")

	grouped := map[string][]analyzer.Finding{}
	for _, finding := range findings {
		grouped[finding.Module] = append(grouped[finding.Module], finding)
	}

	for module, items := range grouped {
		sb.WriteString(fmt.Sprintf("\n[%s] - %d findings\n", strings.ToUpper(module), len(items)))
		for _, item := range items {
			sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", item.Severity, item.Code, item.Message))
			if item.Evidence != "" {
				sb.WriteString(fmt.Sprintf("      Evidence: %s\n", item.Evidence))
			}
			if item.Recommendation != "" {
				sb.WriteString(fmt.Sprintf("      Action: %s\n", item.Recommendation))
			}
			sb.WriteString("\n")
		}
	}

	if _, err := f.WriteString(sb.String()); err != nil {
		return fmt.Errorf("failed to write txt report body: %w", err)
	}

	fmt.Printf("📄 Text report saved: %s\n", outPath)
	return nil
}
