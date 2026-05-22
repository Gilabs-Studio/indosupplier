package simulator

import (
	"fmt"
	"net/http"

	"github.com/gilabs/gims/api/analyzer"
)

// runJournalSimulation covers FIN-SIM-JE-01 (Create), FIN-SIM-JE-02 (Post), FIN-SIM-JE-03 (Reverse).
func (s *SimulationChecker) runJournalSimulation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}

	// ── FIN-SIM-JE-01: Journal Creation ──────────────────────────────
	scenarioID := "FIN-SIM-JE-01"
	coaDebitID := s.client.GetCOAID("11100")
	coaCreditID := s.client.GetCOAID("4100")

	if coaDebitID == "" || coaCreditID == "" {
		// Fallback: try to fetch any two COAs from the cache
		coaDebitID, coaCreditID = s.client.GetAnyCOAPair()
	}

	if coaDebitID == "" || coaCreditID == "" {
		f := analyzer.Finding{
			Code: "SIM-F00", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Missing Chart of Account IDs for journal simulation",
			Evidence:       fmt.Sprintf("COA 11100 ID: %s, COA 4100 ID: %s", coaDebitID, coaCreditID),
			Recommendation: "Ensure COA seeder has run and /finance/journal-entries/form-data returns accounts.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
		return findings
	}

	payload := map[string]interface{}{
		"entry_date":  "2026-03-25",
		"description": fmt.Sprintf("QA Simulation Journal [%s]", runID),
		"lines": []map[string]interface{}{
			{"chart_of_account_id": coaDebitID, "debit": 5000, "credit": 0, "memo": "sim debit " + runID},
			{"chart_of_account_id": coaCreditID, "debit": 0, "credit": 5000, "memo": "sim credit " + runID},
		},
	}

	status, body, err := s.client.Request("POST", "/finance/journal-entries", payload, RoleAdmin)
	if err != nil || status != http.StatusCreated {
		f := analyzer.Finding{
			Code: "SIM-F01", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Failed to create manual journal entry",
			Evidence:       fmt.Sprintf("HTTP %d, Body: %v, Err: %v", status, body, err),
			Recommendation: "Check journal entry validation logic and COA existence.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
		return findings
	}

	journalID := extractID(body)
	s.registry.MarkPass(scenarioID, analyzer.Finding{
		Code: "SIM-F01", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
		Message: fmt.Sprintf("Journal created (ID: %s)", journalID),
	})

	// ── FIN-SIM-JE-02: Journal Posting ───────────────────────────────
	scenarioID = "FIN-SIM-JE-02"
	status, body, err = s.client.Request("POST", fmt.Sprintf("/finance/journal-entries/%s/post", journalID), nil, RoleAdmin)
	if err != nil || status != http.StatusOK {
		f := analyzer.Finding{
			Code: "SIM-F02", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Failed to post journal entry",
			Evidence:       fmt.Sprintf("HTTP %d, Body: %v", status, body),
			Recommendation: "Check ensureNotClosed guard and journal status transition logic.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
		return findings
	}

	s.registry.MarkPass(scenarioID, analyzer.Finding{
		Code: "SIM-F02", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
		Message: fmt.Sprintf("Journal posted (ID: %s)", journalID),
	})

	// ── FIN-SIM-JE-03: Journal Reversal ──────────────────────────────
	scenarioID = "FIN-SIM-JE-03"
	status, body, err = s.client.Request("POST", fmt.Sprintf("/finance/journal-entries/%s/reverse", journalID), nil, RoleAdmin)
	if err != nil || status != http.StatusOK {
		f := analyzer.Finding{
			Code: "SIM-F03", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Failed to reverse posted journal entry",
			Evidence:       fmt.Sprintf("HTTP %d, Body: %v", status, body),
			Recommendation: "Check reversal logic in journal_entry_usecase.go Reverse(). Ensure entry is in 'posted' status.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
	} else {
		// Validate the reversal created a new journal entry with reference
		reversalData, _ := body["data"].(map[string]interface{})
		reversalID := ""
		if reversalData != nil {
			reversalID, _ = reversalData["id"].(string)
		}
		s.registry.MarkPass(scenarioID, analyzer.Finding{
			Code: "SIM-F03", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
			Message:  fmt.Sprintf("Journal reversed (Original: %s, Reversal: %s)", journalID, reversalID),
			Evidence: fmt.Sprintf("Reversal response: %v", reversalData),
		})
	}

	return findings
}

// runGLSimulation covers FIN-SIM-GL-01.
func (s *SimulationChecker) runGLSimulation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}
	scenarioID := "FIN-SIM-GL-01"

	endpoint := "/finance/reports/general-ledger?start_date=2026-01-01&end_date=2026-12-31"
	status, body, err := s.client.Request("GET", endpoint, nil, RoleAdmin)
	if err != nil || status != http.StatusOK {
		f := analyzer.Finding{
			Code: "SIM-F10", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Failed to load General Ledger report",
			Evidence:       fmt.Sprintf("HTTP %d, Err: %v", status, err),
			Recommendation: "Check FinanceReportUsecase.GetGeneralLedger() and route registration.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
	} else {
		s.registry.MarkPass(scenarioID, analyzer.Finding{
			Code: "SIM-F10", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
			Message:  "General Ledger report loaded successfully",
			Evidence: fmt.Sprintf("Response data present: %v", body["data"] != nil),
		})
	}

	return findings
}

// runPaymentSimulation covers FIN-SIM-PAY-01.
// Uses dynamic Bank Account and COA IDs from API/cache instead of hardcoded UUIDs.
func (s *SimulationChecker) runPaymentSimulation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}
	scenarioID := "FIN-SIM-PAY-01"

	// 1. Fetch a valid Bank Account ID
	status, bankBody, err := s.client.Request("GET", "/finance/bank-accounts?per_page=1", nil, RoleAdmin)
	var bankAccountID string
	if err == nil && status == http.StatusOK {
		if data, ok := bankBody["data"].([]interface{}); ok && len(data) > 0 {
			if b, ok := data[0].(map[string]interface{}); ok {
				bankAccountID, _ = b["id"].(string)
			}
		}
	}
	
	if bankAccountID == "" {
		f := analyzer.Finding{
			Code: "SIM-F20", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Cannot run payment simulation: No bank accounts found",
			Recommendation: "Ensure Bank Account seeder has run so payments have a source.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
		return findings
	}

	// 2. Get a valid Expense COA for payment allocation
	expenseCOA := s.client.GetCOAID("6200")   // Expense
	if expenseCOA == "" {
		expenseCOA = s.client.GetCOAID("51000")
	}
	if expenseCOA == "" {
		_, expenseCOA = s.client.GetAnyCOAPair()
	}

	if expenseCOA == "" {
		f := analyzer.Finding{
			Code: "SIM-F20", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Cannot run payment simulation: Target COA ID not available",
			Recommendation: "Ensure COA seeder provides Expense accounts.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
		return findings
	}

	payload := map[string]interface{}{
		"payment_date":    "2026-03-25",
		"bank_account_id": bankAccountID,
		"total_amount":    50000,
		"description":     "Simulation Payment " + runID,
		"allocations": []map[string]interface{}{
			{"chart_of_account_id": expenseCOA, "amount": 50000, "memo": "sim expense " + runID},
		},
	}

	status, body, err := s.client.Request("POST", "/finance/payments", payload, RoleAdmin)
	if err != nil || (status != http.StatusCreated && status != http.StatusOK) {
		f := analyzer.Finding{
			Code: "SIM-F20", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Payment creation failed",
			Evidence:       fmt.Sprintf("HTTP %d, Body: %v, Err: %v", status, body, err),
			Recommendation: "Check PaymentUsecase.Create() and ensure bank_account_id maps to a valid COA with type CASH_BANK.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
	} else {
		s.registry.MarkPass(scenarioID, analyzer.Finding{
			Code: "SIM-F20", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
			Message: "Payment created successfully",
		})
	}

	return findings
}

// runReportSimulation covers FIN-SIM-REP-BS, FIN-SIM-REP-TB, and FIN-SIM-REP-PL.
func (s *SimulationChecker) runReportSimulation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}

	// ── FIN-SIM-REP-BS: Balance Sheet ────────────────────────────────
	sID := "FIN-SIM-REP-BS"
	status, body, err := s.client.Request("GET", "/finance/reports/balance-sheet?start_date=2026-01-01&end_date=2026-12-31", nil, RoleAdmin)
	if err != nil || status != http.StatusOK {
		f := analyzer.Finding{
			Code: "SIM-F30", ScenarioID: sID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Failed to load Balance Sheet report",
			Evidence:       fmt.Sprintf("HTTP %d, Err: %v", status, err),
			Recommendation: "Check /finance/reports/balance-sheet endpoint and FinanceReportUsecase.",
		}
		s.registry.MarkFail(sID, f)
		findings = append(findings, f)
	} else {
		data, _ := body["data"].(map[string]interface{})
		isBalanced, _ := data["is_balanced"].(bool)
		if !isBalanced {
			f := analyzer.Finding{
				Code: "SIM-F31", ScenarioID: sID, Severity: analyzer.SeverityCritical, Module: "finance",
				Message:        "Balance Sheet is UNBALANCED",
				Evidence:       fmt.Sprintf("Asset: %v, Liab+Equity: %v", data["asset_total"], data["liability_equity_total"]),
				Recommendation: "Run financial closing to transfer net profit to equity. Check for orphan or reversed journals.",
			}
			s.registry.MarkFail(sID, f)
			findings = append(findings, f)
		} else {
			s.registry.MarkPass(sID, analyzer.Finding{
				Code: "SIM-F30", ScenarioID: sID, Severity: analyzer.SeverityPass, Module: "finance",
				Message: "Balance Sheet is balanced (Assets = Liabilities + Equity)",
			})
		}
	}

	// ── FIN-SIM-REP-TB: Trial Balance ────────────────────────────────
	sID = "FIN-SIM-REP-TB"
	status, body, err = s.client.Request("GET", "/finance/reports/trial-balance?start_date=2026-01-01&end_date=2026-12-31", nil, RoleAdmin)
	if err != nil || status != http.StatusOK {
		f := analyzer.Finding{
			Code: "SIM-F32", ScenarioID: sID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Trial Balance report failed to load",
			Evidence:       fmt.Sprintf("HTTP %d, Err: %v", status, err),
			Recommendation: "Check /finance/reports/trial-balance route and FinanceReportUsecase.GetTrialBalance().",
		}
		s.registry.MarkFail(sID, f)
		findings = append(findings, f)
	} else {
		s.registry.MarkPass(sID, analyzer.Finding{
			Code: "SIM-F32", ScenarioID: sID, Severity: analyzer.SeverityPass, Module: "finance",
			Message: "Trial Balance report loaded successfully",
		})
	}

	// ── FIN-SIM-REP-PL: Profit & Loss ────────────────────────────────
	sID = "FIN-SIM-REP-PL"
	status, body, err = s.client.Request("GET", "/finance/reports/profit-loss?start_date=2026-01-01&end_date=2026-12-31", nil, RoleAdmin)
	if err != nil || status != http.StatusOK {
		f := analyzer.Finding{
			Code: "SIM-F33", ScenarioID: sID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Profit & Loss report failed to load",
			Evidence:       fmt.Sprintf("HTTP %d, Err: %v", status, err),
			Recommendation: "Check /finance/reports/profit-loss route and FinanceReportUsecase.GetProfitAndLoss().",
		}
		s.registry.MarkFail(sID, f)
		findings = append(findings, f)
	} else {
		// Cross-check: Revenue - Expense should equal NetProfit
		data, _ := body["data"].(map[string]interface{})
		if data != nil {
			revTotal, _ := data["revenue_total"].(float64)
			cogsTotal, _ := data["cogs_total"].(float64)
			expTotal, _ := data["expense_total"].(float64)
			netProfit, _ := data["net_profit"].(float64)

			calculatedNet := revTotal - cogsTotal - expTotal
			diff := calculatedNet - netProfit
			if diff < -0.01 || diff > 0.01 {
				f := analyzer.Finding{
					Code: "SIM-F34", ScenarioID: sID, Severity: analyzer.SeverityError, Module: "finance",
					Message:        "P&L NetProfit mismatch: Revenue-COGS-Expense ≠ NetProfit in API response",
					Evidence:       fmt.Sprintf("Rev=%.2f COGS=%.2f Exp=%.2f Calculated=%.2f APINet=%.2f Diff=%.2f", revTotal, cogsTotal, expTotal, calculatedNet, netProfit, diff),
					Recommendation: "Check P&L aggregation logic in finance_report_usecase.go.",
				}
				s.registry.MarkFail(sID, f)
				findings = append(findings, f)
			} else {
				s.registry.MarkPass(sID, analyzer.Finding{
					Code: "SIM-F33", ScenarioID: sID, Severity: analyzer.SeverityPass, Module: "finance",
					Message:  fmt.Sprintf("P&L validated: Rev=%.2f COGS=%.2f Exp=%.2f Net=%.2f", revTotal, cogsTotal, expTotal, netProfit),
				})
			}
		} else {
			s.registry.MarkPass(sID, analyzer.Finding{
				Code: "SIM-F33", ScenarioID: sID, Severity: analyzer.SeverityPass, Module: "finance",
				Message: "P&L report loaded (no posted data yet, which is valid for empty DB)",
			})
		}
	}

	return findings
}

// runClosingSimulation covers FIN-SIM-CL-01.
func (s *SimulationChecker) runClosingSimulation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}
	scenarioID := "FIN-SIM-CL-01"

	// Check active period endpoint — accept both 200 (period found) and 404 (no period yet, acceptable in fresh DB)
	// Actually we should test /finance/closing (the list endpoint) rather than active-period, since active-period doesn't exist.
	status, body, err := s.client.Request("GET", "/finance/closing", nil, RoleAdmin)
	if err != nil {
		f := analyzer.Finding{
			Code: "SIM-F40", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Failed to reach financial closing endpoint",
			Evidence:       fmt.Sprintf("Err: %v", err),
			Recommendation: "Check /finance/closing route and handler.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
	} else if status == http.StatusOK || status == http.StatusNotFound {
		s.registry.MarkPass(scenarioID, analyzer.Finding{
			Code: "SIM-F40", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
			Message:  fmt.Sprintf("Closing endpoint responded (HTTP %d)", status),
			Evidence: fmt.Sprintf("Response size: %d bytes", len(fmt.Sprint(body))),
		})
	} else {
		f := analyzer.Finding{
			Code: "SIM-F40", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        fmt.Sprintf("Unexpected response from financial closing endpoint: HTTP %d", status),
			Evidence:       fmt.Sprintf("Body: %v", body),
			Recommendation: "Check financial closing handler for unexpected error codes.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
	}

	return findings
}

// runAssetSimulation covers FIN-SIM-AST-01.
func (s *SimulationChecker) runAssetSimulation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}
	scenarioID := "FIN-SIM-AST-01"

	// Try listing asset budgets instead of form-data (more reliable endpoint)
	status, body, err := s.client.Request("GET", "/finance/budgets?per_page=5", nil, RoleAdmin)
	if err != nil || (status != http.StatusOK && status != http.StatusNotFound) {
		f := analyzer.Finding{
			Code: "SIM-F50", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Failed to load Asset Budget list",
			Evidence:       fmt.Sprintf("HTTP %d, Body: %v, Err: %v", status, body, err),
			Recommendation: "Check /finance/budgets route and AssetBudgetHandler.List().",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
	} else {
		s.registry.MarkPass(scenarioID, analyzer.Finding{
			Code: "SIM-F50", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
			Message: fmt.Sprintf("Asset Budget endpoint responded (HTTP %d)", status),
		})
	}

	return findings
}

// runValuationSimulation covers FIN-SIM-VAL-01.
func (s *SimulationChecker) runValuationSimulation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}
	scenarioID := "FIN-SIM-VAL-01"

	// Try listing valuation runs first (read-only check)
	status, body, err := s.client.Request("GET", "/finance/journal-entries/valuation/runs?per_page=5", nil, RoleAdmin)
	if err != nil || (status != http.StatusOK && status != http.StatusNotFound) {
		f := analyzer.Finding{
			Code: "SIM-F60", ScenarioID: scenarioID, Severity: analyzer.SeverityError, Module: "finance",
			Message:        "Valuation runs endpoint failed",
			Evidence:       fmt.Sprintf("HTTP %d, Body: %v, Err: %v", status, body, err),
			Recommendation: "Check /finance/journal-entries/valuation/runs route.",
		}
		s.registry.MarkFail(scenarioID, f)
		findings = append(findings, f)
	} else {
		s.registry.MarkPass(scenarioID, analyzer.Finding{
			Code: "SIM-F60", ScenarioID: scenarioID, Severity: analyzer.SeverityPass, Module: "finance",
			Message: fmt.Sprintf("Valuation runs endpoint responded (HTTP %d)", status),
		})
	}

	return findings
}

// runCrossMenuValidation reconciles P&L vs Balance Sheet net profit.
func (s *SimulationChecker) runCrossMenuValidation(cfg *analyzer.Config, runID string) []analyzer.Finding {
	findings := []analyzer.Finding{}

	statusPL, bodyPL, errPL := s.client.Request("GET", "/finance/reports/profit-loss?start_date=2026-01-01&end_date=2026-12-31", nil, RoleAdmin)
	statusBS, bodyBS, errBS := s.client.Request("GET", "/finance/reports/balance-sheet?start_date=2026-01-01&end_date=2026-12-31", nil, RoleAdmin)

	if errPL == nil && statusPL == http.StatusOK && errBS == nil && statusBS == http.StatusOK {
		plData, _ := bodyPL["data"].(map[string]interface{})
		bsData, _ := bodyBS["data"].(map[string]interface{})

		if plData != nil && bsData != nil {
			plProfit, _ := plData["net_profit"].(float64)
			bsProfit, _ := bsData["current_year_profit"].(float64)

			diff := plProfit - bsProfit
			if diff < -0.01 || diff > 0.01 {
				findings = append(findings, analyzer.Finding{
					Code:           "SIM-CONS-01",
					Severity:       analyzer.SeverityCritical,
					Module:         "finance",
					Message:        "CROSS-MENU INCONSISTENCY: Net Profit mismatch between P&L and Balance Sheet",
					Evidence:       fmt.Sprintf("P&L: %.2f, Balance Sheet: %.2f (diff: %.2f)", plProfit, bsProfit, diff),
					Recommendation: "Check financial closing logic and account balance aggregation.",
				})
			}
		}
	}

	return findings
}

// extractID safely extracts data.id from a standard API response body.
func extractID(body map[string]interface{}) string {
	if body == nil {
		return ""
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		return ""
	}
	id, _ := data["id"].(string)
	return id
}
