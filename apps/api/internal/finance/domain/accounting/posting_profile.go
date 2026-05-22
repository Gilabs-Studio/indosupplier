package accounting

import (
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
)

// PostingProfile defines the accounting rules for generating journal entries
// from a specific type of finance transaction.
type PostingProfile struct {
	// ReferenceType is the canonical reference type (from reference package).
	ReferenceType string

	// JournalType is the journal classification to persist for generated journals.
	JournalType models.JournalType

	// Description template. Use %s placeholders for dynamic values.
	DescriptionTemplate string

	// Rules define the debit/credit line generation logic.
	Rules []PostingRule
}

// PostingRule defines a single journal line within a posting profile.
type PostingRule struct {
	// COASettingKey is the finance_settings key used to resolve the COA code at runtime.
	// If empty, the COA ID must be provided by the transaction itself (e.g. user-selected account).
	COASettingKey string

	// COASource indicates where to get the COA ID if COASettingKey is empty.
	// Possible values: "transaction", "bank_account", "line_item"
	COASource string

	// Side is either "debit" or "credit".
	Side string

	// AmountSource indicates which amount field to use.
	// Possible values: "total", "line_amount", "item_total"
	AmountSource string

	// UseTransactionCOA if true, uses the COA ID provided in TransactionData.TransactionCOAID.
	UseTransactionCOA bool

	// MemoTemplate is the memo text template for this line.
	MemoTemplate string
}

// Predefined posting profiles for standard finance transactions.
var (
	// ProfileNonTradePayableApproval generates journal on NTP approval.
	// Debit: Expense Account (transaction.ChartOfAccountID)
	// Credit: NTP Liability Account (from settings)
	ProfileNonTradePayableApproval = PostingProfile{
		ReferenceType:       reference.RefTypeNonTradePayable,
		JournalType:         models.JournalTypeGeneral,
		DescriptionTemplate: "NTP Approval: %s - %s",
		Rules: []PostingRule{
			{
				COASource:    "transaction",
				Side:         "debit",
				AmountSource: "total",
				MemoTemplate: "%s",
			},
			{
				COASettingKey: "coa.non_trade_payable",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Payable record",
			},
		},
	}

	// ProfileNonTradePayablePayment generates journal on NTP payment.
	// Debit: NTP Liability Account (from settings) — settling payable
	// Credit: Bank/Cash Account (user-selected)
	ProfileNonTradePayablePayment = PostingProfile{
		ReferenceType:       reference.RefTypeNTPPayment,
		JournalType:         models.JournalTypeGeneral,
		DescriptionTemplate: "NTP Payment: %s - Ref: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.non_trade_payable",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Settling payable",
			},
			{
				COASource:    "payment_account",
				Side:         "credit",
				AmountSource: "total",
				MemoTemplate: "%s",
			},
		},
	}

	// ProfileUpCountryApproval generates journal on up-country cost manager approval.
	// Debit: Travel Expense Account (from settings)
	// Credit: Accrued Expense Account (from settings)
	ProfileUpCountryApproval = PostingProfile{
		ReferenceType:       reference.RefTypeUpCountryCost,
		JournalType:         models.JournalTypeAdjustment,
		DescriptionTemplate: "Up-Country Cost Approval: %s - %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.travel_expense",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Travel Expense",
			},
			{
				COASettingKey: "coa.accrued_expense",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Reimbursement payable",
			},
		},
	}

	// ProfileCashBankCashIn generates journal for cash-in transactions.
	// Debit: Bank Account COA (from bank_account.chart_of_account_id)
	// Credit: Counter-party accounts (line items)
	ProfileCashBankCashIn = PostingProfile{
		ReferenceType:       reference.RefTypeCashBank,
		JournalType:         models.JournalTypeGeneral,
		DescriptionTemplate: "%s",
		Rules: []PostingRule{
			{
				COASource:    "bank_account",
				Side:         "debit",
				AmountSource: "total",
				MemoTemplate: "Cash/Bank inflow",
			},
			// Line items are generated dynamically per cash bank journal line
		},
	}

	// ProfileCashBankCashOut generates journal for cash-out transactions.
	// Debit: Counter-party accounts (line items)
	// Credit: Bank Account COA (from bank_account.chart_of_account_id)
	ProfileCashBankCashOut = PostingProfile{
		ReferenceType:       reference.RefTypeCashBank,
		JournalType:         models.JournalTypeGeneral,
		DescriptionTemplate: "%s",
		Rules: []PostingRule{
			{
				COASource:    "bank_account",
				Side:         "credit",
				AmountSource: "total",
				MemoTemplate: "Cash/Bank outflow",
			},
		},
	}

	// ProfileCashBankTransfer generates journal for inter-bank transfers.
	// Debit: Destination bank accounts (line items)
	// Credit: Source Bank Account COA
	ProfileCashBankTransfer = PostingProfile{
		ReferenceType:       reference.RefTypeCashBank,
		JournalType:         models.JournalTypeGeneral,
		DescriptionTemplate: "%s",
		Rules: []PostingRule{
			{
				COASource:    "bank_account",
				Side:         "credit",
				AmountSource: "total",
				MemoTemplate: "Inter-bank transfer out",
			},
		},
	}

	// ProfilePayment generates journal on payment approval.
	// Debit: Allocation accounts (line items)
	// Credit: Bank Account COA
	ProfilePayment = PostingProfile{
		ReferenceType:       reference.RefTypePayment,
		JournalType:         models.JournalTypeGeneral,
		DescriptionTemplate: "%s",
		Rules: []PostingRule{
			{
				COASource:    "bank_account",
				Side:         "credit",
				AmountSource: "total",
				MemoTemplate: "Payment bank outflow",
			},
		},
	}

	// ProfilePeriodClosing generates year-end closing journal.
	// Debit/Credit: Revenue/Expense summary → Retained Earnings
	ProfilePeriodClosing = PostingProfile{
		ReferenceType:       reference.RefTypePeriodClosing,
		JournalType:         models.JournalTypeClosing,
		DescriptionTemplate: "Year-End Closing: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.retained_earnings",
				Side:          "dynamic", // determined at runtime based on net income sign
				AmountSource:  "calculated",
				MemoTemplate:  "Year-end closing transfer",
			},
		},
	}

	// ProfileSalesInvoice generates journal for customer invoices.
	// Debit: Trade Receivables (coa.sales_receivable)
	// Credit: Sales Revenue (coa.sales_revenue)
	// Credit: VAT Out (coa.sales_vat_out) if applicable
	ProfileSalesInvoice = PostingProfile{
		ReferenceType:       reference.RefTypeSalesInvoice,
		JournalType:         models.JournalTypeSales,
		DescriptionTemplate: "Invoice %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.sales_receivable",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Trade Receivables from customer invoice",
			},
			{
				COASettingKey: "coa.sales_revenue",
				Side:          "credit",
				AmountSource:  "sub_total", // requires engine support for sub_total source
				MemoTemplate:  "Sales Revenue from invoice items",
			},
			{
				COASettingKey: "coa.sales_vat_out",
				Side:          "credit",
				AmountSource:  "tax_total",
				MemoTemplate:  "VAT Output",
			},
		},
	}

	// ProfileSupplierInvoice generates journal for supplier invoices.
	// Debit: GR/IR (coa.gr_ir)
	// Credit: Trade Payables (coa.purchase_payable)
	ProfileSupplierInvoice = PostingProfile{
		ReferenceType:       reference.RefTypeSupplierInvoice,
		JournalType:         models.JournalTypePurchase,
		DescriptionTemplate: "Supplier Invoice %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.purchase_payable",
				Side:          "credit",
				AmountSource:  "net_total",
				MemoTemplate:  "Trade Payables to supplier (net of DP)",
			},
			{
				COASettingKey: "coa.purchase_advance",
				Side:          "credit",
				AmountSource:  "deposit_total",
				MemoTemplate:  "Applied Supplier Advance Settlement",
			},
			{
				COASettingKey: "coa.gr_ir",
				Side:          "debit",
				AmountSource:  "sub_total",
				MemoTemplate:  "Clearing GR/IR for received goods",
			},
			{
				COASettingKey: "coa.purchase_price_diff",
				Side:          "dynamic",
				AmountSource:  "gr_ir_variance",
				MemoTemplate:  "GR/IR Variance",
			},
			{
				COASettingKey: "coa.purchase_vat_in",
				Side:          "debit",
				AmountSource:  "tax_total",
				MemoTemplate:  "VAT Input",
			},
			{
				COASettingKey: "coa.purchase_expense",
				Side:          "debit",
				AmountSource:  "other_total",
				MemoTemplate:  "Delivery/Other Costs",
			},
		},
	}

	// ProfileSalesPayment generates journal for regular customer payments.
	ProfileSalesPayment = PostingProfile{
		ReferenceType:       reference.RefTypeSalesPayment,
		JournalType:         models.JournalTypeSales,
		DescriptionTemplate: "Customer Payment %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.sales_receivable",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Payment for Regular Invoice",
			},
			{
				UseTransactionCOA: true,
				Side:              "debit",
				AmountSource:      "total",
				MemoTemplate:      "Inbound Payment to Bank/Cash",
			},
		},
	}

	// ProfileSalesPaymentDP generates journal for down payment payments.
	ProfileSalesPaymentDP = PostingProfile{
		ReferenceType:       reference.RefTypeSalesPayment,
		JournalType:         models.JournalTypeSales,
		DescriptionTemplate: "Customer Down Payment %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.sales_advance",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Customer Advance Payment",
			},
			{
				UseTransactionCOA: true,
				Side:              "debit",
				AmountSource:      "total",
				MemoTemplate:      "Inbound Payment to Bank/Cash",
			},
		},
	}

	// ProfileSalesInvoiceDP generates journal for customer down payment invoice approval.
	ProfileSalesInvoiceDP = PostingProfile{
		ReferenceType:       "SALES_INVOICE_DP",
		JournalType:         models.JournalTypeSales,
		DescriptionTemplate: "Customer DP Invoice %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.sales_receivable",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Trade Receivable DP",
			},
			{
				COASettingKey: "coa.sales_advance",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Customer Advance Liability",
			},
		},
	}

	// ProfileSupplierInvoiceDP generates journal for supplier down payment invoice approval.
	ProfileSupplierInvoiceDP = PostingProfile{
		ReferenceType:       "SUPPLIER_INVOICE_DP",
		JournalType:         models.JournalTypePurchase,
		DescriptionTemplate: "Supplier DP Invoice %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.purchase_payable",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Trade Payable DP",
			},
			{
				COASettingKey: "coa.purchase_advance",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Supplier Advance Asset",
			},
		},
	}

	// ProfileSalesReturn generates journal for sales return processing.
	ProfileSalesReturn = PostingProfile{
		ReferenceType:       "SALES_RETURN",
		JournalType:         models.JournalTypeSales,
		DescriptionTemplate: "Sales Return %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.sales_receivable",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "A/R Adjustment for Return",
			},
			{
				COASettingKey: "coa.sales_return",
				Side:          "debit",
				AmountSource:  "sub_total",
				MemoTemplate:  "Sales Return Recognition",
			},
			{
				COASettingKey: "coa.sales_vat_out",
				Side:          "debit",
				AmountSource:  "tax_total",
				MemoTemplate:  "VAT Out Reversal for Return",
			},
		},
	}

	// ProfilePurchaseReturn generates journal for purchase return processing.
	ProfilePurchaseReturn = PostingProfile{
		ReferenceType:       "PURCHASE_RETURN",
		JournalType:         models.JournalTypePurchase,
		DescriptionTemplate: "Purchase Return %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.purchase_payable",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "A/P Adjustment for Return",
			},
			{
				COASettingKey: "coa.purchase_return",
				Side:          "credit",
				AmountSource:  "sub_total",
				MemoTemplate:  "Purchase Return Recognition",
			},
			{
				COASettingKey: "coa.purchase_vat_in",
				Side:          "credit",
				AmountSource:  "tax_total",
				MemoTemplate:  "VAT In Reversal for Return",
			},
		},
	}

	// ProfilePurchasePayment generates journal for regular supplier payments.
	ProfilePurchasePayment = PostingProfile{
		ReferenceType:       "PURCHASE_PAYMENT",
		JournalType:         models.JournalTypePurchase,
		DescriptionTemplate: "Supplier Payment %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.purchase_payable",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Payment for Regular Invoice",
			},
			{
				UseTransactionCOA: true,
				Side:              "credit",
				AmountSource:      "total",
				MemoTemplate:      "Outbound Payment from Bank/Cash (via mapping)",
			},
		},
	}

	// ProfilePurchasePaymentDP generates journal for supplier down payments.
	ProfilePurchasePaymentDP = PostingProfile{
		ReferenceType:       "PURCHASE_PAYMENT_DP",
		JournalType:         models.JournalTypePurchase,
		DescriptionTemplate: "Supplier Down Payment %s: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.purchase_advance",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Supplier Advance Payment",
			},
			{
				UseTransactionCOA: true,
				Side:              "credit",
				AmountSource:      "total",
				MemoTemplate:      "Outbound Payment from Bank/Cash (via mapping)",
			},
		},
	}

	// ProfileGoodsReceipt generates accrual journal when goods are received.
	ProfileGoodsReceipt = PostingProfile{
		ReferenceType:       "GOODS_RECEIPT",
		DescriptionTemplate: "Stock Accrual for GR %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.inventory",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Inventory Increase from GR",
			},
			{
				COASettingKey: "coa.gr_ir",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Accrued Liability (GR/IR)",
			},
		},
	}

	// ProfileInventoryGain generates journal for inventory increase (Stock Opname Gain / Adjustment IN).
	// Debit: Inventory Asset
	// Credit: Inventory Gain (Other Income)
	// ProfileInventoryValuation generates journal for inventory valuation uplift.
	// Debit: Inventory Asset (increase in value)
	// Credit: Revaluation Reserve (EQUITY - not income, per PSAK/IFRS)
	// Note: Revaluation gains/losses are recognized in OCI/Equity, NOT P&L
	ProfileInventoryValuation = PostingProfile{
		ReferenceType:       reference.RefTypeInventoryValuation,
		DescriptionTemplate: "Inventory Revaluation Uplift: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.inventory_asset",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Inventory revaluation uplift (equity)",
			},
			{
				COASettingKey: models.SettingCOAInventoryRevaluationReserve,
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Revaluation surplus - retained earnings",
			},
		},
	}

	// ProfileInventoryValuationLoss generates journal for inventory valuation downswing.
	// Debit: Inventory Loss (expense - recorded in P&L)
	// Credit: Inventory Asset (decrease in value)
	// Note: Losses beyond prior revaluation gains are expensed per PSAK
	ProfileInventoryValuationLoss = PostingProfile{
		ReferenceType:       reference.RefTypeInventoryValuation,
		DescriptionTemplate: "Inventory Revaluation Loss: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.inventory_loss",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Inventory revaluation loss (expense)",
			},
			{
				COASettingKey: "coa.inventory_asset",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Inventory revaluation write-down",
			},
		},
	}

	ProfileInventoryGain = PostingProfile{
		ReferenceType:       "INVENTORY_MOVEMENT",
		DescriptionTemplate: "Inventory Gain: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.inventory_asset",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Inventory increase adjustment",
			},
			{
				COASettingKey: "coa.inventory_gain",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Revenue recognition for inventory gain",
			},
		},
	}

	// ProfileInventoryLoss generates journal for inventory decrease (Stock Opname Loss / Adjustment OUT).
	// Debit: Inventory Loss (Other Expense)
	// Credit: Inventory Asset
	ProfileInventoryLoss = PostingProfile{
		ReferenceType:       "INVENTORY_MOVEMENT",
		DescriptionTemplate: "Inventory Loss: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.inventory_loss",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Expense recognition for inventory loss",
			},
			{
				COASettingKey: "coa.inventory_asset",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Inventory decrease adjustment",
			},
		},
	}

	ProfileFXValuation = PostingProfile{
		ReferenceType:       reference.RefTypeCurrencyRevaluation,
		DescriptionTemplate: "FX Valuation Gain: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.fx_remeasurement",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "FX remeasurement adjustment",
			},
			{
				COASettingKey: "coa.fx_gain",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "FX valuation gain",
			},
		},
	}

	ProfileFXValuationLoss = PostingProfile{
		ReferenceType:       reference.RefTypeCurrencyRevaluation,
		DescriptionTemplate: "FX Valuation Loss: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.fx_loss",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "FX valuation loss",
			},
			{
				COASettingKey: "coa.fx_remeasurement",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "FX remeasurement adjustment",
			},
		},
	}

	// ProfileAssetDepreciation generates journal when asset depreciation is posted.
	// Debit: Depreciation Expense (from settings)
	// Credit: Accumulated Depreciation (from settings)
	ProfileAssetDepreciation = PostingProfile{
		ReferenceType:       reference.RefTypeAssetDepreciation,
		DescriptionTemplate: "Asset Depreciation: %s - %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.depreciation_expense",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Depreciation expense for asset",
			},
			{
				COASettingKey: "coa.depreciation_accumulated",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Accumulated depreciation",
			},
		},
	}

	// ProfileAssetTransaction generates journal for asset purchase, write-off, or disposal.
	// Debit: Fixed Asset Account (from settings)
	// Credit: Cash/Bank Account (user-specified)
	ProfileAssetTransaction = PostingProfile{
		ReferenceType:       reference.RefTypeAssetTransaction,
		DescriptionTemplate: "Asset Transaction: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.fixed_asset",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Asset purchase or addition",
			},
			{
				COASource:    "payment_account",
				Side:         "credit",
				AmountSource: "total",
				MemoTemplate: "Cash payment for asset",
			},
		},
	}

	ProfileDepreciation = PostingProfile{
		ReferenceType:       reference.RefTypeDepreciationValuation,
		DescriptionTemplate: "Depreciation Valuation Loss: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.depreciation_expense",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Depreciation expense adjustment",
			},
			{
				COASettingKey: "coa.depreciation_accumulated",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Accumulated depreciation",
			},
		},
	}

	ProfileDepreciationGain = PostingProfile{
		ReferenceType:       reference.RefTypeDepreciationValuation,
		DescriptionTemplate: "Depreciation Valuation Gain: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.depreciation_accumulated",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Depreciation reversal",
			},
			{
				COASettingKey: "coa.depreciation_gain",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Other income from depreciation reversal",
			},
		},
	}

	// ProfileCOGS generates journal when goods are shipped to customer (COGS recognition).
	// Debit: Cost of Goods Sold
	// Credit: Inventory Asset
	ProfileCOGS = PostingProfile{
		ReferenceType:       "DELIVERY_ORDER",
		DescriptionTemplate: "COGS recognition for DO: %s",
		Rules: []PostingRule{
			{
				COASettingKey: "coa.cogs",
				Side:          "debit",
				AmountSource:  "total",
				MemoTemplate:  "Cost of goods sold",
			},
			{
				COASettingKey: "coa.inventory_asset",
				Side:          "credit",
				AmountSource:  "total",
				MemoTemplate:  "Inventory decrease from shipment",
			},
		},
	}
)
