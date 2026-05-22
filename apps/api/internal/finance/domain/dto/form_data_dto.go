package dto

// EnumOption represents a generic enum option for dropdowns.
type EnumOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// AssetCategoryFormOption represents an asset category option for dropdowns.
type AssetCategoryFormOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AssetLocationFormOption represents an asset location option for dropdowns.
type AssetLocationFormOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// --- Module-specific FormData responses ---

// PaymentFormDataResponse for payment form options.
type PaymentFormDataResponse struct {
	ChartOfAccounts []COAFormOption         `json:"chart_of_accounts"`
	BankAccounts    []BankAccountFormOption `json:"bank_accounts"`
}

// CashBankFormDataResponse for cash bank journal form options.
type CashBankFormDataResponse struct {
	ChartOfAccounts []COAFormOption         `json:"chart_of_accounts"`
	BankAccounts    []BankAccountFormOption `json:"bank_accounts"`
	Types           []EnumOption            `json:"types"`
}

// BudgetFormDataResponse for budget form options.
type BudgetFormDataResponse struct {
	ChartOfAccounts []COAFormOption `json:"chart_of_accounts"`
}

// NonTradePayableFormDataResponse for non-trade payable form options.
type NonTradePayableFormDataResponse struct {
	ChartOfAccounts []COAFormOption `json:"chart_of_accounts"`
}

// AssetFormDataResponse for asset form options.
type AssetFormDataResponse struct {
	Categories []AssetCategoryFormOption `json:"categories"`
	Locations  []AssetLocationFormOption `json:"locations"`
}

// AssetCategoryFormDataResponse for asset category form options.
type AssetCategoryFormDataResponse struct {
	ChartOfAccounts     []COAFormOption `json:"chart_of_accounts"`
	DepreciationMethods []EnumOption    `json:"depreciation_methods"`
}

// UpCountryCostFormDataResponse for up-country cost form options.
type UpCountryCostFormDataResponse struct {
	CostTypes []EnumOption `json:"cost_types"`
}
