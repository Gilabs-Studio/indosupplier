package migrations

import (
	"log"

	"gorm.io/gorm"
)

// DefaultTenantID is the UUID assigned to the initial (existing) tenant.
// All pre-existing data is backfilled to this tenant.
const DefaultTenantID = "a0000001-0000-0000-0000-000000000001"

// tenantTables lists every domain table that requires tenant_id.
// Geographic tables (countries, provinces, cities, districts, villages)
// and reference tables (time_zones, core_countries) are intentionally
// excluded because they are shared platform-wide.
var tenantTables = []string{
	// Auth & RBAC
	"users",
	"roles",
	"permissions",
	"menus",
	"role_permissions",
	"refresh_tokens",
	"password_reset_requests",

	// Core master data
	"audit_logs",
	"bank_accounts",
	"courier_agencies",
	"currencies",
	"leave_types",
	"payment_terms",
	"so_sources",

	// Organization
	"divisions",
	"job_positions",
	"business_units",
	"business_types",
	"areas",
	"companies",
	"outlets",
	"employees",
	"employee_areas",
	"employee_contracts",
	"employee_assets",
	"employee_certifications",
	"employee_education_histories",
	"employee_signatures",

	// Supplier
	"supplier_types",
	"banks",
	"suppliers",
	"supplier_contacts",
	"supplier_banks",

	// Customer
	"customer_types",
	"customers",
	"customer_banks",

	// Product
	"product_categories",
	"product_brands",
	"product_segments",
	"product_types",
	"units_of_measure",
	"packagings",
	"procurement_types",
	"products",
	"product_recipe_items",
	"product_recipe_versions",
	"product_recipe_version_items",

	// Warehouse
	"warehouses",
	"user_warehouses",

	// Purchase
	"purchase_requisitions",
	"purchase_requisition_items",
	"purchase_orders",
	"purchase_order_items",
	"goods_receipts",
	"goods_receipt_items",
	"supplier_invoices",
	"supplier_invoice_items",
	"purchase_payments",
	"purchase_returns",
	"purchase_return_items",

	// Sales
	"sales_quotations",
	"sales_quotation_items",
	"sales_orders",
	"sales_order_items",
	"delivery_orders",
	"delivery_order_items",
	"customer_invoices",
	"customer_invoice_items",
	"sales_returns",
	"sales_return_items",
	"sales_payments",
	"sales_visits",
	"sales_visit_details",
	"sales_visit_progress_histories",
	"sales_visit_interest_questions",
	"sales_visit_interest_options",
	"sales_visit_interest_answers",
	"yearly_targets",
	"monthly_targets",

	// Inventory
	"inventory_batches",
	"stock_movements",
	"stock_ledgers",

	// Stock Opname
	"stock_opnames",
	"stock_opname_items",

	// Finance
	"chart_of_accounts",
	"journal_entries",
	"journal_lines",
	"journal_templates",
	"journal_attachments",
	"journal_reversals",
	"adjustment_journal_approvals",
	"finance_settings",
	"payments",
	"payment_allocations",
	"budgets",
	"budget_items",
	"cash_bank_journals",
	"cash_bank_journal_lines",
	"cash_bank_transactions",
	"bank_transfers",
	"bank_reconciliations",
	"bank_statement_lines",
	"asset_categories",
	"asset_locations",
	"fixed_assets",
	"asset_depreciations",
	"asset_depreciation_schedules",
	"asset_disposals",
	"asset_transfers",
	"asset_maintenance_logs",
	"asset_transactions",
	"asset_attachments",
	"asset_audit_logs",
	"asset_assignment_histories",
	"fiscal_years",
	"accounting_periods",
	"financial_closings",
	"financial_closing_snapshots",
	"financial_closing_logs",
	"tax_configurations",
	"tax_invoices",
	"inventory_settings",
	"inventory_average_costs",
	"opening_balance_lines",
	"non_trade_payables",
	"salary_structures",
	"valuation_runs",
	"valuation_run_details",
	"up_country_costs",
	"up_country_cost_employees",
	"up_country_cost_items",
	"system_account_mappings",

	// HRD
	"work_schedules",
	"holidays",
	"attendance_records",
	"overtime_requests",
	"leave_requests",
	"evaluation_groups",
	"evaluation_criteria",
	"employee_evaluations",
	"employee_evaluation_criteria",
	"recruitment_requests",
	"recruitment_applicants",
	"applicant_stages",
	"applicant_activities",

	// CRM
	"crm_activity_types",
	"crm_contact_roles",
	"crm_contacts",
	"crm_leads",
	"crm_lead_product_items",
	"crm_deals",
	"crm_deal_product_items",
	"crm_deal_histories",
	"crm_pipeline_stages",
	"crm_lead_sources",
	"crm_lead_statuses",
	"crm_activities",
	"crm_tasks",
	"crm_reminders",
	"crm_schedules",
	"crm_visit_reports",
	"crm_visit_report_details",
	"crm_visit_report_progress_histories",
	"crm_visit_report_interest_answers",
	"crm_area_captures",

	// POS
	"pos_floor_plans",
	"pos_layout_versions",
	"pos_sessions",
	"pos_orders",
	"pos_order_items",
	"pos_payments",
	"pos_configs",
	"xendit_configs",
	"pos_table_status_records",

	// Travel Planner
	"travel_plans",
	"travel_plan_days",
	"travel_plan_stops",
	"travel_plan_day_notes",
	"travel_plan_expenses",

	// Loyalty
	"loyalty_programs",
	"loyalty_members",
	"loyalty_point_ledgers",

	// Feedback
	"feedback_forms",
	"feedback_tokens",
	"feedback_responses",

	// AI
	"ai_chat_sessions",
	"ai_chat_messages",
	"ai_action_logs",
	"ai_intent_registry",

	// General
	"dashboard_layouts",

	// Notification
	"notifications",
}

// MigrateTenantID adds the tenant_id column to all domain tables
// and backfills existing rows with the default tenant ID.
// This migration is idempotent — safe to run multiple times.
func MigrateTenantID(db *gorm.DB) error {
	log.Println("Running SaaS tenant_id migration...")

	for _, table := range tenantTables {
		// Add tenant_id column if it does not already exist
		addCol := `ALTER TABLE ` + table + ` ADD COLUMN IF NOT EXISTS tenant_id UUID`
		if err := db.Exec(addCol).Error; err != nil {
			log.Printf("Warning: could not add tenant_id to %s: %v", table, err)
			continue
		}

		// Backfill existing rows with the default tenant
		backfill := `UPDATE ` + table + ` SET tenant_id = ? WHERE tenant_id IS NULL`
		if err := db.Exec(backfill, DefaultTenantID).Error; err != nil {
			log.Printf("Warning: could not backfill tenant_id in %s: %v", table, err)
		}

		// Create index for tenant-scoped queries
		idxName := "idx_" + table + "_tenant_id"
		createIdx := `CREATE INDEX IF NOT EXISTS ` + idxName + ` ON ` + table + ` (tenant_id)`
		if err := db.Exec(createIdx).Error; err != nil {
			log.Printf("Warning: could not create index %s: %v", idxName, err)
		}
	}

	log.Println("SaaS tenant_id migration completed")
	return nil
}

// TenantTables returns the list of tables that have tenant_id.
// Useful for middleware/scoping logic.
func TenantTables() []string {
	return tenantTables
}
