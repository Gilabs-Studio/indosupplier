package reference

import "strings"

// Standard Reference Types — uppercase SCREAMING_SNAKE_CASE.
// All finance modules MUST use these constants instead of inline strings.
const (
	// ---- Sales ----
	RefTypeSalesInvoice   = "SALES_INVOICE"
	RefTypeSalesInvoiceDP = "SALES_INVOICE_DP"
	RefTypeSalesPayment   = "SALES_PAYMENT"
	RefTypeSalesReturn    = "SALES_RETURN"

	// ---- Purchase ----
	RefTypeGoodsReceipt      = "GOODS_RECEIPT"
	RefTypeSupplierInvoice   = "SUPPLIER_INVOICE"
	RefTypeSupplierInvoiceDP = "SUPPLIER_INVOICE_DP"
	RefTypePurchasePayment   = "PURCHASE_PAYMENT"
	RefTypePurchaseReturn    = "PURCHASE_RETURN"

	// ---- Inventory ----
	RefTypeStockOpname          = "STOCK_OPNAME"
	RefTypeInventoryAdjustment  = "INVENTORY_ADJUSTMENT"
	RefTypeInventoryValuation   = "INVENTORY_VALUATION"
	RefTypeCurrencyRevaluation  = "CURRENCY_REVALUATION"
	RefTypeCostAdjustment       = "COST_ADJUSTMENT"
	RefTypeDepreciationValuation = "DEPRECIATION_VALUATION"

	// ---- Finance ----
	RefTypeCashBank          = "CASH_BANK"
	RefTypePayment           = "PAYMENT"
	RefTypeNonTradePayable   = "NON_TRADE_PAYABLE"
	RefTypeNTPPayment        = "NTP_PAYMENT"
	RefTypeUpCountryCost     = "UP_COUNTRY_COST"
	RefTypeAssetTransaction  = "ASSET_TRANSACTION"
	RefTypeAssetDepreciation = "ASSET_DEPRECIATION"
	RefTypeSalaryExpense     = "SALARY_EXPENSE"

	// ---- Journal ----
	RefTypeManualAdjustment = "MANUAL_ADJUSTMENT"
	RefTypeAdjustment       = "ADJUSTMENT"
	RefTypeCorrection       = "CORRECTION"
	RefTypeReversal         = "REVERSAL"
	RefTypePeriodClosing    = "PERIOD_CLOSING"
	RefTypeGeneral          = "GENERAL"
)

// legacyAliases maps old / inconsistent reference type strings to canonical values.
// This ensures backward compatibility during the migration period.
var legacyAliases = map[string]string{
	// legacy lowercase keys from existing usecases
	"NTP":           RefTypeNonTradePayable,
	"NTP_PAYMENT":   RefTypeNTPPayment,
	"CASH_BANK":     RefTypeCashBank,
	"CASHBANK":      RefTypeCashBank,
	"CASHIN":        RefTypeCashBank,
	"CASHOUT":       RefTypeCashBank,
	"TRANSFER":      RefTypeCashBank,
	"TRF":           RefTypeCashBank,
	"PAYMENT":       RefTypePayment,
	"UP_COUNTRY":    RefTypeUpCountryCost,
	"UPCOUNTRY":     RefTypeUpCountryCost,
	"REVERSAL":      RefTypeReversal,
	"ASSET_TXN":     RefTypeAssetTransaction,
	"ASSET_DEP":     RefTypeAssetDepreciation,

	// already-standard keys (identity mapping for completeness)
	"SALES_INVOICE":         RefTypeSalesInvoice,
	"SALES_INVOICE_DP":      RefTypeSalesInvoiceDP,
	"SALES_PAYMENT":         RefTypeSalesPayment,
	"SALES_RETURN":          RefTypeSalesReturn,
	"SUPPLIER_INVOICE":      RefTypeSupplierInvoice,
	"SUPPLIER_INVOICE_DP":   RefTypeSupplierInvoiceDP,
	"PURCHASE_PAYMENT":      RefTypePurchasePayment,
	"PURCHASE_RETURN":       RefTypePurchaseReturn,
	"GOODS_RECEIPT":         RefTypeGoodsReceipt,
	"DELIVERY_ORDER":        "DELIVERY_ORDER",
	"SALES_ORDER":           "SALES_ORDER",
	"PURCHASE_ORDER":        "PURCHASE_ORDER",
	"STOCK_OPNAME":          RefTypeStockOpname,
	"INVENTORY_ADJUSTMENT":  RefTypeInventoryAdjustment,
	"INVENTORY_VALUATION":   RefTypeInventoryValuation,
	"CURRENCY_REVALUATION":  RefTypeCurrencyRevaluation,
	"COST_ADJUSTMENT":       RefTypeCostAdjustment,
	"DEPRECIATION_VALUATION": RefTypeDepreciationValuation,
	"MANUAL_ADJUSTMENT":     RefTypeManualAdjustment,
	"NON_TRADE_PAYABLE":     RefTypeNonTradePayable,
	"YEAR_END_CLOSING":      RefTypePeriodClosing,
	"PERIOD_CLOSING":        RefTypePeriodClosing,

	// compact forms (no separators)
	"GOODSRECEIPT":      RefTypeGoodsReceipt,
	"DELIVERYORDER":     RefTypeDeliveryOrder,
	"SALESORDER":        "SALES_ORDER",
	"PURCHASEORDER":     "PURCHASE_ORDER",
	"SALESINVOICE":      RefTypeSalesInvoice,
	"SALESINVOICEDP":    RefTypeSalesInvoiceDP,
	"SUPPLIERINVOICE":   RefTypeSupplierInvoice,
	"SUPPLIERINVOICEDP": RefTypeSupplierInvoiceDP,
	"SALESPAYMENT":      RefTypeSalesPayment,
	"PURCHASEPAYMENT":   RefTypePurchasePayment,
	"STOCKOPNAME":       RefTypeStockOpname,
	"OPNAME":            RefTypeStockOpname,
	"INVENTORYADJUSTMENT": RefTypeInventoryAdjustment,
	"DO":                  RefTypeDeliveryOrder,
	"GR":                  RefTypeGoodsReceipt,
}

// Additional constant for backward compatibility in alias maps.
const RefTypeDeliveryOrder = "DELIVERY_ORDER"

// Normalize converts a raw reference type string to its canonical form.
// It handles legacy lowercase values, compact forms, and mixed separators.
func Normalize(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Uppercase and replace spaces with underscores
	s := strings.ToUpper(raw)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")

	// Direct lookup
	if v, ok := legacyAliases[s]; ok {
		return v
	}

	// Compact form (strip separators, then lookup)
	compact := compactForm(s)
	if v, ok := legacyAliases[compact]; ok {
		return v
	}

	return s
}

// NormalizePtr normalizes a *string reference type, returning the canonical form.
func NormalizePtr(refType *string) string {
	if refType == nil {
		return ""
	}
	return Normalize(*refType)
}

// StringPtr returns a pointer to s.
func StringPtr(s string) *string {
	return &s
}

func compactForm(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
