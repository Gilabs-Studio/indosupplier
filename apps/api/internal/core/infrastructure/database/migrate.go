package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	ai "github.com/gilabs/gims/api/internal/ai/data/models"
	core "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database/migrations"
	crm "github.com/gilabs/gims/api/internal/crm/data/models"
	customer "github.com/gilabs/gims/api/internal/customer/data/models"
	feedback "github.com/gilabs/gims/api/internal/feedback/data/models"
	finance "github.com/gilabs/gims/api/internal/finance/data/models"
	general "github.com/gilabs/gims/api/internal/general/data/models"
	geographic "github.com/gilabs/gims/api/internal/geographic/data/models"
	hrd "github.com/gilabs/gims/api/internal/hrd/data/models"
	inventory "github.com/gilabs/gims/api/internal/inventory/data/models"
	loyalty "github.com/gilabs/gims/api/internal/loyalty/data/models"
	notification "github.com/gilabs/gims/api/internal/notification/data/models"
	organization "github.com/gilabs/gims/api/internal/organization/data/models"
	passwordReset "github.com/gilabs/gims/api/internal/password_reset/data/models"
	permission "github.com/gilabs/gims/api/internal/permission/data/models"
	pos "github.com/gilabs/gims/api/internal/pos/data/models"
	product "github.com/gilabs/gims/api/internal/product/data/models"
	purchase "github.com/gilabs/gims/api/internal/purchase/data/models"
	refreshToken "github.com/gilabs/gims/api/internal/refresh_token/data/models"
	role "github.com/gilabs/gims/api/internal/role/data/models"
	sales "github.com/gilabs/gims/api/internal/sales/data/models"
	stockOpname "github.com/gilabs/gims/api/internal/stock_opname/data/models"
	supplier "github.com/gilabs/gims/api/internal/supplier/data/models"
	tenant "github.com/gilabs/gims/api/internal/tenant/data/models"
	travelPlanner "github.com/gilabs/gims/api/internal/travel_planner/data/models"
	user "github.com/gilabs/gims/api/internal/user/data/models"
	warehouse "github.com/gilabs/gims/api/internal/warehouse/data/models"
)

// AutoMigrate runs database migrations
//
// PRODUCTION SAFETY:
// - This function is SAFE for production use
// - Tables are NEVER dropped in production mode (ENV=production)
// - Drop tables only happens in development mode when DROP_TABLES=true
// - Multiple safety checks prevent accidental data loss
// - No code changes needed for production deployment
func AutoMigrate() error {
	// Check if we should drop all tables (development only)
	// This check has built-in production protection
	if shouldDropTables() {
		log.Println("Development mode: Dropping all tables...")
		if err := DropAllTables(); err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
		log.Println("All tables dropped successfully")
	}

	// Try to handle constraint issues by attempting to drop constraints that might cause problems
	// This is a workaround for development environments where schema might be out of sync
	if err := handleConstraintIssues(); err != nil {
		log.Printf("Warning: Could not handle constraint issues (this may be expected): %v", err)
	}

	// PRE-MIGRATION FIX: Change reference_id to varchar to support non-UUID references
	// We ignore the error if the table doesn't exist yet (fresh install)
	_ = DB.Exec(`
		ALTER TABLE journal_entries 
		ALTER COLUMN reference_id TYPE VARCHAR(255) 
		USING reference_id::varchar;
	`)

	// PRE-MIGRATION: Add seat_limit and payment_transactions.tenant_id safely.
	// These must run BEFORE AutoMigrate so GORM sees consistent schema.
	_ = DB.Exec(`ALTER TABLE tenant_subscriptions ADD COLUMN IF NOT EXISTS seat_limit integer NOT NULL DEFAULT 1`)
	_ = DB.Exec(`UPDATE tenant_subscriptions SET seat_limit = GREATEST(user_count, 1) WHERE seat_limit = 0`)
	_ = DB.Exec(`ALTER TABLE tenant_subscriptions ADD COLUMN IF NOT EXISTS outlet_limit integer NOT NULL DEFAULT 1`)
	_ = DB.Exec(`UPDATE tenant_subscriptions SET outlet_limit = 1 WHERE outlet_limit <= 0`)
	_ = DB.Exec(`ALTER TABLE subscription_plan_configs ADD COLUMN IF NOT EXISTS outlet_addon_monthly_idr bigint NOT NULL DEFAULT 500000`)
	_ = DB.Exec(`ALTER TABLE subscription_plan_configs ADD COLUMN IF NOT EXISTS outlet_addon_yearly_idr bigint NOT NULL DEFAULT 6000000`)
	_ = DB.Exec(`UPDATE subscription_plan_configs SET outlet_addon_monthly_idr = 500000 WHERE outlet_addon_monthly_idr <= 0`)
	_ = DB.Exec(`UPDATE subscription_plan_configs SET outlet_addon_yearly_idr = 6000000 WHERE outlet_addon_yearly_idr <= 0`)
	_ = DB.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS deletion_requested_at timestamptz`)
	_ = DB.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS deletion_scheduled_at timestamptz`)
	_ = DB.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS deletion_requested_by uuid`)
	_ = DB.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS deletion_recovered_at timestamptz`)
	_ = DB.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS deletion_previous_status varchar(20)`)
	_ = DB.Exec(`ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS tenant_id uuid`)
	_ = DB.Exec(`CREATE INDEX IF NOT EXISTS idx_payment_transactions_tenant_id ON payment_transactions (tenant_id) WHERE tenant_id IS NOT NULL`)
	_ = DB.Exec(`ALTER TABLE pos_orders ADD COLUMN IF NOT EXISTS order_source varchar(20) NOT NULL DEFAULT 'STAFF'`)
	_ = DB.Exec(`UPDATE pos_orders SET order_source = 'STAFF' WHERE order_source IS NULL OR trim(order_source) = ''`)

	// PRE-MIGRATION: Replace is_pos_available boolean with pos_scope enum.
	// Add column first, then migrate data, so existing products are not lost.
	_ = DB.Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS pos_scope varchar(20) NOT NULL DEFAULT 'none'`)
	_ = DB.Exec(`UPDATE products SET pos_scope = CASE WHEN is_pos_available = true THEN 'global' ELSE 'none' END WHERE pos_scope = 'none' AND is_pos_available IS NOT NULL`)

	// Perform actual migrations
	err := migrateWithErrorHandling(
		// SaaS: Tenant & System Admin tables (must come before domain tables)
		&tenant.Tenant{},
		&tenant.SystemAdmin{},
		&tenant.Coupon{},
		&tenant.CouponUsage{},
		&tenant.SubscriptionPlanConfig{},
		&tenant.PlanModuleEntitlement{},
		&tenant.PlanPermissionEntitlement{},
		&tenant.TenantSubscription{},
		&tenant.PaymentTransaction{},
		&tenant.TenantPaymentMethod{},
		&user.User{},
		&user.UserWarehouse{},
		&role.RolePermission{},
		&role.RoleMenuAccess{},
		&role.Role{},
		&permission.Permission{},
		&permission.Menu{},
		&refreshToken.RefreshToken{},
		&passwordReset.PasswordResetRequest{},
		&core.AuditLog{},
		&notification.Notification{},
		// Geographic entities (Sprint 1)
		&geographic.Country{},
		&geographic.Province{},
		&geographic.City{},
		&geographic.District{},
		&geographic.Village{},
		// Timezone data for auto-detection
		&core.TimeZone{},
		&core.Country{},
		// Organization entities (Sprint 2)
		&organization.Division{},
		&organization.JobPosition{},
		&organization.BusinessUnit{},
		&organization.BusinessType{},
		&organization.Area{},
		// NOTE: AreaSupervisor and AreaSupervisorArea removed in Sprint 17.
		// Supervisor role is now captured via EmployeeArea.IsSupervisor flag.
		&organization.Company{},
		// Employee entities (Sprint 3)
		&organization.Employee{},
		&organization.EmployeeArea{},
		&organization.EmployeeOutlet{},
		&organization.EmployeeWarehouse{},
		// Supplier entities (Sprint 4)
		&supplier.SupplierType{},
		&supplier.Bank{},
		&supplier.Supplier{},
		&supplier.SupplierContact{},
		&supplier.SupplierBank{},
		// Customer entities (Master Data)
		&customer.CustomerType{},
		&customer.Customer{},
		&customer.CustomerBank{},
		// Product entities (Sprint 4)
		&product.ProductCategory{},
		&product.ProductBrand{},
		&product.ProductSegment{},
		&product.ProductType{},
		&product.UnitOfMeasure{},
		&product.Packaging{},
		&product.ProcurementType{},
		&product.Product{},
		&product.ProductRecipeItem{},
		&product.ProductRecipeVersion{},
		&product.ProductRecipeVersionItem{},
		// Warehouse entities (Sprint 4)
		&warehouse.Warehouse{},
		// Outlet entity (Organization)
		&organization.Outlet{},
		// Product-Outlet junction (depends on both Product and Outlet)
		&product.ProductOutlet{},
		// Core Master Data entities (Sprint 4)
		&core.PaymentTerms{},
		&core.CourierAgency{},
		&core.SOSource{},
		&core.LeaveType{},
		&core.Currency{},
		&core.BankAccount{},
		// Finance entities (Sprint 10)
		&finance.ChartOfAccount{},
		&finance.FinanceSetting{},
		&finance.JournalEntry{},
		&finance.JournalLine{},
		&finance.JournalReversal{},
		&finance.JournalAttachment{},
		&finance.AdjustmentJournalApproval{},
		&finance.JournalTemplate{},
		// Finance entities (Sprint 11)
		&finance.Payment{},
		&finance.PaymentAllocation{},
		&finance.Budget{},
		&finance.BudgetItem{},
		&finance.CashBankJournal{},
		&finance.CashBankJournalLine{},
		&finance.CashBankTransaction{},
		&finance.BankTransfer{},
		&finance.BankReconciliation{},
		&finance.BankStatementLine{},
		// Finance entities (Sprint 12)
		&finance.AssetCategory{},
		&finance.AssetLocation{},
		&finance.Asset{},
		&finance.AssetDepreciation{},
		&finance.AssetDepreciationSchedule{},
		&finance.AssetTransaction{},
		// Extended Asset domain entities (final business records)
		&finance.AssetDisposal{},
		&finance.AssetTransfer{},
		&finance.AssetRevaluation{},
		&finance.AssetMaintenanceLog{},
		// Asset workflow & audit
		&finance.AssetAttachment{},
		&finance.AssetAuditLog{},
		&finance.AssetAssignmentHistory{},
		&finance.FinancialClosing{},
		&finance.FiscalYear{},
		&finance.AccountingPeriod{},
		&finance.FinancialClosingSnapshot{},
		&finance.FinancialClosingLog{},
		&finance.TaxConfiguration{},
		&finance.InventorySettings{},
		&finance.InventoryAverageCost{},
		&finance.OpeningBalanceLine{},
		&finance.TaxInvoice{},
		&finance.NonTradePayable{},
		&hrd.SalaryStructure{},
		&finance.ValuationRun{},
		&finance.ValuationRunDetail{},
		&finance.UpCountryCost{},
		&finance.UpCountryCostEmployee{},
		&finance.UpCountryCostItem{},
		&finance.SystemAccountMapping{},
		// Travel Planner entities
		&travelPlanner.TravelPlan{},
		&travelPlanner.TravelPlanDay{},
		&travelPlanner.TravelPlanStop{},
		&travelPlanner.TravelPlanDayNote{},
		&travelPlanner.TravelPlanExpense{},
		// Sales entities (Sprint 5)
		&sales.SalesQuotation{},
		&sales.SalesQuotationItem{},
		// Sales Order entities (Sprint 6)
		&sales.SalesOrder{},
		&sales.SalesOrderItem{},
		// Delivery Order entities (Sprint 6)
		&sales.DeliveryOrder{},
		&sales.DeliveryOrderItem{},
		// Customer Invoice entities (Sprint 7)
		&sales.CustomerInvoice{},
		&sales.CustomerInvoiceItem{},
		&sales.SalesReturn{},
		&sales.SalesReturnItem{},
		// Sales Payment entities
		&sales.SalesPayment{},
		// Sales Visit entities (Sprint 7)
		&sales.SalesVisit{},
		&sales.SalesVisitDetail{},
		&sales.SalesVisitProgressHistory{},
		// Sales Targets per Employee (Sales Target feature)
		&sales.SalesTarget{},
		&sales.MonthlySalesTarget{},
		// Sales Visit Interest Survey (Sprint 7)
		&sales.SalesVisitInterestQuestion{},
		&sales.SalesVisitInterestOption{},
		&sales.SalesVisitInterestAnswer{},
		// HRD Attendance entities (Sprint 13)
		&hrd.WorkSchedule{},
		&hrd.Holiday{},
		&hrd.AttendanceRecord{},
		&hrd.OvertimeRequest{},
		// HRD Leave Management entities (Sprint 14)
		&hrd.LeaveRequest{},
		// Organization Employee Contracts entities (moved from HRD)
		&organization.EmployeeContract{},
		// Organization Employee Education History entities (moved from HRD)
		&organization.EmployeeEducationHistory{},
		// Organization Employee Certifications entities (moved from HRD)
		&organization.EmployeeCertification{},
		// Organization Employee Assets entities (moved from HRD)
		&organization.EmployeeAsset{},
		// HRD Evaluation entities (Sprint 15)
		&hrd.EvaluationGroup{},
		&hrd.EvaluationCriteria{},
		&hrd.EmployeeEvaluation{},
		&hrd.EmployeeEvaluationCriteria{},
		// HRD Recruitment entities (Sprint 15)
		&hrd.RecruitmentRequest{},
		// HRD Recruitment Applicant entities
		&hrd.RecruitmentApplicant{},
		&hrd.ApplicantStage{},
		&hrd.ApplicantActivity{},
		// Organization Employee Signature
		&organization.EmployeeSignature{},
		// Inventory entities (Sprint 9)
		&inventory.InventoryBatch{},
		&inventory.StockMovement{},
		&inventory.StockLedger{},
		// Stock Opname entities (Sprint 9)
		&stockOpname.StockOpname{},
		&stockOpname.StockOpnameItem{},
		// Purchase entities (Sprint 8)
		&purchase.PurchaseRequisition{},
		&purchase.PurchaseRequisitionItem{},
		&purchase.PurchaseOrder{},
		&purchase.PurchaseOrderItem{},
		&purchase.GoodsReceipt{},
		&purchase.GoodsReceiptItem{},
		&purchase.SupplierInvoice{},
		&purchase.SupplierInvoiceItem{},
		&purchase.PurchasePayment{},
		&purchase.PurchaseReturn{},
		&purchase.PurchaseReturnItem{},
		// AI Assistant entities
		&ai.AIChatSession{},
		&ai.AIChatMessage{},
		&ai.AIActionLog{},
		&ai.AIIntentRegistry{},
		// CRM Settings entities (Sprint 17)
		&crm.PipelineStage{},
		&crm.LeadSource{},
		&crm.LeadStatus{},
		&crm.ContactRole{},
		&crm.ActivityType{},
		// CRM Contact entity (Sprint 18)
		&crm.Contact{},
		// CRM Lead entity (Sprint 19)
		&crm.Lead{},
		&crm.LeadProductItem{},
		// CRM Deal entities (Sprint 20)
		&crm.Deal{},
		&crm.DealProductItem{},
		&crm.DealHistory{},
		// CRM Visit Report entities (Sprint 22)
		&crm.VisitReport{},
		&crm.VisitReportDetail{},
		&crm.VisitReportProgressHistory{},
		&crm.VisitReportInterestAnswer{},
		// CRM Activity, Task & Schedule entities (Sprint 23)
		&crm.Activity{},
		&crm.Task{},
		&crm.Reminder{},
		&crm.Schedule{},
		// CRM Area Mapping entities (Sprint 24)
		&crm.AreaCapture{},
		// General: user dashboard layout preferences
		&general.DashboardLayout{},
		// General: tenant onboarding state
		&general.TenantOnboarding{},
		// POS entities (Floor Layout Designer + Full POS System)
		&pos.FloorPlan{},
		&pos.LayoutVersion{},
		&pos.PosTableQRToken{},
		&pos.PosSession{},
		&pos.PosOrder{},
		&pos.PosOrderItem{},
		&pos.POSPayment{},
		&pos.POSConfig{},
		&pos.XenditConfig{},
		&pos.PosTableStatusRecord{},
		&pos.POSDeviceToken{},
		// Feedback module
		&feedback.FeedbackForm{},
		&feedback.FeedbackToken{},
		&feedback.FeedbackResponse{},
		// Loyalty module
		&loyalty.LoyaltyProgram{},
		&loyalty.LoyaltyMember{},
		&loyalty.LoyaltyPointLedger{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed")

	// SaaS: Add tenant_id to all domain tables and backfill existing data
	if err := migrations.MigrateTenantID(DB); err != nil {
		return fmt.Errorf("failed to run SaaS tenant migration: %w", err)
	}

	// Normalize lifetime subscriptions created by older coupon logic
	if err := migrations.NormalizeLifetimeSubscriptions(DB); err != nil {
		log.Printf("Warning: Could not normalize lifetime subscriptions: %v", err)
	}

	// FIX: Ensure RemainingAmount is initialized for all invoices (Customer & Supplier)
	// This fixes issues where seeders or manual imports missed the remaining amount.
	DB.Exec(`UPDATE customer_invoices 
             SET remaining_amount = amount - paid_amount 
             WHERE (remaining_amount = 0 OR remaining_amount IS NULL) 
             AND amount > 0 AND paid_amount < amount`)
	DB.Exec(`UPDATE supplier_invoices 
             SET remaining_amount = amount - paid_amount 
             WHERE (remaining_amount = 0 OR remaining_amount IS NULL) 
             AND amount > 0 AND paid_amount < amount`)

	// Migrate contract data from employees table to employee_contracts table
	if err := migrateEmployeeContractData(); err != nil {
		log.Printf("Warning: Could not migrate employee contract data: %v", err)
	}

	// Migrate legacy POS floor plans from company scope to outlet scope.
	if err := migratePOSFloorPlansToOutletScope(); err != nil {
		log.Printf("Warning: Could not migrate POS floor plans to outlet scope: %v", err)
	}

	// Sessionless POS path: existing databases may still have NOT NULL on session_id.
	if err := ensurePOSOrderSessionColumnNullable(); err != nil {
		log.Printf("Warning: Could not relax pos_orders.session_id nullability: %v", err)
	}

	// Live table object IDs are string identifiers (e.g. tbl-a1), not UUID.
	if err := ensurePOSOrderTableIDColumnType(); err != nil {
		log.Printf("Warning: Could not migrate pos_orders.table_id column type: %v", err)
	}

	// Safety net: ensure role_permissions.scope column exists even when GORM's
	// AutoMigrate did not add it (e.g. the many2many relationship on Role
	// created the table first without the scope column).
	if err := DB.Exec(`
		ALTER TABLE role_permissions
		ADD COLUMN IF NOT EXISTS scope VARCHAR(20) NOT NULL DEFAULT 'ALL'
	`).Error; err != nil {
		log.Printf("Warning: could not ensure role_permissions.scope column: %v", err)
	}

	// Ensure multi-company uniqueness for system account mappings.
	if err := DB.Exec(`DROP INDEX IF EXISTS idx_system_account_mappings_key`).Error; err != nil {
		log.Printf("Warning: could not drop legacy system_account_mappings key index: %v", err)
	}
	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_system_account_mappings_key_company
		ON system_account_mappings ("key", company_id)
	`).Error; err != nil {
		log.Printf("Warning: could not ensure composite index idx_system_account_mappings_key_company: %v", err)
	}

	// Sprint 17: Migrate area_supervisors data to employee_areas
	if err := migrateAreaSupervisorsToEmployeeAreas(); err != nil {
		log.Printf("Warning: Area supervisor migration skipped or failed: %v", err)
	}

	// Safety net for rollout compatibility: older databases may miss the
	// newly introduced goods_receipts.warehouse_id column.
	if err := ensureGoodsReceiptWarehouseColumn(); err != nil {
		return fmt.Errorf("failed to ensure goods receipt warehouse column: %w", err)
	}

	// Safety net for rollout compatibility: older databases may miss
	// customer_contact_id columns introduced in sales documents.
	if err := ensureSalesCustomerContactColumns(); err != nil {
		return fmt.Errorf("failed to ensure sales customer contact columns: %w", err)
	}

	// Create search indexes for performance
	if err := createSearchIndexes(); err != nil {
		log.Printf("Warning: Failed to create search indexes (this is non-fatal): %v", err)
	}

	// Create triggers to enforce closed accounting periods on journal entries
	if err := createJournalEntryPeriodLockTrigger(); err != nil {
		log.Printf("Warning: Failed to create journal entry period lock trigger (this is non-fatal): %v", err)
	}

	// Migrate timezone data for auto-detection (using longitude-based detection for Indonesia)
	if err := migrateTimezoneData(); err != nil {
		log.Printf("Warning: Could not migrate timezone data: %v", err)
	}

	// Remove status column from employee_evaluations table (Sprint 16)
	if err := migrations.RemoveEvaluationStatusColumn(DB); err != nil {
		log.Printf("Warning: Could not remove evaluation status column: %v", err)
	}

	// Remove days_requested column from leave_requests table (Sprint 14)
	// WHY: Consolidate to using TotalDays only with inclusive calendar days calculation
	if err := migrations.RemoveLeaveRequestDaysRequestedMigration(DB); err != nil {
		log.Printf("Warning: Could not remove leave request days_requested column: %v", err)
	}

	// Add linkedin_url column to recruitment_applicants table (Sprint 15)
	// WHY: Allow storing LinkedIn profile URLs for applicants
	if err := migrations.AddApplicantLinkedInURLMigration(DB); err != nil {
		log.Printf("Warning: Could not add linkedin_url column: %v", err)
	}

	// Add menu metadata fields required by canonical module navigation refactor.
	if err := migrations.AddMenuMetadataMigration(DB); err != nil {
		return fmt.Errorf("failed to add menu metadata migration: %w", err)
	}

	// Relax area location columns so users can store free-form city/region text.
	if err := migrations.EnsureAreaLocationTextMigration(DB); err != nil {
		log.Printf("Warning: failed to relax area location columns: %v", err)
	}

	// Performance indexes for slow list endpoints (identified by k6 profiler)
	if err := migrations.AddPerformanceIndexesMigration(DB); err != nil {
		log.Printf("Warning: failed to add performance indexes: %v", err)
	}

	// NEW: Normalize journal data (casing/consistent naming)
	if err := normalizeJournalData(); err != nil {
		log.Printf("Warning: Failed to normalize journal data: %v", err)
	}

	// SaaS: Backfill tenant_id for all existing rows that have a NULL tenant_id.
	// New rows will have tenant_id set by the application layer (via model struct or handler).
	// Existing seeded data in development is assigned to DefaultTenantID so it remains visible.
	if err := backfillTenantIDs(); err != nil {
		log.Printf("Warning: tenant_id backfill failed (non-fatal): %v", err)
	}

	if err := ensureCOACodeTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: COA tenant-scoped unique index migration failed (non-fatal): %v", err)
	}

	if err := ensureRoleNameTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: role tenant-scoped name index migration failed (non-fatal): %v", err)
	}

	if err := ensureAreaTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: area tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureDivisionTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: division tenant-scoped unique index migration failed (non-fatal): %v", err)
	}

	if err := ensureJobPositionTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: job position tenant-scoped unique index migration failed (non-fatal): %v", err)
	}

	if err := ensureEmployeeCodeTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: employee code tenant-scoped unique index migration failed (non-fatal): %v", err)
	}

	if err := ensureCustomerCodeTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: customer code tenant-scoped unique index migration failed (non-fatal): %v", err)
	}

	if err := ensureSupplierCodeTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: supplier code tenant-scoped unique index migration failed (non-fatal): %v", err)
	}

	if err := ensureBankCodeTenantScopedUniqueIndex(); err != nil {
		log.Printf("Warning: bank code tenant-scoped unique index migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMContactRoleTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM contact role tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMPipelineStageTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM pipeline stage tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMLeadSourceTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM lead source tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMLeadStatusTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM lead status tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMActivityTypeTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM activity type tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMLeadsTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM leads tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMVisitReportsTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM visit reports tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureCRMDealsTenantScopedUniqueIndexes(); err != nil {
		log.Printf("Warning: CRM deals tenant-scoped unique indexes migration failed (non-fatal): %v", err)
	}

	if err := ensureFixedAssetListIndexes(); err != nil {
		log.Printf("Warning: fixed asset list index migration failed (non-fatal): %v", err)
	}

	if err := ensureDefaultAreasForAllTenants(); err != nil {
		log.Printf("Warning: failed to ensure default areas for existing tenants: %v", err)
	}

	if err := ensureDefaultBanksForAllTenants(); err != nil {
		log.Printf("Warning: failed to ensure default banks for existing tenants: %v", err)
	}

	if err := dropCompanyDirectorColumnIfExists(); err != nil {
		log.Printf("Warning: failed to drop deprecated companies.director_id column: %v", err)
	}

	if err := syncGeneratedTenantRolesFromPlanTemplates(); err != nil {
		log.Printf("Warning: generated tenant role sync from plan templates failed (non-fatal): %v", err)
	}

	return nil
}

func ensureRoleNameTenantScopedUniqueIndex() error {
	// Drop legacy global unique constraints/indexes on roles.name if they exist.
	_ = DB.Exec(`ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_name_key`).Error
	_ = DB.Exec(`ALTER TABLE roles DROP CONSTRAINT IF EXISTS idx_roles_name`).Error
	_ = DB.Exec(`ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_name_unique`).Error
	_ = DB.Exec(`ALTER TABLE roles DROP CONSTRAINT IF EXISTS uq_roles_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_roles_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_roles_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS roles_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS roles_name_unique`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS uq_roles_name`).Error

	// Dynamic cleanup for any leftover unique constraints/indexes on roles.name
	// across schemas / custom naming conventions from previous versions.
	_ = DB.Exec(`
		DO $$
		DECLARE
			constraint_row RECORD;
			index_row RECORD;
		BEGIN
			FOR constraint_row IN
				SELECT n.nspname AS schema_name, c.relname AS table_name, con.conname AS constraint_name
				FROM pg_constraint con
				JOIN pg_class c ON c.oid = con.conrelid
				JOIN pg_namespace n ON n.oid = c.relnamespace
				WHERE con.contype = 'u'
				  AND c.relname = 'roles'
				  AND EXISTS (
					SELECT 1
					FROM unnest(con.conkey) AS key_col(attnum)
					JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = key_col.attnum
					WHERE a.attname = 'name'
				  )
			LOOP
				EXECUTE format(
					'ALTER TABLE %I.%I DROP CONSTRAINT IF EXISTS %I',
					constraint_row.schema_name,
					constraint_row.table_name,
					constraint_row.constraint_name,
				);
			END LOOP;

			FOR index_row IN
				SELECT schemaname, indexname
				FROM pg_indexes
				WHERE tablename = 'roles'
				  AND lower(indexdef) LIKE '%unique%'
				  AND lower(indexdef) LIKE '%name%'
			LOOP
				EXECUTE format('DROP INDEX IF EXISTS %I.%I', index_row.schemaname, index_row.indexname);
			END LOOP;
		END $$;
	`).Error

	// Enforce uniqueness by tenant + name (case-insensitive), while keeping soft-deletes reusable.
	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_tenant_name_active
		ON roles ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error
}

func ensureCOACodeTenantScopedUniqueIndex() error {
	_ = DB.Exec(`ALTER TABLE chart_of_accounts DROP CONSTRAINT IF EXISTS chart_of_accounts_code_key`).Error
	_ = DB.Exec(`ALTER TABLE chart_of_accounts DROP CONSTRAINT IF EXISTS idx_chart_of_accounts_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_chart_of_accounts_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_chart_of_accounts_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS chart_of_accounts_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.chart_of_accounts_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_chart_of_accounts_tenant_code_active
		ON chart_of_accounts (tenant_id, code)
	`).Error
}

func ensureAreaTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE areas DROP CONSTRAINT IF EXISTS areas_name_key`).Error
	_ = DB.Exec(`ALTER TABLE areas DROP CONSTRAINT IF EXISTS idx_areas_name`).Error
	_ = DB.Exec(`ALTER TABLE areas DROP CONSTRAINT IF EXISTS areas_code_key`).Error
	_ = DB.Exec(`ALTER TABLE areas DROP CONSTRAINT IF EXISTS idx_areas_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_areas_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_areas_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS areas_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.areas_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_areas_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_areas_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS areas_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.areas_code_key`).Error

	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_areas_tenant_name_active
		ON areas ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_areas_tenant_code_active
		ON areas ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureDivisionTenantScopedUniqueIndex() error {
	_ = DB.Exec(`ALTER TABLE divisions DROP CONSTRAINT IF EXISTS divisions_name_key`).Error
	_ = DB.Exec(`ALTER TABLE divisions DROP CONSTRAINT IF EXISTS idx_divisions_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_divisions_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_divisions_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS divisions_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.divisions_name_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_divisions_tenant_name_active
		ON divisions ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error
}

func ensureJobPositionTenantScopedUniqueIndex() error {
	_ = DB.Exec(`ALTER TABLE job_positions DROP CONSTRAINT IF EXISTS job_positions_name_key`).Error
	_ = DB.Exec(`ALTER TABLE job_positions DROP CONSTRAINT IF EXISTS idx_job_positions_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_job_positions_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_job_positions_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS job_positions_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.job_positions_name_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_job_positions_tenant_name_active
		ON job_positions ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error
}

func ensureEmployeeCodeTenantScopedUniqueIndex() error {
	_ = DB.Exec(`ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_employee_code_key`).Error
	_ = DB.Exec(`ALTER TABLE employees DROP CONSTRAINT IF EXISTS idx_employees_employee_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_employees_employee_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_employees_employee_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS employees_employee_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.employees_employee_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_employees_tenant_code_active
		ON employees ((COALESCE(tenant_id::text, '__global__')), lower(employee_code))
		WHERE deleted_at IS NULL
	`).Error
}

func ensureCustomerCodeTenantScopedUniqueIndex() error {
	_ = DB.Exec(`ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_code_key`).Error
	_ = DB.Exec(`ALTER TABLE customers DROP CONSTRAINT IF EXISTS idx_customers_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_customers_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_customers_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS customers_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.customers_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_tenant_code_active
		ON customers ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureSupplierCodeTenantScopedUniqueIndex() error {
	_ = DB.Exec(`ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS suppliers_code_key`).Error
	_ = DB.Exec(`ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS idx_suppliers_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_suppliers_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_suppliers_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS suppliers_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.suppliers_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_suppliers_tenant_code_active
		ON suppliers ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureBankCodeTenantScopedUniqueIndex() error {
	_ = DB.Exec(`ALTER TABLE banks DROP CONSTRAINT IF EXISTS banks_code_key`).Error
	_ = DB.Exec(`ALTER TABLE banks DROP CONSTRAINT IF EXISTS idx_banks_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_banks_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_banks_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS banks_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.banks_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_banks_tenant_code_active
		ON banks ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMContactRoleTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_contact_roles DROP CONSTRAINT IF EXISTS crm_contact_roles_name_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_contact_roles DROP CONSTRAINT IF EXISTS idx_crm_contact_roles_name`).Error
	_ = DB.Exec(`ALTER TABLE crm_contact_roles DROP CONSTRAINT IF EXISTS crm_contact_roles_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_contact_roles DROP CONSTRAINT IF EXISTS idx_crm_contact_roles_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_contact_roles_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_contact_roles_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_contact_roles_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_contact_roles_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_contact_roles_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_contact_roles_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_contact_roles_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_contact_roles_code_key`).Error

	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_contact_roles_tenant_name_active
		ON crm_contact_roles ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_contact_roles_tenant_code_active
		ON crm_contact_roles ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMPipelineStageTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_pipeline_stages DROP CONSTRAINT IF EXISTS crm_pipeline_stages_name_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_pipeline_stages DROP CONSTRAINT IF EXISTS idx_crm_pipeline_stages_name`).Error
	_ = DB.Exec(`ALTER TABLE crm_pipeline_stages DROP CONSTRAINT IF EXISTS crm_pipeline_stages_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_pipeline_stages DROP CONSTRAINT IF EXISTS idx_crm_pipeline_stages_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_pipeline_stages_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_pipeline_stages_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_pipeline_stages_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_pipeline_stages_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_pipeline_stages_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_pipeline_stages_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_pipeline_stages_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_pipeline_stages_code_key`).Error

	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_pipeline_stages_tenant_name_active
		ON crm_pipeline_stages ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_pipeline_stages_tenant_code_active
		ON crm_pipeline_stages ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMLeadSourceTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_lead_sources DROP CONSTRAINT IF EXISTS crm_lead_sources_name_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_lead_sources DROP CONSTRAINT IF EXISTS idx_crm_lead_sources_name`).Error
	_ = DB.Exec(`ALTER TABLE crm_lead_sources DROP CONSTRAINT IF EXISTS crm_lead_sources_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_lead_sources DROP CONSTRAINT IF EXISTS idx_crm_lead_sources_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_lead_sources_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_lead_sources_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_lead_sources_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_lead_sources_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_lead_sources_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_lead_sources_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_lead_sources_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_lead_sources_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_lead_sources_tenant_code_active
		ON crm_lead_sources ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMLeadStatusTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_lead_statuses DROP CONSTRAINT IF EXISTS crm_lead_statuses_name_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_lead_statuses DROP CONSTRAINT IF EXISTS idx_crm_lead_statuses_name`).Error
	_ = DB.Exec(`ALTER TABLE crm_lead_statuses DROP CONSTRAINT IF EXISTS crm_lead_statuses_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_lead_statuses DROP CONSTRAINT IF EXISTS idx_crm_lead_statuses_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_lead_statuses_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_lead_statuses_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_lead_statuses_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_lead_statuses_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_lead_statuses_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_lead_statuses_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_lead_statuses_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_lead_statuses_code_key`).Error

	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_lead_statuses_tenant_name_active
		ON crm_lead_statuses ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_lead_statuses_tenant_code_active
		ON crm_lead_statuses ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMActivityTypeTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_activity_types DROP CONSTRAINT IF EXISTS crm_activity_types_name_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_activity_types DROP CONSTRAINT IF EXISTS idx_crm_activity_types_name`).Error
	_ = DB.Exec(`ALTER TABLE crm_activity_types DROP CONSTRAINT IF EXISTS crm_activity_types_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_activity_types DROP CONSTRAINT IF EXISTS idx_crm_activity_types_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_activity_types_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_activity_types_name`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_activity_types_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_activity_types_name_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_activity_types_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_activity_types_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_activity_types_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_activity_types_code_key`).Error

	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_activity_types_tenant_name_active
		ON crm_activity_types ((COALESCE(tenant_id::text, '__global__')), lower(name))
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_activity_types_tenant_code_active
		ON crm_activity_types ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMLeadsTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_leads DROP CONSTRAINT IF EXISTS crm_leads_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_leads DROP CONSTRAINT IF EXISTS idx_crm_leads_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_leads_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_leads_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_leads_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_leads_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_leads_tenant_code_active
		ON crm_leads ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMVisitReportsTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_visit_reports DROP CONSTRAINT IF EXISTS crm_visit_reports_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_visit_reports DROP CONSTRAINT IF EXISTS idx_crm_visit_reports_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_visit_reports_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_visit_reports_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_visit_reports_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_visit_reports_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_visit_reports_tenant_code_active
		ON crm_visit_reports ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureCRMDealsTenantScopedUniqueIndexes() error {
	_ = DB.Exec(`ALTER TABLE crm_deals DROP CONSTRAINT IF EXISTS crm_deals_code_key`).Error
	_ = DB.Exec(`ALTER TABLE crm_deals DROP CONSTRAINT IF EXISTS idx_crm_deals_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_crm_deals_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_crm_deals_code`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS crm_deals_code_key`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.crm_deals_code_key`).Error

	return DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_crm_deals_tenant_code_active
		ON crm_deals ((COALESCE(tenant_id::text, '__global__')), lower(code))
		WHERE deleted_at IS NULL AND COALESCE(code, '') <> ''
	`).Error
}

func ensureFixedAssetListIndexes() error {
	statements := []string{
		`CREATE INDEX IF NOT EXISTS idx_fixed_assets_tenant_status_active ON fixed_assets (tenant_id, status) WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_fixed_assets_tenant_category_active ON fixed_assets (tenant_id, category_id) WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_fixed_assets_tenant_warranty_end_active ON fixed_assets (tenant_id, warranty_end) WHERE deleted_at IS NULL`,
	}

	for _, statement := range statements {
		if err := DB.Exec(statement).Error; err != nil {
			return err
		}
	}

	return nil
}

func dropCompanyDirectorColumnIfExists() error {
	_ = DB.Exec(`DROP INDEX IF EXISTS idx_companies_director_id`).Error
	_ = DB.Exec(`DROP INDEX IF EXISTS public.idx_companies_director_id`).Error

	return DB.Exec(`ALTER TABLE companies DROP COLUMN IF EXISTS director_id`).Error
}

type areaDefaultSeed struct {
	Name        string
	Description string
	Code        string
	Color       string
	Province    string
}

type bankDefaultSeed struct {
	Name      string
	Code      string
	SwiftCode string
}

var defaultAreaSeeds = []areaDefaultSeed{}

var defaultBankSeeds = []bankDefaultSeed{
	{Name: "Bank Central Asia", Code: "BCA", SwiftCode: "CENAIDJA"},
	{Name: "Bank Mandiri", Code: "MANDIRI", SwiftCode: "BMRIIDJA"},
	{Name: "Bank Negara Indonesia", Code: "BNI", SwiftCode: "BNIAIDJAXXX"},
	{Name: "Bank Rakyat Indonesia", Code: "BRI", SwiftCode: "BRINIDJA"},
	{Name: "Bank CIMB Niaga", Code: "CIMB", SwiftCode: "BNIAIDJA"},
}

func ensureDefaultAreasForAllTenants() error {
	return nil
}

func ensureDefaultAreasForTenant(tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil
	}

	for _, seed := range defaultAreaSeeds {
		if err := DB.Exec(`
			INSERT INTO areas (
				id, tenant_id, name, description, is_active, code, color, province, created_at, updated_at
			)
			SELECT
				gen_random_uuid(),
				?,
				?,
				?,
				true,
				?,
				?,
				?,
				NOW(),
				NOW()
			WHERE NOT EXISTS (
				SELECT 1
				FROM areas
				WHERE tenant_id = ?
				  AND lower(name) = lower(?)
				  AND deleted_at IS NULL
			)
		`, tenantID, seed.Name, seed.Description, seed.Code, seed.Color, seed.Province, tenantID, seed.Name).Error; err != nil {
			return err
		}
	}

	return nil
}

func ensureDefaultBanksForAllTenants() error {
	type tenantRow struct {
		ID string
	}

	var tenants []tenantRow
	if err := DB.Table("tenants").Select("id").Find(&tenants).Error; err != nil {
		return err
	}

	for _, tenantRow := range tenants {
		if err := ensureDefaultBanksForTenant(tenantRow.ID); err != nil {
			return err
		}
	}

	return nil
}

func ensureDefaultBanksForTenant(tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil
	}

	for _, seed := range defaultBankSeeds {
		if err := DB.Exec(`
			INSERT INTO banks (
				id, tenant_id, name, code, swift_code, is_active, created_at, updated_at
			)
			SELECT
				gen_random_uuid(),
				?,
				?,
				?,
				?,
				true,
				NOW(),
				NOW()
			WHERE NOT EXISTS (
				SELECT 1
				FROM banks
				WHERE tenant_id = ?
				  AND lower(code) = lower(?)
				  AND deleted_at IS NULL
			)
		`, tenantID, seed.Name, seed.Code, seed.SwiftCode, tenantID, seed.Code).Error; err != nil {
			return err
		}
	}

	return nil
}

// EnsureRoleNameTenantScopedUniqueIndex force-applies the tenant-scoped role-name
// uniqueness rule and removes legacy global role-name constraints.
func EnsureRoleNameTenantScopedUniqueIndex() error {
	return ensureRoleNameTenantScopedUniqueIndex()
}

func syncGeneratedTenantRolesFromPlanTemplates() error {
	var plans []tenant.SubscriptionPlanConfig
	if err := DB.Find(&plans).Error; err != nil {
		return err
	}

	planTemplates := make(map[string]tenant.RoleTemplateList, len(plans))
	for _, plan := range plans {
		if len(plan.RoleTemplates) == 0 {
			continue
		}
		planTemplates[strings.TrimSpace(strings.ToLower(plan.Slug))] = plan.RoleTemplates
	}

	if len(planTemplates) == 0 {
		return nil
	}

	type tenantRow struct {
		ID   string
		Plan string
	}

	var tenants []tenantRow
	if err := DB.Table("tenants").Select("id, plan").Find(&tenants).Error; err != nil {
		return err
	}

	for _, tenantRow := range tenants {
		templates := planTemplates[strings.TrimSpace(strings.ToLower(tenantRow.Plan))]
		for _, template := range templates {
			code := strings.TrimSpace(strings.ToLower(template.Code))
			if code == "" {
				continue
			}

			name := strings.TrimSpace(template.Name)
			if name == "" {
				continue
			}

			codePattern := code + "_%"
			if code == "tenant_owner" {
				codePattern = "tenant_owner_%"
			}

			var tenantRole role.Role
			if err := DB.Where("tenant_id = ? AND code LIKE ? AND deleted_at IS NULL", tenantRow.ID, codePattern).
				Order("created_at ASC").
				First(&tenantRole).Error; err != nil {
				continue
			}

			updates := map[string]any{
				"name":        name,
				"description": strings.TrimSpace(template.Description),
			}
			if code == "tenant_owner" {
				updates["is_protected"] = true
			}

			if err := DB.Model(&tenantRole).Updates(updates).Error; err != nil {
				log.Printf("Warning: failed to sync role %s for tenant %s: %v", code, tenantRow.ID, err)
			}
		}
	}

	return nil
}

func normalizeJournalData() error {
	log.Println("Normalizing Journal Entry reference types...")
	// Normalize to SCREAMING_SNAKE_CASE
	return DB.Exec(`
		UPDATE journal_entries 
		SET reference_type = CASE 
			WHEN lower(reference_type) IN ('goodsreceipt', 'goods_receipt', 'goods receipt') THEN 'GOODS_RECEIPT'
			WHEN lower(reference_type) IN ('supplierinvoice', 'supplier_invoice', 'supplier invoice') THEN 'SUPPLIER_INVOICE'
			WHEN lower(reference_type) IN ('salesinvoice', 'sales_invoice', 'sales invoice', 'customerinvoice', 'customer_invoice') THEN 'SALES_INVOICE'
			WHEN lower(reference_type) IN ('salespayment', 'sales_payment', 'sales payment') THEN 'SALES_PAYMENT'
			WHEN lower(reference_type) IN ('purchasepayment', 'purchase_payment', 'purchase payment') THEN 'PURCHASE_PAYMENT'
			WHEN lower(reference_type) IN ('stockopname', 'stock_opname', 'stock opname') THEN 'STOCK_OPNAME'
			WHEN lower(reference_type) IN ('assetdepreciation', 'asset_depreciation', 'asset depreciation') THEN 'ASSET_DEPRECIATION'
			ELSE reference_type
		END
		WHERE reference_type NOT IN (
			'GOODS_RECEIPT', 'SUPPLIER_INVOICE', 'SALES_INVOICE', 'SALES_PAYMENT', 'PURCHASE_PAYMENT', 'STOCK_OPNAME', 'ASSET_DEPRECIATION'
		);
	`).Error
}

func createJournalEntryPeriodLockTrigger() error {
	// The trigger will prevent inserts/updates of journal entries if their entry_date falls within a closed accounting period.
	// It is intentionally permissive for historical entries when the accounting_periods table is empty.
	if err := DB.Exec(`
		CREATE OR REPLACE FUNCTION enforce_journal_entry_period_not_closed()
		RETURNS trigger AS $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM accounting_periods
				WHERE status = 'closed'
				  AND NEW.entry_date BETWEEN start_date AND end_date
			) THEN
				RAISE EXCEPTION 'Period is closed (%), cannot modify journal entries in this period', NEW.entry_date;
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		DROP TRIGGER IF EXISTS trg_journal_entry_period_lock ON journal_entries;
		CREATE TRIGGER trg_journal_entry_period_lock
		BEFORE INSERT OR UPDATE ON journal_entries
		FOR EACH ROW EXECUTE FUNCTION enforce_journal_entry_period_not_closed();
	`).Error; err != nil {
		return err
	}
	return nil
}

// tables no longer exist the function silently returns nil.
func migrateAreaSupervisorsToEmployeeAreas() error {
	// Check if the legacy tables still exist
	var exists bool
	DB.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'area_supervisor_areas'
	)`).Scan(&exists)
	if !exists {
		log.Println("Sprint 17 migration: area_supervisor_areas table not found, skipping.")
		return nil
	}

	// Only migrate if there is data to migrate and the source references employees.
	// The legacy area_supervisors table has name/email/phone but no employee_id.
	// We try to match by email first, then by name as a fallback.
	migrationSQL := `
		INSERT INTO employee_areas (id, employee_id, area_id, is_supervisor, created_at)
		SELECT
			gen_random_uuid(),
			e.id,
			asa.area_id,
			true,
			NOW()
		FROM area_supervisor_areas asa
		JOIN area_supervisors asup ON asup.id = asa.area_supervisor_id
		JOIN employees e ON (
			(asup.email <> '' AND lower(e.email) = lower(asup.email))
			OR (asup.email = '' AND lower(e.name) = lower(asup.name))
		)
		WHERE NOT EXISTS (
			SELECT 1 FROM employee_areas ea
			WHERE ea.employee_id = e.id AND ea.area_id = asa.area_id
		)
		ON CONFLICT DO NOTHING;
	`
	result := DB.Exec(migrationSQL)
	if result.Error != nil {
		return fmt.Errorf("area supervisor migration failed: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("Sprint 17 migration: Migrated %d area supervisor records to employee_areas", result.RowsAffected)
	} else {
		log.Println("Sprint 17 migration: No new records to migrate (already migrated or no matches).")
	}

	return nil
}

func ensureGoodsReceiptWarehouseColumn() error {
	if err := DB.Exec(`
		ALTER TABLE goods_receipts
		ADD COLUMN IF NOT EXISTS warehouse_id uuid
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_goods_receipts_warehouse_id
		ON goods_receipts (warehouse_id)
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'fk_goods_receipts_warehouse'
			) THEN
				ALTER TABLE goods_receipts
				ADD CONSTRAINT fk_goods_receipts_warehouse
				FOREIGN KEY (warehouse_id)
				REFERENCES warehouses(id)
				ON UPDATE CASCADE
				ON DELETE SET NULL;
			END IF;
		END $$;
	`).Error; err != nil {
		return err
	}

	return nil
}

func ensureSalesCustomerContactColumns() error {
	if err := DB.Exec(`
		ALTER TABLE sales_quotations
		ADD COLUMN IF NOT EXISTS customer_contact_id uuid
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		ALTER TABLE sales_orders
		ADD COLUMN IF NOT EXISTS customer_contact_id uuid
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sales_quotations_customer_contact_id
		ON sales_quotations (customer_contact_id)
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sales_orders_customer_contact_id
		ON sales_orders (customer_contact_id)
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'fk_sales_quotations_customer_contact'
			) THEN
				ALTER TABLE sales_quotations
				ADD CONSTRAINT fk_sales_quotations_customer_contact
				FOREIGN KEY (customer_contact_id)
				REFERENCES crm_contacts(id)
				ON UPDATE CASCADE
				ON DELETE SET NULL;
			END IF;
		END $$;
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'fk_sales_orders_customer_contact'
			) THEN
				ALTER TABLE sales_orders
				ADD CONSTRAINT fk_sales_orders_customer_contact
				FOREIGN KEY (customer_contact_id)
				REFERENCES crm_contacts(id)
				ON UPDATE CASCADE
				ON DELETE SET NULL;
			END IF;
		END $$;
	`).Error; err != nil {
		return err
	}

	return nil
}

func ensurePOSOrderSessionColumnNullable() error {
	if err := DB.Exec(`
		ALTER TABLE pos_orders
		ALTER COLUMN session_id DROP NOT NULL
	`).Error; err != nil {
		return err
	}
	return nil
}

func ensurePOSOrderTableIDColumnType() error {
	if err := DB.Exec(`
		ALTER TABLE pos_orders
		ALTER COLUMN table_id TYPE VARCHAR(100)
		USING table_id::text
	`).Error; err != nil {
		return err
	}
	return nil
}

func migratePOSFloorPlansToOutletScope() error {
	if err := DB.Exec(`
		ALTER TABLE pos_floor_plans
		ADD COLUMN IF NOT EXISTS outlet_id uuid
	`).Error; err != nil {
		return err
	}

	type floorPlanRow struct {
		ID        string
		CompanyID *string
		OutletID  *string
		Name      string
	}

	var plans []floorPlanRow
	if err := DB.Raw(`
		SELECT id, company_id, outlet_id, name
		FROM pos_floor_plans
		WHERE deleted_at IS NULL
	`).Scan(&plans).Error; err != nil {
		return err
	}

	type outletRow struct {
		ID        string
		CompanyID *string
		Code      string
		Name      string
		IsActive  bool
	}

	var outlets []outletRow
	if err := DB.Raw(`
		SELECT id, company_id, code, name, is_active
		FROM outlets
		WHERE deleted_at IS NULL
		ORDER BY is_active DESC, name ASC, id ASC
	`).Scan(&outlets).Error; err != nil {
		return err
	}

	companyOutlets := make(map[string][]outletRow)
	for _, outlet := range outlets {
		if outlet.CompanyID == nil || *outlet.CompanyID == "" {
			continue
		}
		companyOutlets[*outlet.CompanyID] = append(companyOutlets[*outlet.CompanyID], outlet)
	}

	updated := 0
	unresolved := 0
	for _, plan := range plans {
		if plan.OutletID != nil && *plan.OutletID != "" {
			continue
		}
		if plan.CompanyID == nil || *plan.CompanyID == "" {
			unresolved++
			continue
		}

		candidates := companyOutlets[*plan.CompanyID]
		if len(candidates) == 0 {
			unresolved++
			continue
		}

		planName := strings.ToLower(strings.TrimSpace(plan.Name))
		selected := candidates[0]
		for _, candidate := range candidates {
			if !candidate.IsActive {
				continue
			}
			candidateName := strings.ToLower(strings.TrimSpace(candidate.Name))
			candidateCode := strings.ToLower(strings.TrimSpace(candidate.Code))
			if candidateName != "" && strings.Contains(planName, candidateName) {
				selected = candidate
				break
			}
			if candidateCode != "" && strings.Contains(planName, candidateCode) {
				selected = candidate
				break
			}
		}

		if err := DB.Exec(`
			UPDATE pos_floor_plans
			SET outlet_id = ?, company_id = COALESCE(company_id, ?)
			WHERE id = ?
		`, selected.ID, *plan.CompanyID, plan.ID).Error; err != nil {
			return err
		}
		updated++
	}

	if unresolved > 0 {
		log.Printf("Warning: %d floor plans could not be auto-mapped to outlet_id", unresolved)
	}

	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_floor_plans_outlet
		ON pos_floor_plans (outlet_id)
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uidx_floor_plans_outlet_floor_active
		ON pos_floor_plans (outlet_id, floor_number)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := DB.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'fk_pos_floor_plans_outlet'
			) THEN
				ALTER TABLE pos_floor_plans
				ADD CONSTRAINT fk_pos_floor_plans_outlet
				FOREIGN KEY (outlet_id)
				REFERENCES outlets(id)
				ON UPDATE CASCADE
				ON DELETE RESTRICT;
			END IF;
		END $$;
	`).Error; err != nil {
		return err
	}

	if unresolved == 0 {
		if err := DB.Exec(`
			ALTER TABLE pos_floor_plans
			ALTER COLUMN outlet_id SET NOT NULL
		`).Error; err != nil {
			return err
		}
	}

	if updated > 0 {
		log.Printf("POS floor plan outlet migration completed: %d records updated", updated)
	}

	return nil
}

// createSearchIndexes creates GIN indexes for optimized text search
func createSearchIndexes() error {
	// Ensure pg_trgm extension exists (required for GIN trgm_ops)
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		return fmt.Errorf("failed to create pg_trgm extension: %w", err)
	}

	indexes := []string{
		// User module indexes
		"CREATE INDEX IF NOT EXISTS idx_users_name_gin ON users USING gin (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_users_email_gin ON users USING gin (email gin_trgm_ops)",

		// HRD leave request search indexes (added for leave request search feature)
		"CREATE INDEX IF NOT EXISTS idx_employees_name_gin ON employees USING gin (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_leave_types_name_gin ON leave_types USING gin (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_leave_requests_reason_gin ON leave_requests USING gin (reason gin_trgm_ops)",

		// HRD employee certification search indexes (Sprint 14)
		"CREATE INDEX IF NOT EXISTS idx_employee_certifications_name_gin ON employee_certifications USING gin (certificate_name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_employee_certifications_issued_by_gin ON employee_certifications USING gin (issued_by gin_trgm_ops)",

		// HRD employee asset search indexes (Sprint 14)
		"CREATE INDEX IF NOT EXISTS idx_employee_assets_name_gin ON employee_assets USING gin (asset_name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_employee_assets_code_gin ON employee_assets USING gin (asset_code gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_employee_assets_category_gin ON employee_assets USING gin (asset_category gin_trgm_ops)",

		// HRD evaluation search indexes (Sprint 15)
		"CREATE INDEX IF NOT EXISTS idx_evaluation_groups_name_gin ON evaluation_groups USING gin (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_evaluation_criteria_name_gin ON evaluation_criteria USING gin (name gin_trgm_ops)",
		// HRD recruitment search indexes (Sprint 15)
		"CREATE INDEX IF NOT EXISTS idx_recruitment_requests_code_gin ON recruitment_requests USING gin (request_code gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_recruitment_requests_desc_gin ON recruitment_requests USING gin (job_description gin_trgm_ops)",
		// HRD recruitment applicant search indexes
		"CREATE INDEX IF NOT EXISTS idx_recruitment_applicants_name_gin ON recruitment_applicants USING gin (full_name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_recruitment_applicants_email_gin ON recruitment_applicants USING gin (email gin_trgm_ops)",
	}

	for _, idx := range indexes {
		if err := DB.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}

// shouldDropTables checks if we should drop all tables (development only)
// This function ensures that tables are NEVER dropped in production mode
func shouldDropTables() bool {
	// CRITICAL: Never drop tables in production
	// Check both config and environment variable for safety
	env := ""
	if config.AppConfig != nil {
		env = config.AppConfig.Server.Env
	}

	// Fallback to environment variable if config is not loaded yet
	if env == "" {
		env = os.Getenv("ENV")
	}

	// Safety check: Never allow in production
	if env == "production" || env == "prod" {
		log.Println("🔒 Production mode detected: Table drop is disabled (safety protection)")
		return false
	}

	// Only allow dropping tables in development mode
	// Check environment variable DROP_TABLES or DROP_ALL_TABLES (from package.json scripts)
	dropTables := os.Getenv("DROP_TABLES")
	if dropTables == "" {
		dropTables = os.Getenv("DROP_ALL_TABLES")
	}

	if dropTables == "true" || dropTables == "1" {
		// Double check: ensure we're not in production
		if env == "" || env == "development" || env == "dev" {
			log.Println("🔧 Development mode: Table drop is enabled")
			return true
		}
		log.Printf("⚠️  Warning: DROP_TABLES is set but ENV=%s is not development. Skipping table drop.", env)
		return false
	}
	return false
}

// DropAllTables drops all tables in the database (development only)
// This function has built-in safety checks to prevent accidental data loss
func DropAllTables() error {
	// Safety check: Never allow dropping tables in production
	// Check both config and environment variable for maximum safety
	env := ""
	if config.AppConfig != nil {
		env = config.AppConfig.Server.Env
	}

	// Fallback to environment variable if config is not loaded yet
	if env == "" {
		env = os.Getenv("ENV")
	}

	if env == "production" || env == "prod" {
		return fmt.Errorf("🔒 CRITICAL: Cannot drop tables in production mode (ENV=%s). This is a safety protection", env)
	}

	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Get all table names in the current schema
	var tables []string
	err := DB.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = CURRENT_SCHEMA()
		AND tablename NOT LIKE 'pg_%'
		AND tablename NOT LIKE '_prisma_%'
	`).Scan(&tables).Error

	if err != nil {
		return fmt.Errorf("failed to get table list: %w", err)
	}

	if len(tables) == 0 {
		log.Println("No tables to drop")
		return nil
	}

	// Disable foreign key checks temporarily and drop all tables
	// PostgreSQL doesn't have a simple way to disable FK checks, so we use CASCADE
	log.Printf("⚠️  DEVELOPMENT MODE: Dropping %d tables...", len(tables))

	for _, table := range tables {
		// Use CASCADE to drop dependent objects
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if err := DB.Exec(dropSQL).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
			// Continue with other tables
		} else {
			log.Printf("Dropped table: %s", table)
		}
	}

	// Also drop any remaining sequences
	var sequences []string
	_ = DB.Raw(`
		SELECT sequence_name 
		FROM information_schema.sequences 
		WHERE sequence_schema = CURRENT_SCHEMA()
	`).Scan(&sequences).Error

	for _, seq := range sequences {
		dropSQL := fmt.Sprintf("DROP SEQUENCE IF EXISTS %s CASCADE", seq)
		_ = DB.Exec(dropSQL).Error // Ignore errors
	}

	log.Println("✅ All tables and sequences dropped successfully")
	return nil
}

// migrateWithErrorHandling migrates models while handling common constraint errors
func migrateWithErrorHandling(models ...interface{}) error {
	for _, model := range models {
		err := DB.AutoMigrate(model)
		if err != nil {
			// Check if error is PostgreSQL error code 42704 (undefined_object)
			// This happens when trying to DROP a constraint that doesn't exist
			errStr := err.Error()
			if strings.Contains(errStr, "SQLSTATE 42704") ||
				(strings.Contains(errStr, "does not exist") && strings.Contains(errStr, "constraint")) {
				log.Printf("Warning: Constraint error during migration (safe to ignore): %v", err)
				log.Println("GORM will create the necessary constraints. This is expected during schema evolution.")
				// Continue with next model - GORM might have partially succeeded
				continue
			}
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	return nil
}

// migrateEmployeeContractData migrates contract data from employees table to employee_contracts table
// and removes contract fields from employees table
func migrateEmployeeContractData() error {
	// Check if employees table has contract columns
	var hasContractStatus bool
	err := DB.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'employees' AND column_name = 'contract_status'
		)
	`).Scan(&hasContractStatus).Error
	if err != nil {
		return fmt.Errorf("check contract_status column: %w", err)
	}

	if !hasContractStatus {
		log.Println("Employee contract data migration: contract fields already removed from employees table")
		return nil
	}

	log.Println("Migrating contract data from employees table to employee_contracts table...")

	// Migrate existing contract data to employee_contracts table
	// Only migrate employees that have contract_status set
	result := DB.Exec(`
		INSERT INTO employee_contracts (
			id, employee_id, contract_number, contract_type, start_date, end_date, 
			document_path, status, created_at, updated_at
		)
		SELECT 
			gen_random_uuid(),
			id,
			'EMP-' || employee_code || '-INITIAL',
			CASE 
				WHEN contract_status = 'permanent' THEN 'PKWTT'
				WHEN contract_status = 'contract' THEN 'PKWT'
				WHEN contract_status = 'intern' THEN 'Intern'
				ELSE 'PKWTT'
			END,
			contract_start_date,
			contract_end_date,
			NULL,
			'ACTIVE',
			NOW(),
			NOW()
		FROM employees
		WHERE contract_status IS NOT NULL 
			AND contract_status != ''
			AND NOT EXISTS (
				SELECT 1 FROM employee_contracts ec 
				WHERE ec.employee_id = employees.id
			)
	`)

	if result.Error != nil {
		return fmt.Errorf("migrate contract data: %w", result.Error)
	}

	log.Printf("Migrated %d employee contracts", result.RowsAffected)

	// Drop contract columns from employees table
	columnsToDrop := []string{
		"contract_status",
		"contract_start_date",
		"contract_end_date",
	}

	for _, col := range columnsToDrop {
		var hasCol bool
		err := DB.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'employees' AND column_name = ?
			)
		`, col).Scan(&hasCol).Error
		if err != nil {
			return fmt.Errorf("check column %s: %w", col, err)
		}

		if hasCol {
			dropSQL := fmt.Sprintf("ALTER TABLE employees DROP COLUMN IF EXISTS %s", col)
			if err := DB.Exec(dropSQL).Error; err != nil {
				return fmt.Errorf("drop column %s: %w", col, err)
			}
			log.Printf("Dropped column employees.%s", col)
		}
	}

	log.Println("Employee contract data migration completed successfully")
	return nil
}

// handleConstraintIssues attempts to fix common constraint issues before migration
func handleConstraintIssues() error {
	// Check if roles table exists
	var exists bool
	err := DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'roles')").Scan(&exists).Error
	if err != nil || !exists {
		return nil // Table doesn't exist, nothing to fix
	}

	// Get all unique constraints on the roles table
	type ConstraintInfo struct {
		ConstraintName string
	}
	var constraints []ConstraintInfo
	err = DB.Raw(`
		SELECT conname as constraint_name
		FROM pg_constraint
		WHERE conrelid = 'roles'::regclass
		AND contype = 'u'
	`).Scan(&constraints).Error

	if err != nil {
		// If we can't query constraints, that's okay - continue anyway
		return nil
	}

	// Drop all unique constraints on code column (GORM will recreate them)
	for _, constraint := range constraints {
		// Check if this constraint is on the 'code' column
		var columnName string
		err = DB.Raw(`
			SELECT a.attname
			FROM pg_constraint c
			JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
			WHERE c.conname = ?
			AND a.attname = 'code'
			LIMIT 1
		`, constraint.ConstraintName).Scan(&columnName).Error

		if err == nil && columnName == "code" {
			// Drop the constraint
			dropSQL := fmt.Sprintf("ALTER TABLE roles DROP CONSTRAINT IF EXISTS %s", constraint.ConstraintName)
			_ = DB.Exec(dropSQL).Error // Ignore errors - constraint might not exist
		}
	}

	return nil
}

// migrateTimezoneData creates timezone tables and inserts Indonesia timezone data
func migrateTimezoneData() error {
	log.Println("Migrating timezone data...")

	// Create timezone tables if not exist
	err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS countries (
			country_code CHAR(2) PRIMARY KEY,
			country_name VARCHAR(45)
		);

		CREATE TABLE IF NOT EXISTS time_zones (
			id SERIAL PRIMARY KEY,
			zone_name VARCHAR(35) NOT NULL,
			country_code CHAR(2) REFERENCES countries(country_code),
			abbreviation VARCHAR(6) NOT NULL,
			time_start BIGINT NOT NULL,
			gmt_offset INT NOT NULL,
			dst CHAR(1) NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_time_zones_zone_name ON time_zones(zone_name);
		CREATE INDEX IF NOT EXISTS idx_time_zones_country_code ON time_zones(country_code);
		CREATE INDEX IF NOT EXISTS idx_time_zones_time_start ON time_zones(time_start);
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create timezone tables: %w", err)
	}

	// Insert Indonesia country
	err = DB.Exec(`
		INSERT INTO countries (country_code, country_name) VALUES (ID, Indonesia)
		ON CONFLICT (country_code) DO NOTHING;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to insert Indonesia country: %w", err)
	}

	// Insert Indonesia timezones (WIB, WITA, WIT)
	err = DB.Exec(`
		INSERT INTO time_zones (zone_name, country_code, abbreviation, time_start, gmt_offset, dst) VALUES
		(Asia/Jakarta, ID, WIB, 0, 25200, 0),
		(Asia/Makassar, ID, WITA, 0, 28800, 0),
		(Asia/Jayapura, ID, WIT, 0, 32400, 0)
		ON CONFLICT DO NOTHING;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to insert Indonesia timezones: %w", err)
	}

	log.Println("Timezone data migration completed")
	return nil
}

// backfillTenantIDs assigns the default tenant ID to all existing rows whose
// tenant_id is NULL. This is a one-time idempotent migration that runs on
// every server start (the UPDATE is a no-op when there are no NULL rows).
// Tables that are platform-wide (geographic data, roles, permissions, currencies)
// are intentionally excluded.
func backfillTenantIDs() error {
	const defaultTenantID = "a0000001-0000-0000-0000-000000000001"

	// All business tables that now carry tenant_id.
	// Sorted by dependency order (parent before child) so FK constraints are satisfied.
	tables := []string{
		"companies",
		"divisions",
		"job_positions",
		"business_units",
		"business_types",
		"areas",
		"outlets",
		"warehouses",
		"employees",
		"employee_areas",
		"employee_assets",
		"employee_certifications",
		"employee_contracts",
		"employee_education_histories",
		"employee_signatures",
		"customer_types",
		"customers",
		"customer_banks",
		"supplier_types",
		"banks",
		"suppliers",
		"supplier_banks",
		"supplier_contacts",
		"product_categories",
		"product_brands",
		"product_segments",
		"product_types",
		"units_of_measures",
		"packagings",
		"procurement_types",
		"products",
		"product_recipe_items",
		"product_recipe_versions",
		"product_recipe_version_items",
		"payment_terms",
		"courier_agencies",
		"so_sources",
		"leave_types",
		"bank_accounts",
		"chart_of_accounts",
		"finance_settings",
		"journal_entries",
		"journal_lines",
		"journal_reversals",
		"journal_attachments",
		"adjustment_journal_approvals",
		"journal_templates",
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
		"financial_closings",
		"fiscal_years",
		"accounting_periods",
		"financial_closing_snapshots",
		"financial_closing_logs",
		"tax_configurations",
		"inventory_settings",
		"opening_balance_lines",
		"tax_invoices",
		"non_trade_payables",
		"salary_structures",
		"valuation_runs",
		"valuation_run_details",
		"up_country_costs",
		"system_account_mappings",
		"travel_plans",
		"travel_plan_days",
		"travel_plan_stops",
		"travel_plan_day_notes",
		"travel_plan_expenses",
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
		"yearly_targets",
		"monthly_targets",
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
		"inventory_batches",
		"stock_movements",
		"stock_ledgers",
		"stock_opnames",
		"stock_opname_items",
		"pipeline_stages",
		"lead_sources",
		"lead_statuses",
		"contact_roles",
		"activity_types",
		"crm_contacts",
		"crm_leads",
		"crm_deals",
		"crm_activities",
		"crm_tasks",
		"crm_reminders",
		"crm_schedules",
		"crm_area_captures",
		"crm_visit_reports",
		"work_schedules",
		"holidays",
		"attendance_records",
		"overtime_requests",
		"leave_requests",
		"evaluation_groups",
		"evaluation_criteria",
		"employee_evaluations",
		"recruitment_requests",
		"recruitment_applicants",
		"pos_sessions",
		"pos_orders",
		"pos_order_items",
		"pos_payments",
		"pos_configs",
		"pos_floor_plans",
		"layout_versions",
		"pos_table_status_records",
		"feedback_forms",
		"feedback_tokens",
		"feedback_responses",
		"loyalty_programs",
		"loyalty_members",
		"loyalty_point_ledgers",
		"notifications",
		"dashboard_layouts",
		"ai_chat_sessions",
		"ai_chat_messages",
		"ai_action_logs",
	}

	for _, table := range tables {
		// Check if the table exists AND has a tenant_id column before attempting the update.
		var hasColumn bool
		checkSQL := `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = $1
				  AND column_name = 'tenant_id'
			)`
		if err := DB.Raw(checkSQL, table).Scan(&hasColumn).Error; err != nil {
			log.Printf("backfillTenantIDs: could not check column on %s: %v", table, err)
			continue
		}
		if !hasColumn {
			continue
		}

		result := DB.Exec(
			"UPDATE "+table+" SET tenant_id = ? WHERE tenant_id IS NULL OR tenant_id = ''",
			defaultTenantID,
		)
		if result.Error != nil {
			log.Printf("backfillTenantIDs: failed to backfill %s: %v", table, result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("backfillTenantIDs: backfilled %d rows in %s", result.RowsAffected, table)
		}
	}

	if DB.Migrator().HasTable("stock_opnames") {
		if err := DB.Exec("DROP INDEX IF EXISTS idx_stock_opnames_opname_number").Error; err != nil {
			log.Printf("backfillTenantIDs: could not drop legacy stock opname index: %v", err)
		}
		if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_opnames_tenant_number ON stock_opnames (tenant_id, opname_number)").Error; err != nil {
			log.Printf("backfillTenantIDs: could not create tenant stock opname index: %v", err)
		}
	}

	return nil
}
