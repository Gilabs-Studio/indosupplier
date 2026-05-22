package dto

import (
	"time"
)

// GLTransactionRow represents a single journal line in the General Ledger detail.
// All saldo fields are derived exclusively from posted journal_entries and journal_lines.
type GLTransactionRow struct {
	ID            string    `json:"id"`
	JournalID     string    `json:"journal_id"`
	EntryDate     time.Time `json:"entry_date"`
	Description   string    `json:"description"`
	Memo          string    `json:"memo"`
	ReferenceType *string   `json:"reference_type"`
	ReferenceID   *string   `json:"reference_id"`
	// ReferenceCode is a human-readable composite of reference_type + "/" + reference_id
	// for quick display. Empty when no reference is set.
	ReferenceCode string  `json:"reference_code"`
	Debit         float64 `json:"debit"`
	Credit        float64 `json:"credit"`
	// RunningBalance is the cumulative account balance after this line, starting from OpeningBalance.
	// Direction follows the normal balance rule for the account type.
	RunningBalance float64 `json:"running_balance"`
}

// GeneralLedgerAccount is the per-account summary row in the General Ledger report.
// Only accounts with posted activity or non-zero opening balance are included.
type GeneralLedgerAccount struct {
	AccountID      string             `json:"account_id"`
	AccountCode    string             `json:"account_code"`
	AccountName    string             `json:"account_name"`
	AccountType    string             `json:"account_type"`
	OpeningBalance float64            `json:"opening_balance"`
	TotalDebit     float64            `json:"total_debit"`
	TotalCredit    float64            `json:"total_credit"`
	ClosingBalance float64            `json:"closing_balance"`
	Transactions   []GLTransactionRow `json:"transactions"`
}

type GeneralLedgerResponse struct {
	StartDate time.Time              `json:"start_date"`
	EndDate   time.Time              `json:"end_date"`
	Accounts  []GeneralLedgerAccount `json:"accounts"`
}

type ReportRow struct {
	AccountID      string      `json:"account_id,omitempty"`
	Code           string      `json:"code"`
	Name           string      `json:"name"`
	AccountType    string      `json:"account_type,omitempty"`
	ParentID       *string     `json:"parent_id,omitempty"`
	Amount         float64     `json:"amount"`
	SubtotalAmount float64     `json:"subtotal_amount,omitempty"`
	Level          int         `json:"level,omitempty"`
	Children       []ReportRow `json:"children,omitempty"`
	Drilldown      *Drilldown  `json:"drilldown,omitempty"`
	IsTotal        bool        `json:"is_total"`
}

type Drilldown struct {
	GeneralLedgerURL string `json:"general_ledger_url"`
}

type BalanceSheetResponse struct {
	StartDate         time.Time   `json:"start_date"`
	EndDate           time.Time   `json:"end_date"`
	IncludeZero       bool        `json:"include_zero"`
	Assets            []ReportRow `json:"assets"`
	AssetTotal        float64     `json:"asset_total"`
	Liabilities       []ReportRow `json:"liabilities"`
	LiabilityTotal    float64     `json:"liability_total"`
	Equities          []ReportRow `json:"equities"`
	EquityTotal       float64     `json:"equity_total"`
	RetainedEarnings  float64     `json:"retained_earnings"`
	CurrentYearProfit float64     `json:"current_year_profit"`
	EquityTotalFinal  float64     `json:"equity_total_final"`
	LiabilityEquity   float64     `json:"liability_equity_total"`
	ImbalanceAmount   float64     `json:"imbalance_amount"`
	IsBalanced        bool        `json:"is_balanced"`
	BalanceTolerance  float64     `json:"balance_tolerance"`
}

type ProfitAndLossComparison struct {
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	RevenueTotal float64   `json:"revenue_total"`
	COGSTotal    float64   `json:"cogs_total"`
	ExpenseTotal float64   `json:"expense_total"`
	GrossProfit  float64   `json:"gross_profit"`
	NetProfit    float64   `json:"net_profit"`
}

// ProfitAndLossResponse is based exclusively on posted journal entries and supports
// hierarchical account grouping, period comparisons, and margin analysis.
type ProfitAndLossResponse struct {
	StartDate        time.Time               `json:"start_date"`
	EndDate          time.Time               `json:"end_date"`
	Revenues         []ReportRow             `json:"revenues"`
	RevenueTotal     float64                 `json:"revenue_total"`
	COGS             []ReportRow             `json:"cogs"`
	COGSTotal        float64                 `json:"cogs_total"`
	Expenses         []ReportRow             `json:"expenses"`
	ExpenseTotal     float64                 `json:"expense_total"`
	GrossProfit      float64                 `json:"gross_profit"`
	NetProfit        float64                 `json:"net_profit"`
	RetainedEarnings float64                 `json:"retained_earnings"`
	GrossMargin      float64                 `json:"gross_margin"`
	NetMargin        float64                 `json:"net_margin"`
	ExpenseRatio     float64                 `json:"expense_ratio"`
	PreviousPeriod   *ProfitAndLossComparison `json:"previous_period,omitempty"`
	YearToDate       *ProfitAndLossComparison `json:"year_to_date,omitempty"`
}
