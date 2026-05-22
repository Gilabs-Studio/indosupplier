package models

import (
	"time"
)

// SystemAccountMapping represents a configuration for well-known accounting accounts.
// This table allows the ERP to stay flexible and avoids hardcoding COA codes in Go logic.
type SystemAccountMapping struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Key       string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_system_account_mappings_key_company,priority:1" json:"key"` // e.g., "sales.accounts_receivable", "purchase.accounts_payable"
	CompanyID *string   `gorm:"type:uuid;uniqueIndex:idx_system_account_mappings_key_company,priority:2;index" json:"company_id"`     // If null, it's a global default
	COACode   string    `gorm:"type:varchar(20);not null" json:"coa_code"`                                                            // The code in chart_of_accounts
	Label     string    `gorm:"type:varchar(200)" json:"label"`                                                                       // Human readable label for the UI
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SystemAccountMapping) TableName() string {
	return "system_account_mappings"
}

// Well-known mapping keys
const (
	MappingKeyPurchaseInventoryAsset      = "purchase.inventory_asset"
	MappingKeyPurchaseGRIRClearing        = "purchase.gr_ir_clearing"
	MappingKeyPurchaseTaxInput            = "purchase.tax_input"
	MappingKeyPurchaseAccountsPayable     = "purchase.accounts_payable"
	MappingKeyPurchaseAdvance             = "purchase.advance"
	MappingKeyPurchaseReturn              = "purchase.purchase_return"
	MappingKeyPurchaseExpense             = "purchase.expense"
	MappingKeySalesAccountsReceivable     = "sales.accounts_receivable"
	MappingKeySalesAdvance                = "sales.advance"
	MappingKeySalesRevenueNew             = "sales.revenue"
	MappingKeySalesTaxOutput              = "sales.tax_output"
	MappingKeySalesCOGSNew                = "sales.cogs"
	MappingKeySalesInventoryAsset         = "sales.inventory_asset"
	MappingKeySalesReturnNew              = "sales.sales_return"
	MappingKeyInventoryAsset              = "inventory.inventory_asset"
	MappingKeyInventoryAdjustmentGain     = "inventory.adjustment_gain"
	MappingKeyInventoryAdjustmentLoss     = "inventory.adjustment_loss"
	MappingKeyInventoryRevaluationReserve = "inventory.revaluation_reserve"
	MappingKeyAssetAccumDepreciation      = "asset.accumulated_depreciation"
	MappingKeyAssetDepreciationExpense    = "asset.depreciation_expense"
	MappingKeyOpeningBalanceEquity        = "finance.opening_balance_equity"
	MappingKeyFinanceBankDefault          = "finance.bank_default"
	MappingKeyFinanceCashDefault          = "finance.cash_default"

	// Direct coa.* aliases for backward compatibility and posting profile keys.
	MappingKeyCOASalesReceivable             = "coa.sales_receivable"
	MappingKeyCOASalesAdvance                = "coa.sales_advance"
	MappingKeyCOASalesRevenue                = "coa.sales_revenue"
	MappingKeyCOASalesVATOut                 = "coa.sales_vat_out"
	MappingKeyCOASalesCOGS                   = "coa.sales_cogs"
	MappingKeyCOASalesInventory              = "coa.sales_inventory"
	MappingKeyCOASalesReturn                 = "coa.sales_return"
	MappingKeyCOAPurchasePayable             = "coa.purchase_payable"
	MappingKeyCOAAccountsPayable             = "coa.accounts_payable"
	MappingKeyCOAPurchaseAdvance             = "coa.purchase_advance"
	MappingKeyCOAPurchaseAdvances            = "coa.purchase_advances"
	MappingKeyCOAGRIR                        = "coa.gr_ir"
	MappingKeyCOAPurchaseGRIR                = "coa.purchase_gr_ir"
	MappingKeyCOAPurchaseVATIn               = "coa.purchase_vat_in"
	MappingKeyCOAPurchasePriceDiff           = "coa.purchase_price_diff"
	MappingKeyCOAVATIn                       = "coa.vat_in"
	MappingKeyCOAPurchaseExpense             = "coa.purchase_expense"
	MappingKeyCOAPurchaseReturn              = "coa.purchase_return"
	MappingKeyCOAInventory                   = "coa.inventory"
	MappingKeyCOAInventoryAsset              = "coa.inventory_asset"
	MappingKeyCOAInventoryGain               = "coa.inventory_gain"
	MappingKeyCOAInventoryLoss               = "coa.inventory_loss"
	MappingKeyCOAInventoryRevaluationReserve = "coa.inventory_revaluation_reserve"
	MappingKeyCOASalesReceivableDP           = "coa.sales_receivable_dp"
	MappingKeyCOAPurchasePayableDP           = "coa.purchase_payable_dp"
	MappingKeyCOACOGS                        = "coa.cogs"

	// Sales related
	MappingKeySalesReceivable    = "SALES_RECEIVABLE"
	MappingKeySalesRevenue       = "SALES_REVENUE"
	MappingKeySalesVatOutput     = "SALES_VAT_OUTPUT"
	MappingKeySalesAdvanceLegacy = "SALES_ADVANCE"
	MappingKeySalesCogs          = "SALES_COGS"
	MappingKeySalesInventory     = "SALES_INVENTORY"

	// Purchase related
	MappingKeyPurchasePayable   = "PURCHASE_PAYABLE"
	MappingKeyPurchaseGrir      = "PURCHASE_GR_IR"
	MappingKeyPurchaseVatInput  = "PURCHASE_VAT_INPUT"
	MappingKeyPurchaseDelivery  = "PURCHASE_DELIVERY_EXPENSE"
	MappingKeyPurchaseOtherCost = "PURCHASE_OTHER_EXPENSE"

	// Finance related
	MappingKeyRetainedEarnings = "RETAINED_EARNINGS"
	MappingKeyClosingSuspense  = "CLOSING_SUSPENSE"
)
