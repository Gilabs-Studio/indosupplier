package usecase

// Well-known Chart of Account codes used by finance business logic.
// These codes must exist in the COA master data (seeder or manual setup).
// Using code-based lookup instead of name-based ILIKE for reliability
// - COA codes are unique and stable, names may be renamed by users.

const (
	// COACodeNonTradePayable is the liability account for non-trade payables (Hutang Non-Dagang).
	// Default COA code: "21600". Adjust if your COA chart uses a different code.
	COACodeNonTradePayable = "21600"

	// COACodeTravelExpense is the expense account for up-country/travel costs (Perjalanan Dinas).
	// Default COA code: "62700". Adjust if your COA chart uses a different code.
	COACodeTravelExpense = "62700"

	// COACodeAccruedExpense is the liability account for accrued expenses / reimbursement payable (Hutang Biaya).
	// Default COA code: "21600". Adjust if your COA chart uses a different code.
	COACodeAccruedExpense = "21600"

	// COACodeRetainedEarnings is the equity account for retained earnings (Laba Ditahan).
	// Used by year-end closing to transfer net income.
	COACodeRetainedEarnings = "33000"
)




