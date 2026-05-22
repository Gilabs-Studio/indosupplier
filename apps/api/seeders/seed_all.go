package seeders

import "os"

// SeedAll runs all seeders
func SeedAll() error {
	if err := ensureSeederTenantIDGuard(); err != nil {
		return err
	}

	// Check for cleanup flag
	if os.Getenv("SEED_CLEANUP_DATABASE") == "true" {
		if err := CleanupDatabase(); err != nil {
			return err
		}
	}

	// Minimal seed mode: small, traceable dataset for debugging and validation.
	// if os.Getenv("SEED_MINIMAL_DATA") == "true" {
	// 	return seedMinimalData()
	// }

	if os.Getenv("SEED_ONLY_MASTER_DATA") == "true" {
		return seedMasterData()
	}

	// Seed in order: SaaS tenant -> roles -> menus -> permissions -> users -> geographic
	// SaaS: Create default tenant (must be first so tenant_id FK is available)
	if err := SeedTenant(); err != nil {
		return err
	}

	// SaaS: Create system admin (separate from tenant users)
	if err := SeedSystemAdmin(); err != nil {
		return err
	}

	// SaaS: Seed subscription plan configs (must come before coupons)
	if err := SeedSubscriptionPlans(); err != nil {
		return err
	}

	// SaaS: Seed coupons (must come before subscriptions)
	if err := SeedCoupons(); err != nil {
		return err
	}

	// SaaS: Seed tenant subscriptions (Tenant 1 = lifetime)
	if err := SeedSubscriptions(); err != nil {
		return err
	}

	if err := SeedRoles(); err != nil {
		return err
	}

	if err := SeedMenus(); err != nil {
		return err
	}

	// Update menu structure for existing menus (migration)
	if err := UpdateMenuStructure(); err != nil {
		return err
	}

	if err := SeedPermissions(); err != nil {
		return err
	}

	if err := SeedUsers(); err != nil {
		return err
	}

	// Geographic seeder (Sprint 1)
	if err := SeedGeographic(); err != nil {
		return err
	}

	// Organization seeder (Sprint 2)
	if err := SeedOrganization(); err != nil {
		return err
	}

	// Outlet seeder (depends on organization/company)
	if err := SeedOutlets(); err != nil {
		return err
	}

	// POS config seeder (stores receipt template in DB for every active outlet)
	if err := SeedPOSConfigs(); err != nil {
		return err
	}

	// POS floor layout seeder (2 outlets live-table scenario)
	if err := SeedPOSFloorLayouts(); err != nil {
		return err
	}

	// Employee seeder (Sprint 3)
	if err := SeedEmployees(); err != nil {
		return err
	}

	// Travel Planner seeder (depends on employee master data)
	if err := SeedTravelPlanner(); err != nil {
		return err
	}

	// Supplier seeder (Sprint 4)
	if err := SeedSupplier(); err != nil {
		return err
	}

	// Product seeder (Sprint 4)
	if err := SeedProduct(); err != nil {
		return err
	}

	// Warehouse seeder (Sprint 4)
	if err := SeedWarehouse(); err != nil {
		return err
	}

	// Inventory seeder (Sprint 4)
	if err := SeedInventory(); err != nil {
		return err
	}

	// POS F&B ingredient products (ingredient stock foundation)
	if err := SeedPosIngredients(); err != nil {
		return err
	}

	// POS F&B ingredient inventory batches (must run after SeedPosIngredients + SeedWarehouse)
	if err := SeedPosIngredientInventory(); err != nil {
		return err
	}

	// POS Goods products and starter stock batches
	if err := SeedPosGoodsProducts(); err != nil {
		return err
	}

	// POS Pharmacy products for Kimia Farma outlets (must run after SeedPosGoodsProducts)
	if err := SeedPosPharmaProducts(); err != nil {
		return err
	}

	// POS F&B recipe menu items, service products, and BOM recipe items
	if err := SeedPosRecipeProducts(); err != nil {
		return err
	}

	// Ensure product aggregate stock/HPP always matches seeded inventory batches.
	if err := ReconcileProductInventoryAggregates(); err != nil {
		return err
	}

	// User-warehouse assignments for POS outlet RBAC
	if err := SeedUserWarehouses(); err != nil {
		return err
	}

	// Master Data seeders (Sprint 4)
	if err := SeedCurrencies(); err != nil {
		return err
	}
	if err := SeedPaymentTerms(); err != nil {
		return err
	}
	if err := SeedCourierAgency(); err != nil {
		return err
	}
	if err := SeedSOSource(); err != nil {
		return err
	}
	if err := SeedLeaveType(); err != nil {
		return err
	}

	// ================================================================
	// FINANCE SEEDER - CRITICAL ORDER (DO NOT REORDER)
	// ================================================================
	// 1. Chart of Accounts (must come before Finance Settings)
	if err := SeedChartOfAccounts(); err != nil {
		return err
	}

	// 2. Finance Settings (maps all COA keys)
	if err := SeedFinanceSettings(); err != nil {
		return err
	}

	// 3. System Account Mappings (required by validation and accounting engine)
	if err := SeedSystemAccountMappings(); err != nil {
		return err
	}

	// 4. Validate integrity (fail-fast if broken)
	if err := ValidateFinanceSeeder(); err != nil {
		return err
	}

	// 5. Finance Foundation Phase 0 defaults for PT. GiLabs
	if err := SeedFinanceFoundationFiscalYear(); err != nil {
		return err
	}
	if err := SeedFinanceFoundationTaxConfiguration(); err != nil {
		return err
	}
	if err := SeedFinanceFoundationInventorySettings(); err != nil {
		return err
	}
	if err := SeedFinanceFoundationOpeningBalance(); err != nil {
		return err
	}
	// ================================================================

	// Bank account seed depends on canonical chart_of_accounts records.
	if err := SeedBankAccounts(); err != nil {
		return err
	}

	// Purchase Requisition seeder (Sprint 8)
	if err := SeedPurchaseRequisition(); err != nil {
		return err
	}

	// Customer Master Data seeder (must run before Sales seeders)
	if err := SeedCustomerTypes(); err != nil {
		return err
	}
	if err := SeedCustomers(); err != nil {
		return err
	}

	// Sales Quotation seeder (Sprint 5)
	if err := SeedSalesQuotation(); err != nil {
		return err
	}

	// Sales Order seeder (Sprint 6)
	if err := SeedSalesOrder(); err != nil {
		return err
	}

	// Delivery Order seeder (Sprint 6)
	if err := SeedDeliveryOrder(); err != nil {
		return err
	}

	// Customer Invoice seeder (Sprint 7)
	if err := SeedCustomerInvoice(); err != nil {
		return err
	}

	// Sales → Finance Integration Flow (SQ → SO → DO → INV → PAY)
	if err := SeedSalesIntegrationFlow(); err != nil {
		return err
	}

	// Finance - Asset & Closing seeder (Sprint 12)
	if err := SeedFinanceSprint12(); err != nil {
		return err
	}

	// Finance - Salary structures seeder
	if err := SeedSalaryStructures(); err != nil {
		return err
	}

	// Purchase → Finance E2E data (2025-2026) with correct business flows
	if err := SeedPurchaseFinanceE2E(); err != nil {
		return err
	}

	// Integration Flow seeder (Purchase → Stock → Sales → Finance)
	if err := SeedIntegrationFlow(); err != nil {
		return err
	}
	// Sales/Purchase Returns seeder (depends on invoices and goods receipts)
	if err := SeedReturns(); err != nil {
		return err
	}

	// Sales Visit seeder (Sprint 7)
	if err := SeedSalesVisitInterestSurvey(); err != nil {
		return err
	}

	if err := SeedSalesVisit(); err != nil {
		return err
	}

	// HRD - Work Schedules seeder (Sprint 13)
	if err := SeedWorkSchedules(); err != nil {
		return err
	}

	// HRD - Holidays seeder (Sprint 13)
	if err := SeedHolidays(); err != nil {
		return err
	}

	// HRD - Leave Requests seeder (Sprint 13)
	if err := SeedLeaveRequests(); err != nil {
		return err
	}

	// HRD - Attendance Records seeder (Sprint 13)
	if err := SeedAttendanceRecords(); err != nil {
		return err
	}

	// HRD - Overtime Requests seeder (Sprint 13)
	if err := SeedOvertimeRequests(); err != nil {
		return err
	}

	// HRD - Employee Contracts seeder (Sprint 14)
	if err := SeedEmployeeContracts(); err != nil {
		return err
	}

	// HRD - Employee Education History seeder (Sprint 14)
	if err := SeedEmployeeEducationHistory(); err != nil {
		return err
	}

	// HRD - Employee Certifications seeder (Sprint 14)
	if err := SeedEmployeeCertifications(); err != nil {
		return err
	}

	// HRD - Employee Assets seeder (Sprint 14)
	if err := SeedEmployeeAssets(); err != nil {
		return err
	}

	// HRD - Evaluation Groups seeder (Sprint 15)
	if err := SeedEvaluationGroups(); err != nil {
		return err
	}

	// HRD - Evaluation Criteria seeder (Sprint 15)
	if err := SeedEvaluationCriteria(); err != nil {
		return err
	}

	// HRD - Employee Evaluations seeder (Sprint 15)
	if err := SeedEmployeeEvaluations(); err != nil {
		return err
	}

	// HRD - Recruitment Requests seeder (Sprint 15)
	if err := SeedRecruitmentRequests(); err != nil {
		return err
	}

	// Stock Movement seeder (Sprint 9)
	if err := SeedStockMovement(); err != nil {
		return err
	}

	// Stock Opname seeder
	if err := SeedStockOpname(); err != nil {
		return err
	}

	// AI Intent Registry seeder
	if err := SeedAIIntentRegistry(); err != nil {
		return err
	}

	// CRM Settings seeder (Sprint 17)
	if err := SeedCRMSettings(); err != nil {
		return err
	}

	// CRM Contacts seeder (Sprint 18 - depends on customers + contact roles)
	if err := SeedCRMContacts(); err != nil {
		return err
	}

	// CRM Leads seeder (Sprint 19 - depends on lead sources, statuses, employees, customers)
	if err := SeedCRMLeads(); err != nil {
		return err
	}

	// CRM Deals seeder (Sprint 20 - depends on pipeline stages, customers, contacts, employees, products)
	if err := SeedCRMDeals(); err != nil {
		return err
	}

	// CRM Visit Reports seeder (Sprint 22 - depends on employees, customers, contacts, deals, leads)
	if err := SeedCRMVisitReports(); err != nil {
		return err
	}

	// CRM Activities, Tasks & Schedules seeder (Sprint 23 - depends on employees, customers, contacts, activity types)
	if err := SeedCRMActivitiesTasksSchedules(); err != nil {
		return err
	}

	// Feedback seeder (default forms per outlet for POS QR code feedback)
	if err := SeedFeedback(); err != nil {
		return err
	}

	// Loyalty seeder (default programs with Bronze/Silver/Gold/Platinum tiers)
	if err := SeedLoyalty(); err != nil {
		return err
	}

	// Loyalty member seeder (demo registered customers for POS lookup/redeem flows)
	if err := SeedLoyaltyMembers(); err != nil {
		return err
	}

	// Opening Balances seeder: Creates GL entries for inventory to match subledger
	// (Must come BEFORE SeedJournalReconciliation)
	if err := SeedOpeningBalances(); err != nil {
		return err
	}

	// Final Journal Reconciliation: ensures all transactional data (Sales, Purchase, Inventory, Returns)
	// has corresponding journal entries.
	if err := SeedJournalReconciliation(); err != nil {
		return err
	}

	// SaaS: Backfill tenant_id for all seeder data.
	// Must run LAST — ensures all rows created above get tenant_id = Tenant1.
	if err := BackfillSeedTenantIDs(); err != nil {
		return err
	}

	// Finance defaults are tenant-scoped; backfill them after all tenants exist.
	if err := BackfillTenantFinanceDefaults(); err != nil {
		return err
	}

	return nil
}

func seedMasterData() error {
	// Auth seeders
	if err := SeedRoles(); err != nil {
		return err
	}

	if err := SeedMenus(); err != nil {
		return err
	}

	if err := UpdateMenuStructure(); err != nil {
		return err
	}

	if err := SeedPermissions(); err != nil {
		return err
	}

	if err := SeedUsers(); err != nil {
		return err
	}

	// Master data seeders and required dependencies
	if err := SeedGeographic(); err != nil {
		return err
	}

	if err := SeedOrganization(); err != nil {
		return err
	}

	if err := SeedOutlets(); err != nil {
		return err
	}

	if err := SeedPOSFloorLayouts(); err != nil {
		return err
	}

	if err := SeedEmployees(); err != nil {
		return err
	}

	if err := SeedSupplier(); err != nil {
		return err
	}

	if err := SeedProduct(); err != nil {
		return err
	}

	if err := SeedWarehouse(); err != nil {
		return err
	}

	if err := SeedInventory(); err != nil {
		return err
	}

	if err := SeedPosGoodsProducts(); err != nil {
		return err
	}

	if err := ReconcileProductInventoryAggregates(); err != nil {
		return err
	}

	if err := SeedPaymentTerms(); err != nil {
		return err
	}

	if err := SeedCurrencies(); err != nil {
		return err
	}

	if err := SeedCourierAgency(); err != nil {
		return err
	}

	if err := SeedSOSource(); err != nil {
		return err
	}

	if err := SeedLeaveType(); err != nil {
		return err
	}

	if err := SeedChartOfAccounts(); err != nil {
		return err
	}

	if err := SeedCustomerTypes(); err != nil {
		return err
	}

	if err := SeedCustomers(); err != nil {
		return err
	}

	if err := SeedBankAccounts(); err != nil {
		return err
	}

	if err := SeedFinanceSettings(); err != nil {
		return err
	}

	if err := SeedSystemAccountMappings(); err != nil {
		return err
	}

	if err := ValidateFinanceSeeder(); err != nil {
		return err
	}

	if err := SeedFinanceFoundationFiscalYear(); err != nil {
		return err
	}

	if err := SeedFinanceFoundationTaxConfiguration(); err != nil {
		return err
	}

	if err := SeedFinanceFoundationInventorySettings(); err != nil {
		return err
	}

	if err := SeedFinanceFoundationOpeningBalance(); err != nil {
		return err
	}

	return nil
}
