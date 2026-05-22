package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FinanceSetting stores runtime-configurable finance parameters.
// This replaces hardcoded constants (e.g. COA codes) so they can be changed
// without recompiling the application.
type FinanceSetting struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SettingKey  string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"setting_key"`
	Value       string         `gorm:"type:text;not null" json:"value"`
	Description string         `gorm:"type:text" json:"description"`
	Category    string         `gorm:"type:varchar(50);index;default:'general'" json:"category"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FinanceSetting) TableName() string {
	return "finance_settings"
}

func (s *FinanceSetting) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// Well-known setting keys for COA code mappings.
// These keys are used by the SettingsService to resolve COA codes at runtime.
const (
	// SettingCOANonTradePayable maps to the liability account for NTP (Hutang Non-Dagang).
	// Default value: "21200"
	SettingCOANonTradePayable = "coa.non_trade_payable"

	// SettingCOATravelExpense maps to the expense account for up-country/travel costs.
	// Default value: "62000"
	SettingCOATravelExpense = "coa.travel_expense"

	// SettingCOAAccruedExpense maps to the liability account for accrued expenses (Hutang Biaya).
	// Default value: "21300"
	SettingCOAAccruedExpense = "coa.accrued_expense"

	// SettingCOARetainedEarnings maps to the equity account for retained earnings (Laba Ditahan).
	// Default value: "32000"
	SettingCOARetainedEarnings = "coa.retained_earnings"

	// SettingCOAAccountsPayable maps to the main AP liability account (Hutang Usaha).
	SettingCOAAccountsPayable = "coa.accounts_payable"

	// SettingCOAPurchaseAdvances maps to the advance payment asset account (Uang Muka Pembelian).
	SettingCOAPurchaseAdvances = "coa.purchase_advances"

	// SettingCOAGoodReceiptInvoiceReceipt maps to the clearing account (GR/IR) for received not yet invoiced goods.
	SettingCOAGoodReceiptInvoiceReceipt = "coa.gr_ir"

	// SettingCOAVATIn maps to the VAT Input asset account (PPN Masukan).
	SettingCOAVATIn = "coa.vat_in"

	// SettingCOAOtherExpense maps to the general other expense account for extra costs.
	SettingCOAOtherExpense = "coa.other_expense"

	// SettingCOASalesReceivable maps to the trade receivables account (Piutang Usaha).
	SettingCOASalesReceivable = "coa.sales_receivable"

	// SettingCOASalesRevenue maps to the main sales revenue account (Pendapatan Penjualan).
	SettingCOASalesRevenue = "coa.sales_revenue"

	// SettingCOASalesVATOut maps to the VAT Output liability account (PPN Keluaran).
	SettingCOASalesVATOut = "coa.sales_vat_out"

	// SettingCOASalesAdvance maps to the customer advance liability account (Uang Muka Penjualan).
	SettingCOASalesAdvance = "coa.sales_advance"

	// SettingCOAPurchasePayable maps to the trade payables liability account (Hutang Usaha).
	SettingCOAPurchasePayable = "coa.purchase_payable"

	// SettingCOAPurchaseGRIR maps to the goods received not invoiced clearing account (Hutang Belum Difakturkan).
	SettingCOAPurchaseGRIR = "coa.purchase_gr_ir"

	// SettingCOAPurchaseVATIn maps to the VAT Input asset account (PPN Masukan).
	SettingCOAPurchaseVATIn = "coa.purchase_vat_in"

	// SettingCOASalesCOGS maps to the cost of goods sold expense account.
	SettingCOASalesCOGS = "coa.sales_cogs"

	// SettingCOASalesInventory maps to the merchandise inventory asset account.
	SettingCOASalesInventory = "coa.sales_inventory"

	// SettingCOAPurchaseExpense maps to the general purchase expense account (delivery/misc).
	SettingCOAPurchaseExpense = "coa.purchase_expense"

	// SettingCOAPurchaseAdvance maps to the supplier advance asset account (Prepaid).
	SettingCOAPurchaseAdvance = "coa.purchase_advance"

	// SettingCOASalesReturn maps to the sales return account.
	SettingCOASalesReturn = "coa.sales_return"

	// SettingCOAPurchaseReturn maps to the purchase return account.
	SettingCOAPurchaseReturn = "coa.purchase_return"

	// SettingCOAInventory maps to the general inventory asset account.
	SettingCOAInventory = "coa.inventory"

	// SettingCOAInventoryAsset maps to the inventory asset account (for gain/loss adjustments).
	SettingCOAInventoryAsset = "coa.inventory_asset"

	// SettingCOAInventoryGain maps to the other income account for inventory gains (opname).
	SettingCOAInventoryGain = "coa.inventory_gain"

	// SettingCOAInventoryLoss maps to the expense account for inventory losses (opname/waste).
	SettingCOAInventoryLoss = "coa.inventory_loss"

	// SettingCOAInventoryRevaluationReserve maps to the equity account for inventory revaluation gains (PSAK/IFRS).
	SettingCOAInventoryRevaluationReserve = "coa.inventory_revaluation_reserve"

	// SettingCOAFXGain maps to the other income account for FX valuation gains.
	SettingCOAFXGain = "coa.fx_gain"

	// SettingCOAFXLoss maps to the expense account for FX valuation losses.
	SettingCOAFXLoss = "coa.fx_loss"

	// SettingCOAFXRemeasurement maps to the balancing account for FX remeasurement.
	SettingCOAFXRemeasurement = "coa.fx_remeasurement"

	// SettingCOADepreciationExpense maps to depreciation expense account.
	SettingCOADepreciationExpense = "coa.depreciation_expense"

	// SettingCOADepreciationAccumulated maps to accumulated depreciation contra-asset account.
	SettingCOADepreciationAccumulated = "coa.depreciation_accumulated"

	// SettingCOADepreciationGain maps to other income account for depreciation reversal gains.
	SettingCOADepreciationGain = "coa.depreciation_gain"

	// SettingCOAInventoryAdjustment maps to the fallback inventory adjustment account.
	SettingCOAInventoryAdjustment = "coa.inventory_adjustment"

	// SettingCOACOGS maps to the cost of goods sold account.
	SettingCOACOGS = "coa.cogs"

	// SettingCOAFixedAsset maps to the fixed asset account for asset purchases.
	SettingCOAFixedAsset = "coa.fixed_asset"

	// SettingCOACash maps to cash on hand account.
	SettingCOACash = "coa.cash"

	// SettingCOABank maps to bank accounts.
	SettingCOABank = "coa.bank"

	// SettingValuationReconciliationTolerance maps to the tolerance for inventory valuation reconciliation.
	// Default value: "0.01" (1% tolerance for rounding differences)
	SettingValuationReconciliationTolerance = "valuation.reconciliation_tolerance"

	// SettingAgingBucketConfig maps to aging report bucket configuration in JSON array format.
	// Example value: [{"key":"current","label":"Current","min_days":null,"max_days":0}, ...]
	SettingAgingBucketConfig = "report.aging.bucket_config"

	// Aging report scopes used for report-specific aging configuration.
	AgingReportTypeAR = "ar"
	AgingReportTypeAP = "ap"
)

func AgingBucketConfigSettingKey(reportType string) string {
	switch strings.ToLower(strings.TrimSpace(reportType)) {
	case AgingReportTypeAP:
		return SettingAgingBucketConfig + ".ap"
	default:
		return SettingAgingBucketConfig + ".ar"
	}
}
