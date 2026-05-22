package dto

import "time"

// ReconciliationSummary is a high-level summary of a reconciliation run.
type ReconciliationSummary struct {
	TotalSubledger float64 `json:"total_subledger"`
	TotalGL        float64 `json:"total_gl"`
	Difference     float64 `json:"difference"`
	Status         string  `json:"status"` // "MATCHED", "MISMATCHED"
}

// ARAPReconciliationRow is a detail row for AR/AP reconciliation (Invoice vs GL Line).
type ARAPReconciliationRow struct {
	InvoiceID       string    `json:"invoice_id"`
	InvoiceCode     string    `json:"invoice_code"`
	PartnerName     string    `json:"partner_name"` // Customer or Supplier Name
	InvoiceAmount   float64   `json:"invoice_amount"`
	RemainingAmount float64   `json:"remaining_amount"`
	GLBalance       float64   `json:"gl_balance"`
	Difference      float64   `json:"difference"`
	Status          string    `json:"status"` // "MATCHED", "MISMATCHED"
}

// ARAPReconciliationReport is the response for AR/AP reconciliation.
type ARAPReconciliationReport struct {
	Type     string                  `json:"type"` // "AR", "AP"
	AsOfDate time.Time               `json:"as_of_date"`
	Account  ChartOfAccountResponse  `json:"account"` // The GL account used for comparison
	Summary  ReconciliationSummary   `json:"summary"`
	Details  []ARAPReconciliationRow `json:"details"`
}
