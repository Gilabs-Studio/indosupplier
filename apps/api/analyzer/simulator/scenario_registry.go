package simulator

import (
	"fmt"

	"github.com/gilabs/gims/api/analyzer"
)

type ScenarioType string

const (
	ScenarioMandatory ScenarioType = "MANDATORY"
	ScenarioOptional  ScenarioType = "OPTIONAL"
)

type ScenarioRecord struct {
	ID          string
	Module      string
	Name        string
	Type        ScenarioType
	Description string
	Status      analyzer.Severity
	Findings    []analyzer.Finding
}

// MandatoryScenarios is the single source of truth for all mandatory test scenarios.
// Every scenario here MUST be implemented and MUST pass for the analyzer to return exit code 0.
var MandatoryScenarios = map[string]ScenarioRecord{
	// ── Finance: Journal Entries ──────────────────────────────────────
	"FIN-SIM-JE-01": {ID: "FIN-SIM-JE-01", Module: "finance", Name: "Journal Creation", Type: ScenarioMandatory,
		Description: "Create a valid manual journal entry with dynamic COA IDs"},
	"FIN-SIM-JE-02": {ID: "FIN-SIM-JE-02", Module: "finance", Name: "Journal Posting", Type: ScenarioMandatory,
		Description: "Post a draft journal entry and verify status transition"},
	"FIN-SIM-JE-03": {ID: "FIN-SIM-JE-03", Module: "finance", Name: "Journal Reversal", Type: ScenarioMandatory,
		Description: "Reverse a posted journal entry and validate reversal reference"},

	// ── Finance: General Ledger ──────────────────────────────────────
	"FIN-SIM-GL-01": {ID: "FIN-SIM-GL-01", Module: "finance", Name: "GL Report Load", Type: ScenarioMandatory,
		Description: "Load General Ledger report with date filter"},

	// ── Finance: Reports ─────────────────────────────────────────────
	"FIN-SIM-REP-PL": {ID: "FIN-SIM-REP-PL", Module: "finance", Name: "P&L Consistency", Type: ScenarioMandatory,
		Description: "Verify P&L endpoint returns valid data and Revenue-Expense=NetProfit"},
	"FIN-SIM-REP-BS": {ID: "FIN-SIM-REP-BS", Module: "finance", Name: "Balance Sheet Equilibrium", Type: ScenarioMandatory,
		Description: "Verify Assets = Liabilities + Equity"},
	"FIN-SIM-REP-TB": {ID: "FIN-SIM-REP-TB", Module: "finance", Name: "Trial Balance Load", Type: ScenarioMandatory,
		Description: "Verify Trial Balance report loads successfully"},

	// ── Finance: Asset Budget ────────────────────────────────────────
	"FIN-SIM-AST-01": {ID: "FIN-SIM-AST-01", Module: "finance", Name: "Asset Budget List", Type: ScenarioMandatory,
		Description: "Verify Asset Budget listing endpoint is accessible"},

	// ── Finance: Payment ─────────────────────────────────────────────
	"FIN-SIM-PAY-01": {ID: "FIN-SIM-PAY-01", Module: "finance", Name: "Payment Lifecycle", Type: ScenarioMandatory,
		Description: "Create payment with valid COA allocations and verify response"},

	// ── Finance: Valuation ───────────────────────────────────────────
	"FIN-SIM-VAL-01": {ID: "FIN-SIM-VAL-01", Module: "finance", Name: "Valuation Run", Type: ScenarioMandatory,
		Description: "Trigger valuation run and verify journal is created"},

	// ── Finance: Closing ─────────────────────────────────────────────
	"FIN-SIM-CL-01": {ID: "FIN-SIM-CL-01", Module: "finance", Name: "Financial Closing Check", Type: ScenarioMandatory,
		Description: "Verify active accounting period endpoint responds"},

	// ── Negative Scenarios ───────────────────────────────────────────
	"FIN-NEG-UB-01": {ID: "FIN-NEG-UB-01", Module: "finance", Name: "Unbalanced Journal Rejection", Type: ScenarioMandatory,
		Description: "System must reject unbalanced journal entries"},
	"FIN-NEG-CL-01": {ID: "FIN-NEG-CL-01", Module: "finance", Name: "Closed Period Mutation Block", Type: ScenarioMandatory,
		Description: "System must reject mutations in closed/old periods"},
}

type ScenarioRegistry struct {
	results map[string]*ScenarioRecord
}

func NewScenarioRegistry() *ScenarioRegistry {
	results := make(map[string]*ScenarioRecord)
	for id, s := range MandatoryScenarios {
		sc := s // copy
		sc.Status = analyzer.SeveritySkipped
		results[id] = &sc
	}
	return &ScenarioRegistry{results: results}
}

func (r *ScenarioRegistry) MarkPass(id string, findings ...analyzer.Finding) {
	if s, ok := r.results[id]; ok {
		s.Status = analyzer.SeverityPass
		s.Findings = findings
	}
}

func (r *ScenarioRegistry) MarkFail(id string, findings ...analyzer.Finding) {
	if s, ok := r.results[id]; ok {
		s.Status = analyzer.SeverityError
		s.Findings = findings
	}
}

// GetCoverage returns total mandatory, implemented (non-skipped), and passed counts.
func (r *ScenarioRegistry) GetCoverage() (total int, implemented int, passed int) {
	total = len(MandatoryScenarios)
	for _, s := range r.results {
		if s.Status == analyzer.SeverityPass {
			passed++
			implemented++
		} else if s.Status == analyzer.SeverityError || s.Status == analyzer.SeverityWarning || s.Status == analyzer.SeverityCritical {
			implemented++
		}
		// SeveritySkipped = not implemented
	}
	return
}

// GetFindings produces actionable findings from the registry.
// Mandatory scenarios that are still SKIPPED are promoted to ERROR with a clear recommendation.
func (r *ScenarioRegistry) GetFindings() []analyzer.Finding {
	all := []analyzer.Finding{}
	for _, s := range r.results {
		if s.Status == analyzer.SeveritySkipped && s.Type == ScenarioMandatory {
			all = append(all, analyzer.Finding{
				Code:           "SIM-MANDATORY-SKIPPED",
				ScenarioID:     s.ID,
				Severity:       analyzer.SeverityError,
				Module:         s.Module,
				Message:        fmt.Sprintf("Mandatory scenario NOT IMPLEMENTED: %s — %s", s.Name, s.Description),
				Recommendation: fmt.Sprintf("Implement scenario %s in the simulator to achieve 100%% coverage.", s.ID),
			})
		}
		all = append(all, s.Findings...)
	}
	return all
}

// GetModuleCoverage returns per-module coverage stats.
func (r *ScenarioRegistry) GetModuleCoverage() map[string][3]int {
	moduleCov := make(map[string][3]int) // [total, implemented, passed]
	for _, s := range r.results {
		m := s.Module
		c := moduleCov[m]
		c[0]++ // total
		if s.Status == analyzer.SeverityPass {
			c[1]++ // implemented
			c[2]++ // passed
		} else if s.Status != analyzer.SeveritySkipped {
			c[1]++ // implemented but failed
		}
		moduleCov[m] = c
	}
	return moduleCov
}
