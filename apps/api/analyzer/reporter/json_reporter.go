package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gilabs/gims/api/analyzer"
)

// JSONReporter writes findings to a JSON file
type JSONReporter struct{}

func NewJSONReporter() *JSONReporter { return &JSONReporter{} }

type jsonReport struct {
	GeneratedAt string            `json:"generated_at"`
	Period      jsonPeriod        `json:"period"`
	Summary     map[string]int    `json:"summary"`
	Findings    []analyzer.Finding `json:"findings"`
}

type jsonPeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (r *JSONReporter) Report(findings []analyzer.Finding, cfg *analyzer.Config) error {
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	counts := map[string]int{}
	for _, f := range findings {
		counts[string(f.Severity)]++
	}

	report := jsonReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Period: jsonPeriod{
			From: cfg.FromDate.Format("2006-01-02"),
			To:   cfg.ToDate.Format("2006-01-02"),
		},
		Summary:  counts,
		Findings: findings,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	outPath := filepath.Join(cfg.OutputDir, "analysis-report.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	fmt.Printf("\n📄 JSON report saved: %s\n", outPath)
	return nil
}
