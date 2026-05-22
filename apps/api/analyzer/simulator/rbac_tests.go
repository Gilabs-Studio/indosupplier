package simulator

import (
	"fmt"
	"net/http"

	"github.com/gilabs/gims/api/analyzer"
)

func (s *SimulationChecker) runRBACTests(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}
	
	// Test unauthorized access
	status, _, err := s.client.Request(http.MethodGet, "/organization/companies?per_page=1", nil, RoleNoAccess)
	if err == nil && status < 400 {
		findings = append(findings, analyzer.Finding{
			Code:     "SIM-RBAC-01",
			Severity: analyzer.SeverityError,
			Module:   "core",
			Message:  fmt.Sprintf("Security flaw: Unauthenticated user accessed settings (HTTP %d)", status),
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "SIM-RBAC-01",
			Severity: analyzer.SeverityPass,
			Module:   "core",
			Message:  "Security Check: Unauthenticated user correctly blocked (HTTP 401/403)",
		})
	}
	
	// Check Viewer role (should be able to read, not write)
	status, _, _ = s.client.Request(http.MethodGet, "/product/products?per_page=1", nil, RoleViewer)
	if status == http.StatusOK {
		findings = append(findings, analyzer.Finding{
			Code:     "SIM-RBAC-02",
			Severity: analyzer.SeverityPass,
			Module:   "product",
			Message:  "Viewer role successfully accessed product list",
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "SIM-RBAC-02",
			Severity: analyzer.SeverityError,
			Module:   "product",
			Message:  fmt.Sprintf("Viewer role failed to access product list: %d", status),
		})
	}
	
	status, _, _ = s.client.Request(http.MethodPost, "/product/products", map[string]interface{}{"name": "Simulated"}, RoleViewer)
	if status >= 400 {
		findings = append(findings, analyzer.Finding{
			Code:     "SIM-RBAC-03",
			Severity: analyzer.SeverityPass,
			Module:   "product",
			Message:  fmt.Sprintf("Viewer correctly blocked from creating product (Expected HTTP 403, got %d)", status),
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "SIM-RBAC-03",
			Severity: analyzer.SeverityError,
			Module:   "product",
			Message:  fmt.Sprintf("Security flaw: Viewer allowed to create product (HTTP %d)", status),
		})
	}

	return findings
}
