package analyzer

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

// Engine orchestrates all checks and reporters
type Engine struct {
	checkers  []Checker
	reporters []Reporter
}

// Reporter writes findings to a destination
type Reporter interface {
	Report(findings []Finding, cfg *Config) error
}

// NewEngine creates a new analyzer engine
func NewEngine() *Engine {
	return &Engine{}
}

// RegisterChecker adds a checker to the engine
func (e *Engine) RegisterChecker(c Checker) {
	e.checkers = append(e.checkers, c)
}

// RegisterReporter adds a reporter to the engine
func (e *Engine) RegisterReporter(r Reporter) {
	e.reporters = append(e.reporters, r)
}

// Run executes all registered checkers and passes findings to reporters
func (e *Engine) Run(cfg *Config) error {
	// Production safety guard
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "production" {
		log.Println("⚠️  Production environment detected — running in READ-ONLY mode")
	}

	if cfg.DryRun {
		fmt.Println("\n=== DRY RUN — would execute the following checks ===")
		for _, c := range e.checkers {
			fmt.Printf("  • %s\n", c.Name())
		}
		return nil
	}

	start := time.Now()
	var allFindings []Finding

	fmt.Println("\n╔══════════════════════════════════════════════════════╗")
	fmt.Println("║      GIMS ERP QA Engine — Production Audit Report   ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Printf("  Period : %s → %s\n", cfg.FromDate.Format("2006-01-02"), cfg.ToDate.Format("2006-01-02"))
	if len(cfg.Modules) > 0 {
		fmt.Printf("  Modules: %v\n", cfg.Modules)
	} else {
		fmt.Printf("  Modules: ALL\n")
	}
	fmt.Printf("  Mode   : %s\n", cfg.Mode)
	fmt.Printf("  Strict : %v\n\n", cfg.Strict)

	for _, checker := range e.checkers {
		fmt.Printf("▶ Running: %s\n", checker.Name())
		findings := checker.Run(cfg)

		// In strict mode, promote warnings to errors
		if cfg.Strict {
			for i := range findings {
				if findings[i].Severity == SeverityWarning {
					findings[i].Severity = SeverityError
					findings[i].Message = "[STRICT] " + findings[i].Message
				}
			}
		}

		allFindings = append(allFindings, findings...)
	}

	elapsed := time.Since(start)

	// ── Severity Summary ─────────────────────────────────────────────
	counts := map[Severity]int{}
	for _, f := range allFindings {
		counts[f.Severity]++
	}

	// ── Module Summary ───────────────────────────────────────────────
	moduleCounts := map[string]map[Severity]int{}
	for _, f := range allFindings {
		if moduleCounts[f.Module] == nil {
			moduleCounts[f.Module] = map[Severity]int{}
		}
		moduleCounts[f.Module][f.Severity]++
	}

	fmt.Println("\n══════════════════════════════════════════════════════")
	fmt.Printf("  Summary (%s)\n", elapsed.Round(time.Millisecond))
	fmt.Println("══════════════════════════════════════════════════════")
	fmt.Printf("  🔴 CRITICAL : %d\n", counts[SeverityCritical])
	fmt.Printf("  ❌ ERROR    : %d\n", counts[SeverityError])
	fmt.Printf("  ⚠️  WARNING  : %d\n", counts[SeverityWarning])
	fmt.Printf("  ℹ️  INFO     : %d\n", counts[SeverityInfo])
	fmt.Printf("  ✅ PASS     : %d\n", counts[SeverityPass])
	fmt.Printf("  ⏭️  SKIPPED  : %d\n", counts[SeveritySkipped])
	fmt.Printf("  Total      : %d findings\n", len(allFindings))
	fmt.Println("──────────────────────────────────────────────────────")

	// Per-module breakdown
	modules := make([]string, 0, len(moduleCounts))
	for m := range moduleCounts {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	fmt.Println("  Per Module:")
	for _, m := range modules {
		mc := moduleCounts[m]
		fmt.Printf("    %-12s  ✅ %d  ❌ %d  ⚠️ %d  🔴 %d\n", m, mc[SeverityPass], mc[SeverityError], mc[SeverityWarning], mc[SeverityCritical])
	}
	fmt.Println("──────────────────────────────────────────────────────")

	// Run reporters
	for _, reporter := range e.reporters {
		if err := reporter.Report(allFindings, cfg); err != nil {
			log.Printf("Reporter error: %v", err)
		}
	}

	// ── Exit Code Decision ───────────────────────────────────────────
	hasFailures := counts[SeverityCritical] > 0 || counts[SeverityError] > 0

	// In strict mode, any warning or skipped is also a failure
	if cfg.Strict && (counts[SeverityWarning] > 0 || counts[SeveritySkipped] > 0) {
		hasFailures = true
	}

	if hasFailures {
		return fmt.Errorf("analysis found %d critical, %d error, %d warning, %d skipped findings",
			counts[SeverityCritical], counts[SeverityError], counts[SeverityWarning], counts[SeveritySkipped])
	}

	return nil
}
