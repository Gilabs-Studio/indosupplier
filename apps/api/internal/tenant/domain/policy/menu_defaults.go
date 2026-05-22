package policy

import "strings"

type PlanEntitlementRow struct {
	PermissionCode string
	MenuURL        string
}

var posGrowthDefaultMenuURLs = []string{
	"/dashboard",
	"/pos",
	"/pos/fb/floor-layout",
	"/pos/fb/terminal",
	"/pos/feedback",
	"/pos/feedback/response",
	"/pos/feedback/forms",
	"/pos/loyalty",
	"/pos/loyalty/config",
	"/pos/loyalty/members",
	"/sales/orders",
	"/sales/invoices",
	"/sales/payments",
	"/purchase/purchase-orders",
	"/purchase/supplier-invoices",
	"/purchase/goods-receipt",
	"/purchase/payments",
	"/stock/inventory",
	"/stock/movements",
	"/finance/journals/purchase",
	"/finance/journals/sales",
	"/finance/bank-accounts",
	"/finance/settings/fiscal-years",
	"/master-data",
	"/master-data/geographic",
	"/master-data/organization",
	"/master-data/users",
	"/master-data/employees",
	"/master-data/company",
	"/master-data/outlet",
	"/master-data/warehouses",
	"/master-data/divisions",
	"/master-data/job-positions",
	"/master-data/business-units",
	"/master-data/business-types",
	"/master-data/areas",
	"/master-data/supplier",
	"/master-data/suppliers",
	"/master-data/supplier-types",
	"/master-data/customer",
	"/master-data/customers",
	"/master-data/customer-types",
	"/master-data/product",
	"/master-data/products",
	"/master-data/product-categories",
	"/master-data/product-types",
	"/master-data/packaging",
	"/master-data/uom",
	"/master-data/procurement-types",
	"/master-data/payment-courier",
	"/master-data/currencies",
	"/master-data/banks",
	"/master-data/payment-terms",
	"/master-data/courier-agencies",
	"/master-data/so-sources",
	"/master-data/leave-types",
}

func NormalizePlanSlug(planSlug string) string {
	normalized := strings.ToLower(strings.TrimSpace(planSlug))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	switch normalized {
	case "pos", "pos_modular":
		return "pos_growth"
	case "erp", "erp_modular":
		return "erp_pro"
	case "crm", "crm_modular":
		return "crm_growth"
	case "hr", "hr_modular":
		return "hr_growth"
	default:
		return normalized
	}
}

func DefaultPlanEntitlementRows(planSlug string) []PlanEntitlementRow {
	switch NormalizePlanSlug(planSlug) {
	case "pos_growth":
		rows := make([]PlanEntitlementRow, 0, len(posGrowthDefaultMenuURLs)+6)
		for _, menuURL := range posGrowthDefaultMenuURLs {
			rows = append(rows, PlanEntitlementRow{MenuURL: menuURL})
		}
		// Finance menu entitlements with permissions
		rows = append(rows,
			PlanEntitlementRow{MenuURL: "/finance/accounting/coa", PermissionCode: "coa.read"},
			PlanEntitlementRow{MenuURL: "/finance/journals/sales", PermissionCode: "sales_journal.read"},
			PlanEntitlementRow{MenuURL: "/finance/journals/purchase", PermissionCode: "purchase_journal.read"},
			PlanEntitlementRow{MenuURL: "/finance/settings/fiscal-years", PermissionCode: "fiscal_year.read"},
			PlanEntitlementRow{MenuURL: "/finance/settings/fiscal-years", PermissionCode: "fiscal_year.write"},
			PlanEntitlementRow{MenuURL: "/finance/settings/fiscal-years", PermissionCode: "fiscal_year.delete"},
			PlanEntitlementRow{MenuURL: "/finance/settings/accounting-mapping", PermissionCode: "account_mappings.read"},
		)
		return rows
	default:
		return nil
	}
}
