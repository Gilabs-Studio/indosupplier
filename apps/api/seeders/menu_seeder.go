package seeders

import (
	"log"
	"strings"
	"unicode"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	permission "github.com/gilabs/gims/api/internal/permission/data/models"
)

func sanitizeSlugToken(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "menu"
	}

	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}

	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "menu"
	}
	return result
}

func inferMenuModule(menu *permission.Menu) string {
	if strings.HasPrefix(menu.URL, "/finance") || strings.EqualFold(menu.Name, "Finance") || strings.EqualFold(menu.Name, "Finance & Accounting") {
		return "finance"
	}
	if strings.HasPrefix(menu.URL, "/sales") || strings.EqualFold(menu.Name, "Sales") {
		return "sales"
	}
	if strings.HasPrefix(menu.URL, "/purchase") || strings.EqualFold(menu.Name, "Purchase") {
		return "purchase"
	}
	if strings.HasPrefix(menu.URL, "/stock") || strings.EqualFold(menu.Name, "Stock") {
		return "stock"
	}
	if strings.HasPrefix(menu.URL, "/hrd") || strings.EqualFold(menu.Name, "HRD") {
		return "hrd"
	}
	if strings.HasPrefix(menu.URL, "/crm") || strings.EqualFold(menu.Name, "CRM") {
		return "crm"
	}
	if strings.HasPrefix(menu.URL, "/master-data") || strings.EqualFold(menu.Name, "Master Data") {
		return "master-data"
	}

	if menu.ParentID != nil {
		var parent permission.Menu
		if err := database.DB.Where("id = ?", *menu.ParentID).First(&parent).Error; err == nil {
			if strings.TrimSpace(parent.Module) != "" {
				return parent.Module
			}
		}
	}

	return "core"
}

func deriveMenuSlug(menu *permission.Menu) string {
	if menu.URL != "" && menu.URL != "#" {
		trimmed := strings.Trim(strings.TrimSpace(menu.URL), "/")
		trimmed = strings.ReplaceAll(trimmed, "/", ".")
		if trimmed != "" {
			return trimmed
		}
	}

	base := sanitizeSlugToken(menu.Name)
	module := strings.TrimSpace(menu.Module)
	if module != "" {
		if menu.ParentID != nil && len(*menu.ParentID) >= 8 {
			return module + "." + base + "." + strings.ToLower((*menu.ParentID)[:8])
		}
		return module + "." + base
	}
	return base
}

func enrichMenuMetadata(menu *permission.Menu) {
	if strings.TrimSpace(menu.Status) == "" {
		menu.Status = "active"
	}

	menu.IsActive = strings.EqualFold(menu.Status, "active")
	menu.IsClickable = !(strings.TrimSpace(menu.URL) == "" || strings.TrimSpace(menu.URL) == "#")

	if strings.TrimSpace(menu.Module) == "" {
		menu.Module = inferMenuModule(menu)
	}
	if strings.TrimSpace(menu.Slug) == "" {
		menu.Slug = deriveMenuSlug(menu)
	}
}

// createMenu is a helper function to create or update a menu
func createMenu(menu *permission.Menu) error {
	enrichMenuMetadata(menu)

	var existing permission.Menu
	query := database.DB.Where("name = ?", menu.Name)
	if menu.ParentID != nil {
		query = query.Where("parent_id = ?", *menu.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.First(&existing).Error; err == nil {
		// Found existing menu, update it
		menu.ID = existing.ID // Important: keep the ID for children
		if err := database.DB.Model(&existing).Updates(map[string]interface{}{
			"icon":         menu.Icon,
			"url":          menu.URL,
			"order":        menu.Order,
			"status":       menu.Status,
			"module":       menu.Module,
			"slug":         menu.Slug,
			"access":       menu.Access,
			"is_active":    menu.IsActive,
			"is_clickable": menu.IsClickable,
		}).Error; err != nil {
			return err
		}
		log.Printf("Updated menu: %s", menu.Name)
		return nil
	}

	// Create new
	if err := database.DB.Create(menu).Error; err != nil {
		return err
	}
	log.Printf("Created menu: %s", menu.Name)
	return nil
}

// createChildMenu creates a child menu under a parent
func createChildMenu(name, icon, url string, parentID *string, order int) (*permission.Menu, error) {
	access := true
	if parentID != nil {
		var parent permission.Menu
		if err := database.DB.Where("id = ?", *parentID).First(&parent).Error; err == nil {
			access = parent.Access
		}
	}

	menu := &permission.Menu{
		Name:     name,
		Icon:     icon,
		URL:      url,
		ParentID: parentID,
		Order:    order,
		Status:   "active",
		Access:   access,
	}
	if err := createMenu(menu); err != nil {
		return nil, err
	}
	return menu, nil
}

// SeedMenus seeds initial ERP menus based on database structure
func SeedMenus() error {

	log.Println("Seeding ERP menu structure...")

	// ============================================================
	// ROOT LEVEL MENUS
	// ============================================================

	// 1. Dashboard
	dashboardMenu := &permission.Menu{
		Name:   "Dashboard",
		Icon:   "layout-dashboard",
		URL:    "/dashboard",
		Order:  1,
		Status: "active",
		Access: true,
	}
	if err := createMenu(dashboardMenu); err != nil {
		return err
	}

	// 2. Master Data
	masterDataMenu := &permission.Menu{
		Name:   "Master Data",
		Icon:   "database",
		URL:    "/master-data",
		Order:  9,
		Status: "active",
		Access: true,
	}
	if err := createMenu(masterDataMenu); err != nil {
		return err
	}

	// 3. Sales
	salesMenu := &permission.Menu{
		Name:   "Sales",
		Icon:   "shopping-cart",
		URL:    "/sales",
		Order:  3,
		Status: "active",
		Access: false,
	}
	if err := createMenu(salesMenu); err != nil {
		return err
	}

	// 4. Purchase
	purchaseMenu := &permission.Menu{
		Name:   "Purchase",
		Icon:   "truck",
		URL:    "/purchase",
		Order:  4,
		Status: "active",
		Access: false,
	}
	if err := createMenu(purchaseMenu); err != nil {
		return err
	}

	// 5. Stock
	stockMenu := &permission.Menu{
		Name:   "Stock",
		Icon:   "warehouse",
		URL:    "/stock",
		Order:  5,
		Status: "active",
		Access: false,
	}
	if err := createMenu(stockMenu); err != nil {
		return err
	}

	// 6. Finance
	financeMenu := &permission.Menu{
		Name:   "Finance",
		Icon:   "credit-card",
		URL:    "/finance",
		Order:  6,
		Status: "active",
		Access: false,
	}
	if err := createMenu(financeMenu); err != nil {
		return err
	}

	// 7. HRD
	hrdMenu := &permission.Menu{
		Name:   "HRD",
		Icon:   "users",
		URL:    "/hrd",
		Order:  7,
		Status: "active",
		Access: false,
	}
	if err := createMenu(hrdMenu); err != nil {
		return err
	}

	// 8. Reports
	reportsMenu := &permission.Menu{
		Name:   "Reports",
		Icon:   "bar-chart-3",
		URL:    "/reports",
		Order:  8,
		Status: "active",
		Access: false,
	}
	if err := createMenu(reportsMenu); err != nil {
		return err
	}

	// 9. AI Assistant
	aiMenu := &permission.Menu{
		Name:   "AI Assistant",
		Icon:   "sparkles",
		URL:    "/ai-assistant",
		Order:  10,
		Status: "active",
		Access: false,
	}
	if err := createMenu(aiMenu); err != nil {
		return err
	}

	// 10. CRM
	crmMenu := &permission.Menu{
		Name:   "CRM",
		Icon:   "handshake",
		URL:    "/crm",
		Order:  2,
		Status: "active",
		Access: false,
	}
	if err := createMenu(crmMenu); err != nil {
		return err
	}

	// 11. Travel Planner
	travelPlannerMenu := &permission.Menu{
		Name:   "Travel Planner",
		Icon:   "route",
		URL:    "/travel/travel-planner",
		Order:  11,
		Status: "active",
		Access: false,
	}
	if err := createMenu(travelPlannerMenu); err != nil {
		return err
	}

	// 12. POS
	posMenu := &permission.Menu{
		Name:   "POS",
		Icon:   "store",
		URL:    "/pos",
		Order:  12,
		Status: "active",
		Access: false,
	}
	if err := createMenu(posMenu); err != nil {
		return err
	}

	// ============================================================
	// MASTER DATA SUB-MENUS
	// ============================================================

	// Geographic - single read-only map page (no CRUD sub-pages)
	if _, err := createChildMenu("Geographic", "globe", "/master-data/geographic", &masterDataMenu.ID, 1); err != nil {
		return err
	}

	// Organization Group
	organizationMenu, err := createChildMenu("Organization", "building-2", "/master-data/organization", &masterDataMenu.ID, 2)
	if err != nil {
		return err
	}

	organizationChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Company", "briefcase", "/master-data/company", 1},
		{"Outlets", "store", "/master-data/outlet", 2},
		{"Divisions", "layers", "/master-data/divisions", 3},
		{"Job Positions", "user-cog", "/master-data/job-positions", 4},
		{"Business Units", "grid", "/master-data/business-units", 5},
		{"Business Types", "tag", "/master-data/business-types", 6},
		{"Areas", "map", "/master-data/areas", 7},
	}
	for _, child := range organizationChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &organizationMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Employee
	if _, err := createChildMenu("Employees", "users", "/master-data/employees", &masterDataMenu.ID, 3); err != nil {
		return err
	}

	// Supplier Group
	supplierMenu, err := createChildMenu("Supplier", "truck", "/master-data/supplier", &masterDataMenu.ID, 4)
	if err != nil {
		return err
	}

	supplierChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Suppliers", "building-2", "/master-data/suppliers", 1},
		{"Supplier Types", "tag", "/master-data/supplier-types", 2},
		{"Banks", "landmark", "/master-data/banks", 3},
	}
	for _, child := range supplierChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &supplierMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Customer Group
	customerMenu, err := createChildMenu("Customer", "users-round", "/master-data/customer", &masterDataMenu.ID, 6)
	if err != nil {
		return err
	}

	customerChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Customers", "user-check", "/master-data/customers", 1},
		{"Customer Types", "tag", "/master-data/customer-types", 2},
	}
	for _, child := range customerChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &customerMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Product Group
	productMenu, err := createChildMenu("Product", "package", "/master-data/product", &masterDataMenu.ID, 5)
	if err != nil {
		return err
	}

	productChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Products", "package", "/master-data/products", 1},
		{"Categories", "folder-tree", "/master-data/product-categories", 2},
		{"Brands", "star", "/master-data/product-brands", 3},
		{"Segments", "pie-chart", "/master-data/product-segments", 4},
		{"Types", "tag", "/master-data/product-types", 5},
		{"Packaging", "box", "/master-data/packaging", 6},
		{"Unit of Measure", "ruler", "/master-data/uom", 7},
		{"Procurement Types", "shopping-bag", "/master-data/procurement-types", 8},
	}
	for _, child := range productChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &productMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Warehouse
	if _, err := createChildMenu("Warehouses", "warehouse", "/master-data/warehouses", &masterDataMenu.ID, 6); err != nil {
		return err
	}

	// Payment & Courier
	paymentMenu, err := createChildMenu("Payment & Courier", "credit-card", "/master-data/payment-courier", &masterDataMenu.ID, 7)
	if err != nil {
		return err
	}

	paymentChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Currencies", "coins", "/master-data/currencies", 1},
		{"Banks", "landmark", "/master-data/banks", 2},
		{"Payment Terms", "clock", "/master-data/payment-terms", 3},
		{"Courier Agencies", "truck", "/master-data/courier-agencies", 4},
		{"SO Sources", "file-text", "/master-data/so-sources", 5},
	}
	for _, child := range paymentChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &paymentMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Leave Types
	if _, err := createChildMenu("Leave Types", "calendar", "/master-data/leave-types", &masterDataMenu.ID, 8); err != nil {
		return err
	}

	// Contact Roles (CRM Settings surfaced in Master Data)
	if _, err := createChildMenu("Contact Roles", "users", "/master-data/contact-roles", &masterDataMenu.ID, 9); err != nil {
		return err
	}

	// Users (RBAC)
	if _, err := createChildMenu("Users", "users", "/master-data/users", &masterDataMenu.ID, 99); err != nil {
		return err
	}

	// ============================================================
	// SALES SUB-MENUS
	// ============================================================

	salesChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Quotations", "file-text", "/sales/quotations", 1},
		{"Sales Orders", "shopping-cart", "/sales/orders", 2},
		{"Delivery Orders", "truck", "/sales/delivery-orders", 3},
		{"Customer Invoices", "receipt", "/sales/invoices", 4},
		{"Customer Invoices Down Payments", "banknote", "/sales/customer-invoice-down-payments", 5},
		{"Returns", "rotate-ccw", "/sales/returns", 6},
		{"Payments", "credit-card", "/sales/payments", 9},
		{"Receivables Recap", "bar-chart-3", "/sales/receivables-recap", 10},
	}
	for _, child := range salesChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &salesMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// PURCHASE SUB-MENUS
	// ============================================================

	purchaseChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Requisitions", "clipboard-list", "/purchase/purchase-requisitions", 1},
		{"Purchase Orders", "file-text", "/purchase/purchase-orders", 2},
		{"Goods Receipt", "package", "/purchase/goods-receipt", 3},
		{"Supplier Invoices", "receipt", "/purchase/supplier-invoices", 4},
		{"Supplier Invoice Down Payments", "receipt", "/purchase/supplier-invoice-down-payments", 5},
		{"Returns", "rotate-ccw", "/purchase/returns", 6},
		{"Payments", "credit-card", "/purchase/payments", 7},
		{"Payable Recap", "bar-chart-3", "/purchase/payable-recap", 8},
	}
	for _, child := range purchaseChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &purchaseMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// STOCK SUB-MENUS
	// ============================================================

	stockChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Inventory", "package", "/stock/inventory", 1},
		{"Stock Ledger", "scroll-text", "/stock/ledger", 2},
		{"Stock Movement", "arrow-right-left", "/stock/movements", 3},
		{"Stock Opname", "clipboard-check", "/stock/opname", 4},
	}
	for _, child := range stockChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &stockMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// FINANCE SUB-MENUS
	// ============================================================
	// Parent groups (hierarchy containers)
	accountingMenu, err := createChildMenu("Accounting", "book-open", "", &financeMenu.ID, 1)
	if err != nil {
		return err
	}

	arMenu, err := createChildMenu("Accounts Receivable (AR)", "receipt", "", &financeMenu.ID, 2)
	if err != nil {
		return err
	}

	apMenu, err := createChildMenu("Accounts Payable (AP)", "receipt", "", &financeMenu.ID, 3)
	if err != nil {
		return err
	}

	cashBankMenu, err := createChildMenu("Cash & Bank", "landmark", "", &financeMenu.ID, 4)
	if err != nil {
		return err
	}

	fixedAssetsMenu, err := createChildMenu("Fixed Assets", "building-2", "/finance/fixed-assets", &financeMenu.ID, 5)
	if err != nil {
		return err
	}

	financialReportsMenu, err := createChildMenu("Financial Reports", "bar-chart-3", "", &financeMenu.ID, 6)
	if err != nil {
		return err
	}

	financeSettingsMenu, err := createChildMenu("Finance Settings", "settings", "/finance/settings", &financeMenu.ID, 7)
	if err != nil {
		return err
	}

	accountingChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Chart of Accounts", "list", "/finance/accounting/coa", 1},
		{"Financial Closing", "lock", "/finance/accounting/closing", 2},
	}
	for _, child := range accountingChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &accountingMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Journal Group under Accounting
	journalMenu, err := createChildMenu("Journal", "book-open", "", &accountingMenu.ID, 3)
	if err != nil {
		return err
	}
	journalSubMenus := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Journal Entries", "file-text", "/finance/accounting/journal-entries", 1},
		{"Sales Journal", "receipt", "/finance/journals/sales", 2},
		{"Purchase Journal", "shopping-cart", "/finance/accounting/journal-entries/purchase", 3},
		{"Adjustment Journal", "edit-3", "/finance/accounting/journal-entries/adjustment", 4},
	}
	for _, sub := range journalSubMenus {
		if _, err := createChildMenu(sub.name, sub.icon, sub.url, &journalMenu.ID, sub.order); err != nil {
			return err
		}
	}

	arChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"AR Aging Reports", "clock", "/finance/ar/aging-reports", 1},
	}
	for _, child := range arChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &arMenu.ID, child.order); err != nil {
			return err
		}
	}

	removedARChildrenURLs := []string{
		"/finance/ar/customer-invoices",
		"/finance/ar/customer-payments",
		"/finance/ar/credit-notes",
	}
	for _, url := range removedARChildrenURLs {
		if err := deactivateMenuURLByParent(url, &arMenu.ID); err != nil {
			return err
		}
	}

	apChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Non-Trade Payables", "file-minus", "/finance/ap/non-trade-payables", 1},
		{"AP Aging Reports", "clock", "/finance/ap/aging-reports", 2},
	}
	for _, child := range apChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &apMenu.ID, child.order); err != nil {
			return err
		}
	}

	removedAPChildrenURLs := []string{
		"/finance/ap/supplier-invoices",
		"/finance/ap/supplier-payments",
		"/finance/ap/debit-notes",
	}
	for _, url := range removedAPChildrenURLs {
		if err := deactivateMenuURLByParent(url, &apMenu.ID); err != nil {
			return err
		}
	}

	cashBankChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Bank Accounts", "landmark", "/finance/bank-accounts", 1},
		{"Cash-Bank Transactions", "book-open", "/finance/journals/cash-bank", 2},
		{"Bank Transfers", "arrow-right-left", "/finance/bank-transfer", 3},
		{"Bank Reconciliation", "scale", "/finance/bank-reconciliation", 4},
	}
	for _, child := range cashBankChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &cashBankMenu.ID, child.order); err != nil {
			return err
		}
	}

	fixedAssetsChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Asset Register", "building-2", "/finance/fixed-assets/assets", 1},
		{"Asset Categories", "folder-tree", "/finance/fixed-assets/categories", 2},
		{"Asset Locations", "map-pin", "/finance/fixed-assets/locations", 3},
		{"Depreciation Schedule", "calendar", "/finance/fixed-assets/depreciation-schedule", 4},
		{"Asset Disposal", "hammer", "/finance/fixed-assets/disposal", 5},
		{"Asset Revaluation", "scale", "/finance/fixed-assets/revaluation", 6},
	}
	for _, child := range fixedAssetsChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &fixedAssetsMenu.ID, child.order); err != nil {
			return err
		}
	}

	financialReportsChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"General Ledger", "book-open", "/finance/reports/general-ledger", 1},
		{"Trial Balance", "scale", "/finance/reports/trial-balance", 2},
		{"Balance Sheet", "scale", "/finance/reports/balance-sheet", 3},
		{"Profit & Loss", "trending-up", "/finance/reports/profit-loss", 4},
		{"Cash Flow Statement", "bar-chart-2", "/finance/reports/cash-flow-statement", 5},
	}
	for _, child := range financialReportsChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &financialReportsMenu.ID, child.order); err != nil {
			return err
		}
	}

	financeSettingsChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Accounting Mapping", "file-text", "/finance/settings/accounting-mapping", 1},
		{"Fiscal Years", "calendar", "/finance/settings/fiscal-years", 2},
		{"Tax Configuration", "receipt", "/finance/settings/tax-config", 3},
		{"Inventory Settings", "boxes", "/finance/settings/inventory", 4},
		{"Opening Balance", "wallet", "/finance/settings/opening-balance", 5},
	}
	for _, child := range financeSettingsChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &financeSettingsMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// TRAVEL PLANNER SUB-MENUS
	// ============================================================

	travelPlannerChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Planner Workspace", "map", "/travel/travel-planner", 1},
		{"Visit Planner", "map-pin", "/travel/visit-planner", 2},
	}
	for _, child := range travelPlannerChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &travelPlannerMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// POS SUB-MENUS
	// ============================================================

	if _, err := createChildMenu("Floor & Layout Designer", "layout-list", "/pos/fb/floor-layout", &posMenu.ID, 1); err != nil {
		return err
	}

	if _, err := createChildMenu("POS Terminal", "monitor-check", "/pos/fb/terminal", &posMenu.ID, 2); err != nil {
		return err
	}

	// Feedback (Customer Feedback via POS QR Code)
	feedbackMenu, err := createChildMenu("Feedback", "message-square-heart", "/pos/feedback", &posMenu.ID, 3)
	if err != nil {
		return err
	}

	feedbackChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Feedback Responses", "message-circle", "/pos/feedback/response", 1},
		{"Form Builder", "clipboard-pen", "/pos/feedback/forms", 2},
	}
	for _, child := range feedbackChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &feedbackMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Loyalty (POS member lookup, enrollment, and member management)
	loyaltyMenu, err := createChildMenu("Loyalty", "star", "/pos/loyalty", &posMenu.ID, 4)
	if err != nil {
		return err
	}

	loyaltyChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Programs", "settings", "/pos/loyalty/config", 1},
		{"Members", "users", "/pos/loyalty/members", 2},
	}
	for _, child := range loyaltyChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &loyaltyMenu.ID, child.order); err != nil {
			return err
		}
	}

	// Journal Group (6 domain journal pages — Journal Lines removed, merged into Journal Entries)
	financeJournalMenu, err := createChildMenu("Journal", "book-open", "", &financeMenu.ID, 2)
	if err != nil {
		return err
	}
	financeJournalSubMenus := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Journal Entries", "file-text", "/finance/journals", 1},
		{"Sales Journal", "receipt", "/finance/journals/sales", 2},
		{"Purchase Journal", "shopping-cart", "/finance/journals/purchase", 3},
		{"Adjustment Journal", "edit-3", "/finance/journals/adjustment", 4},
		{"Journal Valuation", "calculator", "/finance/journals/valuation", 5},
		{"Cash Transactions (Journal View)", "banknote", "/finance/journals/cash-bank", 6},
	}
	for _, sub := range financeJournalSubMenus {
		if _, err := createChildMenu(sub.name, sub.icon, sub.url, &financeJournalMenu.ID, sub.order); err != nil {
			return err
		}
	}

	// Legacy finance menu cleanup after hierarchy refactor.
	legacyFinanceChildURLs := []string{
		"/finance/coa",
		"/finance/bank-accounts",
		"/finance/payments",
		"/finance/tax-invoices",
		"/finance/non-trade-payables",
		"/finance/ap/payments",
		"/finance/reports/aging",
		"/finance/reports/reconciliation/arap",
		"/finance/budget",
		"/finance/closing",
		"/finance/asset-categories",
		"/finance/asset-locations",
		"/finance/asset-budgets",
		"/finance/asset-maintenance",
		"/finance/up-country-cost",
		"/finance/salary",
		"/finance/reconciliation/arap",
		"/finance/reports",
		"/finance/reports/general-ledger",
		"/finance/reports/balance-sheet",
		"/finance/reports/profit-loss",
		"/finance/aging-reports",
		"/finance/settings/payment-terms",
		"/finance/settings/currency",
	}
	for _, url := range legacyFinanceChildURLs {
		if err := deactivateMenuURLByParent(url, &financeMenu.ID); err != nil {
			return err
		}
	}

	if err := deactivateMenuByNameAndParent("Journal", &financeMenu.ID); err != nil {
		return err
	}
	if err := deactivateMenuChildrenByNameAndParent("Journal", &financeMenu.ID); err != nil {
		return err
	}
	if err := deactivateMenuByNameAndParent("Reports", &financeMenu.ID); err != nil {
		return err
	}
	if err := deactivateMenuChildrenByNameAndParent("Reports", &financeMenu.ID); err != nil {
		return err
	}

	legacyFinanceGroups := []string{
		"Receivables and Payables",
		"Budgeting and Cost",
		"Asset Management",
		"Financial Statements",
		"Taxation",
	}
	for _, groupName := range legacyFinanceGroups {
		if err := deactivateMenuByNameAndParent(groupName, &financeMenu.ID); err != nil {
			return err
		}
		if err := deactivateMenuChildrenByNameAndParent(groupName, &financeMenu.ID); err != nil {
			return err
		}
	}

	// ============================================================
	// HRD SUB-MENUS
	// ============================================================

	hrdChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Attendance", "clock", "/hrd/attendance", 1},
		{"Leave Requests", "calendar", "/hrd/leave-requests", 2},
		{"Overtime", "clock-arrow-up", "/hrd/overtime", 3},
		{"Evaluation", "star", "/hrd/evaluation", 5},
		{"Recruitment", "user-plus", "/hrd/recruitment", 6},
		{"Work Schedule", "calendar-days", "/hrd/work-schedule", 7},
		{"Holidays", "calendar-check", "/hrd/holidays", 8},
		{"Salary Structures", "wallet", "/hrd/salary-structures", 9},
	}
	for _, child := range hrdChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &hrdMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// AI ASSISTANT SUB-MENUS
	// ============================================================

	aiChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"AI Chatbot", "bot", "/ai-chatbot", 1},
	}
	for _, child := range aiChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &aiMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// CRM SUB-MENUS
	// ============================================================

	// CRM Leads menu (Sprint 19)
	if _, err := createChildMenu("Leads", "user-plus", "/crm/leads", &crmMenu.ID, 1); err != nil {
		return err
	}

	// CRM Pipeline menu (Sprint 20)
	if _, err := createChildMenu("Pipeline", "kanban", "/crm/pipeline", &crmMenu.ID, 2); err != nil {
		return err
	}

	// CRM Activities menu (Sprint 23)
	if _, err := createChildMenu("Activities", "activity", "/crm/activities", &crmMenu.ID, 3); err != nil {
		return err
	}

	// CRM Tasks menu (Sprint 23)
	if _, err := createChildMenu("Tasks", "check-square", "/crm/tasks", &crmMenu.ID, 4); err != nil {
		return err
	}

	// CRM Schedules menu (Sprint 23)
	if _, err := createChildMenu("Schedules", "calendar", "/crm/schedules", &crmMenu.ID, 5); err != nil {
		return err
	}

	// CRM Visit Reports menu (Sprint 22)
	if _, err := createChildMenu("Visit Reports", "map-pin", "/crm/visits", &crmMenu.ID, 6); err != nil {
		return err
	}

	// CRM Area Mapping menu (Sprint 24)
	if _, err := createChildMenu("Area Mapping", "map", "/crm/area-mapping", &crmMenu.ID, 7); err != nil {
		return err
	}

	// CRM Sales Targets menu
	if _, err := createChildMenu("Sales Targets", "target", "/crm/sales-targets", &crmMenu.ID, 8); err != nil {
		return err
	}

	// CRM Settings Group
	crmSettingsMenu, err := createChildMenu("CRM Settings", "settings", "/crm/settings", &crmMenu.ID, 10)
	if err != nil {
		return err
	}

	crmSettingsChildren := []struct {
		name  string
		icon  string
		url   string
		order int
	}{
		{"Pipeline Stages", "git-branch", "/crm/settings/pipeline-stages", 1},
		{"Lead Sources", "target", "/crm/settings/lead-sources", 2},
		{"Lead Statuses", "tag", "/crm/settings/lead-statuses", 3},
		{"Activity Types", "calendar-check", "/crm/settings/activity-types", 4},
	}
	for _, child := range crmSettingsChildren {
		if _, err := createChildMenu(child.name, child.icon, child.url, &crmSettingsMenu.ID, child.order); err != nil {
			return err
		}
	}

	// ============================================================
	// REPORTS SUB-MENUS
	// ============================================================

	if _, err := createChildMenu("Sales Overview", "trending-up", "/reports/sales-overview", &reportsMenu.ID, 1); err != nil {
		return err
	}
	if _, err := createChildMenu("Top Product", "bar-chart-3", "/reports/product-analysis", &reportsMenu.ID, 2); err != nil {
		return err
	}
	if _, err := createChildMenu("Geo Performance", "map", "/reports/geo-performance", &reportsMenu.ID, 3); err != nil {
		return err
	}
	if _, err := createChildMenu("Customer Research", "users", "/reports/customer-research", &reportsMenu.ID, 4); err != nil {
		return err
	}
	if _, err := createChildMenu("Supplier Research", "truck", "/reports/supplier-research", &reportsMenu.ID, 5); err != nil {
		return err
	}

	log.Println("ERP menus seeded successfully")
	return nil
}

func migrateMenuURL(oldURL, newURL string) error {
	var oldMenu permission.Menu
	if err := database.DB.Where("url = ?", oldURL).First(&oldMenu).Error; err != nil {
		return nil
	}

	var newMenu permission.Menu
	if err := database.DB.Where("url = ?", newURL).First(&newMenu).Error; err == nil {
		if err := database.DB.Model(&oldMenu).Updates(map[string]interface{}{
			"status": "inactive",
		}).Error; err != nil {
			return err
		}
		log.Printf("Menu URL migration skipped (target exists): %s -> %s", oldURL, newURL)
		return nil
	}

	if err := database.DB.Model(&oldMenu).Updates(map[string]interface{}{
		"url":    newURL,
		"status": "active",
	}).Error; err != nil {
		return err
	}
	log.Printf("Migrated menu URL: %s -> %s", oldURL, newURL)
	return nil
}

func deactivateMenuURL(url string) error {
	var legacyMenu permission.Menu
	if err := database.DB.Where("url = ?", url).First(&legacyMenu).Error; err != nil {
		return nil
	}
	if legacyMenu.Status == "inactive" {
		return nil
	}
	if err := database.DB.Model(&legacyMenu).Updates(map[string]interface{}{
		"status": "inactive",
	}).Error; err != nil {
		return err
	}
	log.Printf("Deprecated menu URL: %s", url)
	return nil
}

func deactivateMenuURLByParent(url string, parentID *string) error {
	query := database.DB.Where("url = ?", url)
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var legacyMenus []permission.Menu
	if err := query.Find(&legacyMenus).Error; err != nil {
		return err
	}

	for _, legacyMenu := range legacyMenus {
		if legacyMenu.Status == "inactive" {
			continue
		}
		if err := database.DB.Model(&legacyMenu).Updates(map[string]interface{}{
			"status": "inactive",
		}).Error; err != nil {
			return err
		}
		log.Printf("Deprecated menu URL by parent: %s", url)
	}

	return nil
}

func deactivateMenuByNameAndParent(name string, parentID *string) error {
	query := database.DB.Where("name = ?", name)
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var legacyMenus []permission.Menu
	if err := query.Find(&legacyMenus).Error; err != nil {
		return err
	}

	for _, legacyMenu := range legacyMenus {
		if legacyMenu.Status == "inactive" {
			continue
		}
		if err := database.DB.Model(&legacyMenu).Updates(map[string]interface{}{
			"status": "inactive",
		}).Error; err != nil {
			return err
		}
		log.Printf("Deprecated menu by name and parent: %s", name)
	}

	return nil
}

func deactivateMenuChildrenByNameAndParent(name string, parentID *string) error {
	query := database.DB.Where("name = ?", name)
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var parentMenus []permission.Menu
	if err := query.Find(&parentMenus).Error; err != nil {
		return err
	}

	for _, parentMenu := range parentMenus {
		if err := deactivateMenuTreeByParentID(parentMenu.ID); err != nil {
			return err
		}
	}

	return nil
}

func deactivateMenuTreeByParentID(parentID string) error {
	var children []permission.Menu
	if err := database.DB.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		if err := deactivateMenuTreeByParentID(child.ID); err != nil {
			return err
		}
		if child.Status == "inactive" {
			continue
		}
		if err := database.DB.Model(&child).Updates(map[string]interface{}{
			"status": "inactive",
		}).Error; err != nil {
			return err
		}
		log.Printf("Deprecated child menu: %s (%s)", child.Name, child.URL)
	}

	return nil
}

// UpdateMenuStructure updates existing menu structure (migration helper)
func UpdateMenuStructure() error {
	log.Println("Updating menu structure (migration helper)...")

	type urlMigration struct {
		oldURL string
		newURL string
	}

	// Keep this list small and surgical: only known historical paths.
	migrations := []urlMigration{
		{oldURL: "/purchase/orders", newURL: "/purchase/purchase-orders"},
		{oldURL: "/purchase/requisitions", newURL: "/purchase/purchase-requisitions"},
		{oldURL: "/pos/floor-layout", newURL: "/pos/fb/floor-layout"},
	}

	for _, m := range migrations {
		if err := migrateMenuURL(m.oldURL, m.newURL); err != nil {
			return err
		}
	}

	// Deprecate Sales Estimation menu — replaced by CRM Pipeline
	deprecatedMenuURLs := []string{
		"/sales/estimations",
		"/master-data/area-supervisors",
		"/pos/fb",
		"/pos/goods",
		"/hrd/documents",
		"/hrd/contracts",
		"/hrd/education",
		"/hrd/certifications",
		"/ai-settings",
		"/finance/journal-lines",
		"/finance/reports",
		"/finance/cash-bank",
	}

	for _, deprecatedURL := range deprecatedMenuURLs {
		if err := deactivateMenuURL(deprecatedURL); err != nil {
			return err
		}
	}

	log.Println("Menu structure update completed")
	return nil
}
