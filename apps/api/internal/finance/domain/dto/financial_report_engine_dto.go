package dto

import (
	"time"

	"github.com/google/uuid"
)

type GLParams struct {
	CompanyID    uuid.UUID
	FiscalYearID uuid.UUID
	AccountID    uuid.UUID
	FromDate     time.Time
	ToDate       time.Time
	Page         int
	Limit        int
}

type TBParams struct {
	CompanyID    uuid.UUID
	FiscalYearID uuid.UUID
	AsOfDate     time.Time
	FromDate     time.Time
	ToDate       time.Time
}

type BSParams struct {
	CompanyID       uuid.UUID
	FiscalYearID    uuid.UUID
	AsOfDate        time.Time
	ComparativeDate *time.Time
	IncludeZero     bool
}

type PLParams struct {
	CompanyID    uuid.UUID
	FiscalYearID uuid.UUID
	FromDate     time.Time
	ToDate       time.Time
	Comparative  bool
}

type CFParams struct {
	CompanyID    uuid.UUID
	FiscalYearID uuid.UUID
	FromDate     time.Time
	ToDate       time.Time
}

type AccountBalance struct {
	OpeningBalance float64 `json:"opening_balance"`
	DebitTotal     float64 `json:"debit_total"`
	CreditTotal    float64 `json:"credit_total"`
	ClosingBalance float64 `json:"closing_balance"`
}

type CashFlowSection struct {
	Name   string      `json:"name"`
	Items  []ReportRow `json:"items"`
	Amount float64     `json:"amount"`
}

type CashFlowReport struct {
	FromDate      time.Time       `json:"from_date"`
	ToDate        time.Time       `json:"to_date"`
	Method        string          `json:"method,omitempty"`
	Operating     CashFlowSection `json:"operating"`
	Investing     CashFlowSection `json:"investing"`
	Financing     CashFlowSection `json:"financing"`
	NetChange     float64         `json:"net_change"`
	BeginningCash float64         `json:"beginning_cash"`
	EndingCash    float64         `json:"ending_cash"`
	IsReconciled  bool            `json:"is_reconciled"`
	ReconciliationGap float64     `json:"reconciliation_gap"`
}
