package simulator

import (
	"fmt"
	"os"
	"time"

	"github.com/gilabs/gims/api/analyzer"
)

type SimulationChecker struct {
	client   *ApiClient
	registry *ScenarioRegistry
}

func NewSimulationChecker() *SimulationChecker {
	return &SimulationChecker{
		client:   NewApiClient(),
		registry: NewScenarioRegistry(),
	}
}

func (s *SimulationChecker) Name() string {
	return "GIMS Enterprise QA Simulation (Sales + Purchase + Finance)"
}

func (s *SimulationChecker) Run(cfg *analyzer.Config) []analyzer.Finding {
	// Activation: run if module is simulator, qa, finance, sales, or purchase
	shouldRun := cfg.ShouldRunModule("qa") || cfg.ShouldRunModule("simulator") ||
		cfg.ShouldRunModule("finance") || cfg.ShouldRunModule("sales") || cfg.ShouldRunModule("purchase")

	findings := []analyzer.Finding{}
	if !shouldRun || cfg.Mode == analyzer.ModeValidate {
		return findings
	}

	fmt.Println("\n🚀 Starting Unified Enterprise QA Simulation (Sales → Purchase → Finance)...")

	// ── Security: Hard block mutation in production ───────────────────
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "production" {
		findings = append(findings, analyzer.Finding{
			Code:           "SIM-SEC-01",
			Severity:       analyzer.SeverityError,
			Module:         "core",
			Message:        "QA Simulator BLOCKED in production environment",
			Recommendation: "Run analyzer in staging/development environment only.",
		})
		return findings
	}

	runID := fmt.Sprintf("QA-%d", time.Now().Unix())

	// ══════════════════════════════════════════════════════════════════
	// Phase 1: Sales & Purchase E2E Flows
	// ══════════════════════════════════════════════════════════════════
	fmt.Println("  📦 Phase 1: Sales & Purchase Flows...")
	findings = append(findings, s.runSalesFlow(cfg, runID)...)
	findings = append(findings, s.runPurchaseFlow(cfg, runID)...)

	// ══════════════════════════════════════════════════════════════════
	// Phase 2: Finance Core Simulation
	// ══════════════════════════════════════════════════════════════════
	fmt.Println("  💰 Phase 2: Finance Core (Journal → GL → Reports)...")
	findings = append(findings, s.runJournalSimulation(cfg, runID)...)
	findings = append(findings, s.runGLSimulation(cfg, runID)...)
	findings = append(findings, s.runPaymentSimulation(cfg, runID)...)
	findings = append(findings, s.runAssetSimulation(cfg, runID)...)
	findings = append(findings, s.runValuationSimulation(cfg, runID)...)
	findings = append(findings, s.runReportSimulation(cfg, runID)...)
	findings = append(findings, s.runClosingSimulation(cfg, runID)...)

	// ══════════════════════════════════════════════════════════════════
	// Phase 3: Consistency, Negative & RBAC
	// ══════════════════════════════════════════════════════════════════
	fmt.Println("  🔒 Phase 3: Consistency, Negative Tests & RBAC...")
	findings = append(findings, s.runCrossMenuValidation(cfg, runID)...)
	findings = append(findings, s.runNegativeTests(cfg, runID)...)
	findings = append(findings, s.runRBACTests(cfg, runID)...)

	// ══════════════════════════════════════════════════════════════════
	// Phase 4: Coverage Gate & Registry Report
	// ══════════════════════════════════════════════════════════════════
	fmt.Println("  📊 Phase 4: Coverage Analysis...")

	// Collect findings from registry (skipped mandatory → ERROR)
	findings = append(findings, s.registry.GetFindings()...)

	// Coverage summary
	totalScenarios, implementedScenarios, passedScenarios := s.registry.GetCoverage()
	coveragePercent := float64(0)
	if totalScenarios > 0 {
		coveragePercent = float64(implementedScenarios) / float64(totalScenarios) * 100
	}

	findings = append(findings, analyzer.Finding{
		Code:     "SIM-COV-SUMMARY",
		Severity: analyzer.SeverityInfo,
		Module:   "simulator",
		Message: fmt.Sprintf("Coverage: %d/%d implemented (%.0f%%), %d/%d passed",
			implementedScenarios, totalScenarios, coveragePercent,
			passedScenarios, totalScenarios),
	})

	// Mandatory coverage gate
	if cfg.FailOnSkippedMandatory && implementedScenarios < totalScenarios {
		findings = append(findings, analyzer.Finding{
			Code:           "SIM-GATE-FAIL",
			Severity:       analyzer.SeverityCritical,
			Module:         "simulator",
			Message:        fmt.Sprintf("RELEASE GATE FAILED: Only %d/%d mandatory scenarios implemented", implementedScenarios, totalScenarios),
			Recommendation: "Implement all mandatory scenarios before release.",
		})
	}

	// Strict mode gate: if any mandatory scenario failed, emit critical
	if cfg.Strict && passedScenarios < totalScenarios {
		findings = append(findings, analyzer.Finding{
			Code:           "SIM-STRICT-FAIL",
			Severity:       analyzer.SeverityCritical,
			Module:         "simulator",
			Message:        fmt.Sprintf("STRICT MODE: %d/%d mandatory scenarios not passing", totalScenarios-passedScenarios, totalScenarios),
			Recommendation: "Fix all failing scenarios to achieve 100%% pass rate.",
		})
	}

	// Cleanup test data
	if cfg.Cleanup {
		s.cleanup(runID)
	}

	return findings
}

func (s *SimulationChecker) cleanup(runID string) {
	// API-based cleanup using runID prefix on generated data.
	// Draft journals/orders with QA prefix can be identified and deleted.
	// This is a best-effort operation and should not affect the test results.
}
