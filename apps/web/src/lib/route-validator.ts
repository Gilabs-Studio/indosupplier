import { navigationConfig, type NavItem } from "@/lib/navigation-config";

function flattenNavigationRoutes(items: readonly NavItem[]): string[] {
  const routes: string[] = [];

  for (const item of items) {
    if (item.url) {
      routes.push(item.url);
    }

    if (item.children && item.children.length > 0) {
      routes.push(...flattenNavigationRoutes(item.children));
    }
  }

  return routes;
}

const EXTRA_VALID_ROUTES = [
  "/stock/movements/create",
  "/stock/ledger",
  "/stock/opname",
  "/master-data/currencies",
  "/master-data/payment-terms",
  // Finance routes
  "/finance/accounting/coa",
  "/finance/accounting/journal-entries",
  "/finance/accounting/journal-entries/purchase",
  "/finance/accounting/journal-entries/adjustment",
  "/finance/fixed-assets",
  "/finance/settings",
  "/finance/bank-accounts",
  "/finance/bank-accounts/[id]",
  "/finance/journals/cash-bank",
  "/finance/ar/customer-invoices",
  "/finance/ar/customer-payments",
  "/finance/ar/credit-notes",
  "/finance/ar/aging-reports",
  "/finance/ap/supplier-invoices",
  "/finance/ap/supplier-payments",
  "/finance/ap/debit-notes",
  "/finance/ap/aging-reports",
  "/finance/ap/non-trade-payables",
  "/finance/cash-bank/reconciliation",
  "/finance/bank-transfer",
  "/finance/bank-reconciliation",
  "/finance/accounting/closing",
  "/finance/fixed-assets/assets",
  "/finance/fixed-assets/categories",
  "/finance/fixed-assets/locations",
  "/finance/fixed-assets/depreciation-schedule",
  "/finance/asset-batch-depreciation",
  "/finance/asset-disposal",
  "/finance/asset-revaluation",
  "/finance/settings/accounting-mapping",
  "/finance/settings/fiscal-years",
  "/finance/settings/tax-config",
  "/finance/settings/inventory",
  "/finance/settings/opening-balance",
  "/finance/reports/general-ledger",
  "/finance/reports/trial-balance",
  "/finance/reports/balance-sheet",
  "/finance/reports/profit-loss",
  "/finance/reports/cash-flow-statement",
  // Travel Planner routes
  "/travel/travel-planner",
  "/travel/visit-planner",
  // HRD routes
  "/hrd/attendance",
  "/hrd/leave-requests",
  "/hrd/overtime",
  "/hrd/evaluation",
  "/hrd/recruitment",
  "/hrd/work-schedule",
  "/hrd/holidays",
  "/hrd/salary-structures",
  // CRM
  "/crm/leads",
  "/crm/pipeline",
  "/crm/visits",
  "/crm/activities",
  "/crm/tasks",
  "/crm/schedules",
  "/crm/area-mapping",
  // CRM Settings
  "/crm/settings/pipeline-stages",
  "/crm/settings/lead-sources",
  "/crm/settings/lead-statuses",
  "/crm/settings/contact-roles",
  "/crm/settings/activity-types",
  // Reports
  "/reports",
  "/reports/sales-overview",
  "/reports/product-analysis",
  "/reports/geo-performance",
  "/reports/customer-research",
  "/reports/supplier-research",
  // AI Assistant
  "/ai-chatbot",
  "/ai-settings",
  // POS
  "/pos",
  "/pos/fb",
  "/pos/fb/floor-layout",
  "/pos/fb/terminal",
  "/pos/fb/live-table",
  "/pos/goods",
  "/pos/goods/quick-sale",
  "/pos/goods/basket",
  "/pos/goods/delivery-pickup",
  "/pos/goods/returns-exchange",
  "/pos/feedback/response",
  "/pos/feedback/forms",
  "/pos/loyalty",
  "/pos/loyalty/config",
  "/pos/loyalty/members",
] as const;

const VALID_DASHBOARD_ROUTES = Array.from(
  new Set([
    ...flattenNavigationRoutes(navigationConfig),
    ...EXTRA_VALID_ROUTES,
  ]),
).sort();

/**
 * Checks if a given route path is valid and exists in the application
 *
 * @param href - The route path to validate
 * @returns true if the route exists, false if it would result in 404
 */
export function isValidRoute(href: string | null | undefined): boolean {
  if (!href || href.trim() === "") {
    return false;
  }

  // Remove leading/trailing slashes and normalize path
  const normalizedPath = href.trim().replace(/^\/+|\/+$/g, "");

  // Empty path after normalization
  if (normalizedPath === "") {
    return false;
  }

  // Check if the path matches any valid route
  const pathWithSlash = `/${normalizedPath}`;
  return (VALID_DASHBOARD_ROUTES as readonly string[]).includes(pathWithSlash);
}

/**
 * Get all valid dashboard routes
 * Useful for debugging or generating route lists
 */
export function getValidRoutes(): readonly string[] {
  return VALID_DASHBOARD_ROUTES;
}
