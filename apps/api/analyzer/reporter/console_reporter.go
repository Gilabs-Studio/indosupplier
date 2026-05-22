package reporter

import (
	"fmt"
	"sort"

	"github.com/gilabs/gims/api/analyzer"
)

// ConsoleReporter prints findings to stdout grouped by module with severity icons
type ConsoleReporter struct{}

func NewConsoleReporter() *ConsoleReporter { return &ConsoleReporter{} }

func (r *ConsoleReporter) Report(findings []analyzer.Finding, cfg *analyzer.Config) error {
	// Group by module
	grouped := map[string][]analyzer.Finding{}
	for _, f := range findings {
		grouped[f.Module] = append(grouped[f.Module], f)
	}

	// Sort module names for deterministic output
	modules := make([]string, 0, len(grouped))
	for m := range grouped {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	for _, module := range modules {
		items := grouped[module]
		// Count per severity
		counts := map[analyzer.Severity]int{}
		for _, f := range items {
			counts[f.Severity]++
		}

		fmt.Printf("\n┌─ Module: %s (%d findings — ✅%d ❌%d ⚠️%d 🔴%d)\n",
			module, len(items),
			counts[analyzer.SeverityPass],
			counts[analyzer.SeverityError],
			counts[analyzer.SeverityWarning],
			counts[analyzer.SeverityCritical])

		for _, f := range items {
			fmt.Printf("│  %s\n", f.String())
		}
		fmt.Println("└──")
	}

	// ── Mandatory Failures Summary ───────────────────────────────────
	var mandatoryFails []analyzer.Finding
	for _, f := range findings {
		if f.IsFailure() && f.ScenarioID != "" {
			mandatoryFails = append(mandatoryFails, f)
		}
	}

	if len(mandatoryFails) > 0 {
		fmt.Println("\n🚨 MANDATORY SCENARIO FAILURES (must fix before release):")
		for i, f := range mandatoryFails {
			fmt.Printf("  %d. [%s] %s — %s\n", i+1, f.ScenarioID, f.Module, f.Message)
			if f.Recommendation != "" {
				fmt.Printf("     → Fix: %s\n", f.Recommendation)
			}
		}
	}

	return nil
}
