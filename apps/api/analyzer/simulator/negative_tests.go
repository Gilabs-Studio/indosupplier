package simulator

import (
	"fmt"
	"net/http"

	"github.com/gilabs/gims/api/analyzer"
)

func (s *SimulationChecker) runNegativeTests(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}
	
	// FIN-NEG-UB-01: Unbalanced Journal
	sID := "FIN-NEG-UB-01"
	unbalancedJE := map[string]interface{}{
		"entry_date": "2026-03-24",
		"description": "Unbalanced Test Journal " + runID,
		"lines": []map[string]interface{}{
			{"chart_of_account_id": s.client.GetCOAID("11100"), "debit": 1000, "credit": 0},
			{"chart_of_account_id": s.client.GetCOAID("11100"), "debit": 0, "credit": 500}, 
		},
	}
	
	status, _, _ := s.client.Request("POST", "/finance/journal-entries", unbalancedJE, RoleAdmin)
	if status == http.StatusBadRequest || status == http.StatusUnprocessableEntity {
		s.registry.MarkPass(sID, analyzer.Finding{Code: "SIM-NEG-01", Severity: analyzer.SeverityPass, Module: "finance", Message: "Blocked unbalanced journal creation"})
	} else {
		f := analyzer.Finding{Code: "SIM-NEG-01", Severity: analyzer.SeverityError, Module: "finance", Message: "Failed to block unbalanced journal entry creation", Evidence: fmt.Sprintf("HTTP %d", status)}
		s.registry.MarkFail(sID, f)
		findings = append(findings, f)
	}

	// FIN-NEG-CL-01: Mutation in Closed Period
	sID = "FIN-NEG-CL-01"
	// Attempt journal on a very old date (assumed closed or invalid)
	closedPeriodJE := map[string]interface{}{
		"entry_date": "2010-01-01",
		"description": "Closed Period Test Journal " + runID,
		"lines": []map[string]interface{}{
			{"chart_of_account_id": s.client.GetCOAID("11100"), "debit": 1000, "credit": 1000},
		},
	}
	status, _, _ = s.client.Request("POST", "/finance/journal-entries", closedPeriodJE, RoleAdmin)
	if status == http.StatusForbidden || status == http.StatusBadRequest {
		s.registry.MarkPass(sID, analyzer.Finding{Code: "SIM-NEG-02", Severity: analyzer.SeverityPass, Module: "finance", Message: "Blocked mutation in closed/invalid period"})
	} else {
		f := analyzer.Finding{Code: "SIM-NEG-02", Severity: analyzer.SeverityError, Module: "finance", Message: "Failed to block mutation in old/closed period", Evidence: fmt.Sprintf("HTTP %d", status)}
		s.registry.MarkFail(sID, f)
		findings = append(findings, f)
	}

	return findings
}
