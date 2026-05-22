package security

import (
	"context"
	"log"

	"gorm.io/gorm"
)

// ScopeQueryOptions configures which database columns to use for scope resolution.
// Different modules use different fields (e.g., sales uses sales_rep_id, purchase uses created_by).
type ScopeQueryOptions struct {
	// OwnerUserIDColumn is the column that stores the owner's user ID (default: "created_by")
	OwnerUserIDColumn string
	// OwnerEmployeeIDColumn is the column that stores the owner's employee ID (e.g., "sales_rep_id")
	OwnerEmployeeIDColumn string
	// DivisionJoinSQL is a subquery that resolves division membership.
	// Example: "sales_rep_id IN (SELECT id FROM employees WHERE division_id = ?)"
	DivisionJoinSQL string
	// AreaIDColumn is the column that stores the area ID (e.g., "delivery_area_id")
	AreaIDColumn string
	// AreaJoinSQL is an optional custom subquery for area scoping when the table
	// does not have a direct area_id column.
	AreaJoinSQL string
	// OutletIDColumn is the column that stores the outlet ID (e.g., "outlet_id").
	// Used when scope is OUTLET to filter records by employee's assigned outlets.
	OutletIDColumn string
	// WarehouseIDColumn is the column that stores the warehouse ID (e.g., "warehouse_id").
	// Used when scope is WAREHOUSE to filter records by employee's assigned warehouses.
	WarehouseIDColumn string
	// OutletJoinSQL is an optional custom subquery for outlet scoping when the table
	// doesn't have a direct outlet_id column. Example:
	// "created_by IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ?) AND user_id IS NOT NULL)"
	OutletJoinSQL string
	// WarehouseJoinSQL is an optional custom subquery for warehouse scoping.
	WarehouseJoinSQL string
}

// DefaultScopeQueryOptions returns options suitable for most modules
func DefaultScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
	}
}

// SalesScopeQueryOptions returns options for sales module
func SalesScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn:     "created_by",
		OwnerEmployeeIDColumn: "sales_rep_id",
		DivisionJoinSQL:       "sales_rep_id IN (SELECT id FROM employees WHERE division_id = ?)",
		AreaIDColumn:          "delivery_area_id",
	}
}

// PurchaseScopeQueryOptions returns options for purchase module
func PurchaseScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn:     "created_by",
		OwnerEmployeeIDColumn: "employee_id",
	}
}

// HRDScopeQueryOptions returns options for HRD module
func HRDScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerEmployeeIDColumn: "employee_id",
		DivisionJoinSQL:       "employee_id IN (SELECT id FROM employees WHERE division_id = ?)",
	}
}

// SalesTargetScopeQueryOptions returns options for sales target module
// where the record owner is identified by employee_id (the sales rep)
func SalesTargetScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerEmployeeIDColumn: "employee_id",
		DivisionJoinSQL:       "employee_id IN (SELECT id FROM employees WHERE division_id = ?)",
	}
}

// ScopeFilter holds the resolved scope context for query filtering
type ScopeFilter struct {
	Scope        string // OWN, DIVISION, AREA, OUTLET, WAREHOUSE, ALL
	UserID       string
	EmployeeID   string
	DivisionID   string
	AreaIDs      []string
	OutletIDs    []string
	WarehouseIDs []string
}

// ApplyToQuery applies scope-based WHERE conditions to a GORM query.
// Returns the same db instance with appropriate filters applied.
// If scope is ALL or empty, no additional filters are added.
func (sf *ScopeFilter) ApplyToQuery(db *gorm.DB, opts ScopeQueryOptions) *gorm.DB {
	switch sf.Scope {
	case "OWN":
		return sf.applyOwn(db, opts)
	case "DIVISION":
		return sf.applyDivision(db, opts)
	case "AREA":
		return sf.applyArea(db, opts)
	case "OUTLET":
		return sf.applyOutlet(db, opts)
	case "WAREHOUSE":
		return sf.applyWarehouse(db, opts)
	default: // ALL or empty
		return db
	}
}

func (sf *ScopeFilter) applyOwn(db *gorm.DB, opts ScopeQueryOptions) *gorm.DB {
	// Build OR conditions: owner by user_id OR owner by employee_id
	conditions := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)

	if columnExists(db, opts.OwnerUserIDColumn) && sf.UserID != "" {
		conditions = append(conditions, opts.OwnerUserIDColumn+" = ?")
		args = append(args, sf.UserID)
	}
	// Prefer exact employee match when available
	if columnExists(db, opts.OwnerEmployeeIDColumn) && sf.EmployeeID != "" {
		conditions = append(conditions, opts.OwnerEmployeeIDColumn+" = ?")
		args = append(args, sf.EmployeeID)
	}

	// If no employee id in context but we have a user id, allow matching
	// records where the employee column points to an employee whose user_id
	// equals the current user. This covers users who don't have a resolved
	// employee in middleware but records still reference employees linked to
	// that user's account.
	if columnExists(db, opts.OwnerEmployeeIDColumn) && sf.EmployeeID == "" && sf.UserID != "" {
		// Cover two cases:
		// 1) some records may have employee_id set to the user's UUID (legacy/mistmatch)
		// 2) normal case where employee_id refers to an employee row linked to user_id
		conditions = append(conditions, opts.OwnerEmployeeIDColumn+" = ?")
		args = append(args, sf.UserID)
		conditions = append(conditions, opts.OwnerEmployeeIDColumn+" IN (SELECT id FROM employees WHERE user_id = ? AND deleted_at IS NULL)")
		args = append(args, sf.UserID)
	}

	if len(conditions) == 0 {
		// No owner columns configured — restrict to nothing for safety
		return db.Where("1 = 0")
	}

	// Join conditions with OR
	combined := conditions[0]
	for i := 1; i < len(conditions); i++ {
		combined += " OR " + conditions[i]
	}
	return db.Where("("+combined+")", args...)
}

func (sf *ScopeFilter) applyDivision(db *gorm.DB, opts ScopeQueryOptions) *gorm.DB {
	if sf.DivisionID == "" {
		// User has no division — fall back to OWN
		return sf.applyOwn(db, opts)
	}

	if opts.DivisionJoinSQL != "" {
		return db.Where(opts.DivisionJoinSQL, sf.DivisionID)
	}

	// Column-aware fallback for modules with heterogeneous ownership columns.
	conditions := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)

	if columnExists(db, opts.OwnerUserIDColumn) {
		conditions = append(conditions, opts.OwnerUserIDColumn+" IN (SELECT user_id FROM employees WHERE division_id = ? AND user_id IS NOT NULL)")
		args = append(args, sf.DivisionID)
	}

	if columnExists(db, opts.OwnerEmployeeIDColumn) {
		conditions = append(conditions, opts.OwnerEmployeeIDColumn+" IN (SELECT id FROM employees WHERE division_id = ?)")
		args = append(args, sf.DivisionID)
	}

	if len(conditions) > 0 {
		combined := conditions[0]
		for i := 1; i < len(conditions); i++ {
			combined += " OR " + conditions[i]
		}
		return db.Where("("+combined+")", args...)
	}

	return sf.applyOwn(db, opts)
}

func columnExists(db *gorm.DB, columnName string) bool {
	if columnName == "" {
		return false
	}

	if db == nil || db.Statement == nil || db.Statement.Model == nil {
		// For queries without a concrete model, keep previous permissive behavior.
		return true
	}

	return db.Migrator().HasColumn(db.Statement.Model, columnName)
}

func (sf *ScopeFilter) applyArea(db *gorm.DB, opts ScopeQueryOptions) *gorm.DB {
	if len(sf.AreaIDs) == 0 {
		// User has no areas — fall back to OWN
		return sf.applyOwn(db, opts)
	}

	if opts.AreaJoinSQL != "" {
		return db.Where(opts.AreaJoinSQL, sf.AreaIDs)
	}

	if opts.AreaIDColumn != "" {
		return db.Where(opts.AreaIDColumn+" IN ?", sf.AreaIDs)
	}

	// No area column configured — fall back to division
	return sf.applyDivision(db, opts)
}

// applyOutlet filters records by the user's assigned outlet IDs.
// Falls back to custom OutletJoinSQL when no direct outlet_id column exists,
// then to OWN scope if no outlet assignments are found.
func (sf *ScopeFilter) applyOutlet(db *gorm.DB, opts ScopeQueryOptions) *gorm.DB {
	if len(sf.OutletIDs) == 0 {
		// User has no outlets — fall back to OWN
		return sf.applyOwn(db, opts)
	}

	// Direct column filter (most common: tables with outlet_id)
	if opts.OutletIDColumn != "" {
		return db.Where(opts.OutletIDColumn+" IN ?", sf.OutletIDs)
	}

	// Custom join SQL for tables without direct outlet_id
	if opts.OutletJoinSQL != "" {
		return db.Where(opts.OutletJoinSQL, sf.OutletIDs)
	}

	// Fallback: filter by owner employees assigned to the same outlets.
	// This works for tables with created_by or employee_id columns.
	conditions := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)

	if columnExists(db, opts.OwnerUserIDColumn) {
		conditions = append(conditions,
			opts.OwnerUserIDColumn+" IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL) AND user_id IS NOT NULL AND deleted_at IS NULL)")
		args = append(args, sf.OutletIDs)
	}

	if columnExists(db, opts.OwnerEmployeeIDColumn) {
		conditions = append(conditions,
			opts.OwnerEmployeeIDColumn+" IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL)")
		args = append(args, sf.OutletIDs)
	}

	if len(conditions) > 0 {
		combined := conditions[0]
		for i := 1; i < len(conditions); i++ {
			combined += " OR " + conditions[i]
		}
		return db.Where("("+combined+")", args...)
	}

	// No outlet column or join configured — fall back to OWN
	return sf.applyOwn(db, opts)
}

// applyWarehouse filters records by the user's assigned warehouse IDs.
// Falls back to custom WarehouseJoinSQL when no direct warehouse_id column exists,
// then to OUTLET scope (which cascades to OWN if no outlets).
func (sf *ScopeFilter) applyWarehouse(db *gorm.DB, opts ScopeQueryOptions) *gorm.DB {
	if len(sf.WarehouseIDs) == 0 {
		// User has no warehouses — fall back to OUTLET, then OWN
		return sf.applyOutlet(db, opts)
	}

	// Direct column filter (most common: tables with warehouse_id)
	if opts.WarehouseIDColumn != "" {
		return db.Where(opts.WarehouseIDColumn+" IN ?", sf.WarehouseIDs)
	}

	// Custom join SQL for tables without direct warehouse_id
	if opts.WarehouseJoinSQL != "" {
		return db.Where(opts.WarehouseJoinSQL, sf.WarehouseIDs)
	}

	// No warehouse column — fall back to outlet scope
	return sf.applyOutlet(db, opts)
}

// NewScopeFilterFromContext creates a ScopeFilter from middleware-injected context values.
// Pass the permission_scope and scope_context from Gin context.
func NewScopeFilterFromContext(scope string, userID, employeeID, divisionID string, areaIDs, outletIDs, warehouseIDs []string) *ScopeFilter {
	return &ScopeFilter{
		Scope:        scope,
		UserID:       userID,
		EmployeeID:   employeeID,
		DivisionID:   divisionID,
		AreaIDs:      areaIDs,
		OutletIDs:    outletIDs,
		WarehouseIDs: warehouseIDs,
	}
}

// ApplyScopeFilter reads scope values from the request context (set by ScopeMiddleware
// and RequirePermission middleware) and applies WHERE conditions to the GORM query.
// This is the main entry point for repositories to enforce data-level authorization.
//
// Supported scopes: ALL, OWN, DIVISION, AREA, OUTLET, WAREHOUSE
// The scope is read from "permission_scope" context key (set by RequirePermission middleware).
// Employee assignments (outlets, warehouses, areas, division) are read from ScopeMiddleware context keys.
func ApplyScopeFilter(db *gorm.DB, ctx context.Context, opts ScopeQueryOptions) *gorm.DB {
	scope, _ := ctx.Value("permission_scope").(string)
	if scope == "" || scope == "ALL" {
		log.Printf("[security][scope] skipping filter scope=%s", scope)
		return db
	}

	userID, _ := ctx.Value("scope_user_id").(string)
	employeeID, _ := ctx.Value("scope_employee_id").(string)
	divisionID, _ := ctx.Value("scope_division_id").(string)
	areaIDs, _ := ctx.Value("scope_area_ids").([]string)
	outletIDs, _ := ctx.Value("scope_outlet_ids").([]string)
	warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string)

	log.Printf("[security][scope] applying filter scope=%s user_id=%s outlets=%d warehouses=%d", scope, userID, len(outletIDs), len(warehouseIDs))

	filter := &ScopeFilter{
		Scope:        scope,
		UserID:       userID,
		EmployeeID:   employeeID,
		DivisionID:   divisionID,
		AreaIDs:      areaIDs,
		OutletIDs:    outletIDs,
		WarehouseIDs: warehouseIDs,
	}
	return filter.ApplyToQuery(db, opts)
}

// StockMovementScopeQueryOptions returns options for stock/inventory module
func StockMovementScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
		DivisionJoinSQL:   "created_by IN (SELECT user_id FROM employees WHERE division_id = ? AND user_id IS NOT NULL)",
		WarehouseIDColumn: "warehouse_id",
	}
}

// FinanceScopeQueryOptions returns options for finance module
func FinanceScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
		DivisionJoinSQL:   "created_by IN (SELECT user_id FROM employees WHERE division_id = ? AND user_id IS NOT NULL)",
		AreaJoinSQL:       "created_by IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_areas WHERE area_id IN ?) AND user_id IS NOT NULL AND deleted_at IS NULL)",
		OutletJoinSQL:     "created_by IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL) AND user_id IS NOT NULL AND deleted_at IS NULL)",
	}
}

// SalesPaymentScopeQueryOptions returns options for sales payment module
func SalesPaymentScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
		DivisionJoinSQL:   "created_by IN (SELECT user_id FROM employees WHERE division_id = ? AND user_id IS NOT NULL)",
		OutletJoinSQL:     "created_by IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL) AND user_id IS NOT NULL AND deleted_at IS NULL)",
	}
}

// POSScopeQueryOptions returns options for POS module (orders, floor plans, etc.)
func POSScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
		OutletIDColumn:    "outlet_id",
	}
}

// FeedbackScopeQueryOptions returns options for feedback/loyalty module
func FeedbackScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
		OutletIDColumn:    "outlet_id",
	}
}

// InventoryWarehouseScopeQueryOptions returns options for inventory with warehouse_id column
func InventoryWarehouseScopeQueryOptions() ScopeQueryOptions {
	return ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
		WarehouseIDColumn: "warehouse_id",
	}
}

// MixedOwnershipScopeQueryOptions returns options for records owned by either a user or an employee/assignee.
// The employee column is used for division scoping and employee-based OWN access.
func MixedOwnershipScopeQueryOptions(ownerEmployeeIDColumn string) ScopeQueryOptions {
	opts := ScopeQueryOptions{
		OwnerUserIDColumn: "created_by",
	}

	if ownerEmployeeIDColumn != "" {
		opts.OwnerEmployeeIDColumn = ownerEmployeeIDColumn
		opts.DivisionJoinSQL = ownerEmployeeIDColumn + " IN (SELECT id FROM employees WHERE division_id = ?)"
	}

	return opts
}
