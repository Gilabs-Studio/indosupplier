package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
	coreUsecase "github.com/gilabs/gims/api/internal/core/domain/usecase"
	financeDTO "github.com/gilabs/gims/api/internal/finance/domain/dto"
	financeUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	hrdDTO "github.com/gilabs/gims/api/internal/hrd/domain/dto"
	hrdUsecase "github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	inventoryDTO "github.com/gilabs/gims/api/internal/inventory/domain/dto"
	inventoryUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	purchaseRepos "github.com/gilabs/gims/api/internal/purchase/data/repositories"
	purchaseDTO "github.com/gilabs/gims/api/internal/purchase/domain/dto"
	purchaseUsecase "github.com/gilabs/gims/api/internal/purchase/domain/usecase"
	salesDTO "github.com/gilabs/gims/api/internal/sales/domain/dto"
	salesUsecase "github.com/gilabs/gims/api/internal/sales/domain/usecase"
)

// ActionResult holds the outcome of executing an AI-requested action
type ActionResult struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data,omitempty"`
	Message      string      `json:"message"`
	EntityType   string      `json:"entity_type,omitempty"`
	EntityID     string      `json:"entity_id,omitempty"`
	Action       string      `json:"action"`
	DurationMs   int64       `json:"duration_ms"`
	ErrorCode    string      `json:"error_code,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

const (
	dbServiceUnavailableMessage             = "Database service is not available"
	leaveRequestServiceUnavailableMessage   = "Leave request service is not available"
	salesQuotationServiceUnavailableMessage = "Sales quotation service is not available"
	salesOrderServiceUnavailableMessage     = "Sales order service is not available"
	purchaseOrderServiceUnavailableMessage  = "Purchase order service is not available"
	invalidParamsMessage                    = "Invalid parameters"
)

// ActionExecutorDeps holds the domain usecase dependencies for the executor
type ActionExecutorDeps struct {
	HolidayUsecase             hrdUsecase.HolidayUsecase
	LeaveRequestUsecase        hrdUsecase.LeaveRequestUsecase
	AttendanceUsecase          hrdUsecase.AttendanceRecordUsecase
	SalesQuotationUsecase      salesUsecase.SalesQuotationUsecase
	SalesOrderUsecase          salesUsecase.SalesOrderUsecase
	DeliveryOrderUsecase       salesUsecase.DeliveryOrderUsecase
	CustomerInvoiceUsecase     salesUsecase.CustomerInvoiceUsecase
	InventoryUsecase           inventoryUsecase.InventoryUsecase
	PurchaseOrderUsecase       purchaseUsecase.PurchaseOrderUsecase
	PurchaseRequisitionUsecase purchaseUsecase.PurchaseRequisitionUsecase
	GoodsReceiptUsecase        purchaseUsecase.GoodsReceiptUsecase
	SupplierInvoiceUsecase     purchaseUsecase.SupplierInvoiceUsecase
	CoaUsecase                 financeUsecase.ChartOfAccountUsecase
	JournalUsecase             financeUsecase.JournalEntryUsecase
	FinancePaymentUsecase      financeUsecase.PaymentUsecase
	BudgetUsecase              financeUsecase.BudgetUsecase
	CashBankUsecase            financeUsecase.CashBankJournalUsecase
	TaxInvoiceUsecase          financeUsecase.TaxInvoiceUsecase
	AssetUsecase               financeUsecase.AssetUsecase
	SalaryUsecase              hrdUsecase.SalaryStructureUsecase
	BankAccountUsecase         coreUsecase.BankAccountUsecase
}

// ActionExecutor dispatches resolved intents to the appropriate domain usecases
type ActionExecutor struct {
	deps           *ActionExecutorDeps
	entityResolver *EntityResolver
}

// NewActionExecutor creates a new ActionExecutor
func NewActionExecutor(deps *ActionExecutorDeps, entityResolver *EntityResolver) *ActionExecutor {
	return &ActionExecutor{
		deps:           deps,
		entityResolver: entityResolver,
	}
}

// Execute dispatches the action based on intent code and parameters
func (e *ActionExecutor) Execute(ctx context.Context, intent *IntentResult, resolvedEntities map[string]*ResolvedEntity, currentUserID string) *ActionResult {
	start := apptime.Now()

	var result *ActionResult
	switch intent.IntentCode {
	// ==========================================
	//  HRD Module — Holidays
	// ==========================================
	case "CREATE_HOLIDAY":
		result = e.executeCreateHoliday(ctx, intent.Parameters)
	case "LIST_HOLIDAYS":
		result = e.executeListHolidays(ctx, intent.Parameters)

	// ==========================================
	//  HRD Module — Leave
	// ==========================================
	case "CREATE_LEAVE_REQUEST":
		result = e.executeCreateLeaveRequest(ctx, intent.Parameters, currentUserID)
	case "LIST_LEAVE_REQUESTS":
		result = e.executeListLeaveRequests(ctx, intent.Parameters, currentUserID)
	case "APPROVE_LEAVE_REQUEST":
		result = e.executeApproveLeaveRequest(ctx, intent.Parameters, currentUserID)
	case "REJECT_LEAVE_REQUEST":
		result = e.executeRejectLeaveRequest(ctx, intent.Parameters, currentUserID)

	// ==========================================
	//  HRD Module — Attendance
	// ==========================================
	case "QUERY_ATTENDANCE":
		result = e.executeQueryAttendance(ctx, intent.Parameters, resolvedEntities)

	// ==========================================
	//  HRD Module — Employees, Contracts, etc.
	// ==========================================
	case "LIST_EMPLOYEES", "QUERY_EMPLOYEE":
		if intent.IntentCode == "LIST_EMPLOYEES" {
			result = e.executeListEmployees(ctx, intent.Parameters)
		} else {
			result = e.executeQueryEmployee(ctx, intent.Parameters)
		}
	case "LIST_CONTRACTS":
		result = e.executeSimpleListByTable(ctx, "employee_contracts", "employee_contract", "contracts", []string{"contract_number", "status", "contract_type"}, intent.Parameters)
	case "QUERY_CONTRACT":
		result = e.executeSimpleQueryByTable(ctx, "employee_contracts", "employee_contract", []string{"contract_number", "status", "contract_type"}, intent.Parameters)
	case "CREATE_CONTRACT":
		result = e.executeSimpleCreateByTable(ctx, "employee_contracts", "employee_contract", []string{"employee_id", "contract_type", "contract_number", "start_date", "end_date", "status", "salary_amount", "notes", "created_by"}, []string{"employee_id", "contract_type", "contract_number", "start_date", "status"}, intent.Parameters)
	case "LIST_OVERTIME":
		result = e.executeSimpleListByTable(ctx, "overtime_requests", "overtime", "overtime_requests", []string{"reason", "request_type", "status"}, intent.Parameters)
	case "CREATE_OVERTIME":
		result = e.executeSimpleCreateByTable(ctx, "overtime_requests", "overtime", []string{"employee_id", "date", "request_type", "start_time", "end_time", "reason", "description", "task_details", "status", "approved_by", "approved_minutes"}, []string{"employee_id", "date", "start_time", "end_time", "reason"}, intent.Parameters)
	case "APPROVE_OVERTIME":
		result = e.executeSimpleStatusUpdateByTable(ctx, simpleStatusUpdateRequest{
			table:        "overtime_requests",
			entityType:   "overtime",
			idField:      "id",
			statusField:  "status",
			statusValue:  "APPROVED",
			params:       intent.Parameters,
			extraUpdates: map[string]interface{}{"approved_by": currentUserID},
		})
	case "LIST_WORK_SCHEDULES":
		result = e.executeSimpleListByTable(ctx, "work_schedules", "work_schedule", "work_schedules", []string{"name", "description"}, intent.Parameters)
	case "CREATE_WORK_SCHEDULE":
		result = e.executeSimpleCreateByTable(ctx, "work_schedules", "work_schedule", []string{"name", "description", "division_id", "is_default", "is_active", "start_time", "end_time", "is_flexible", "flexible_start_time", "flexible_end_time", "working_days", "working_hours_per_day", "late_tolerance_minutes", "early_leave_tolerance_minutes", "require_gps", "gps_radius_meter", "office_latitude", "office_longitude"}, []string{"name", "start_time", "end_time"}, intent.Parameters)
	case "LIST_RECRUITMENTS":
		result = e.executeSimpleListByTable(ctx, "recruitment_requests", "recruitment", "recruitments", []string{"request_code", "status", "priority", "employment_type"}, intent.Parameters)
	case "CREATE_RECRUITMENT":
		result = e.executeSimpleCreateByTable(ctx, "recruitment_requests", "recruitment", []string{"request_code", "requested_by_id", "request_date", "division_id", "position_id", "required_count", "employment_type", "expected_start_date", "salary_range_min", "salary_range_max", "job_description", "qualifications", "notes", "priority", "status", "created_by"}, []string{"requested_by_id", "request_date", "division_id", "position_id", "required_count", "expected_start_date", "job_description", "qualifications"}, intent.Parameters)
	case "LIST_EVALUATIONS":
		result = e.executeSimpleListByTable(ctx, "employee_evaluations", "employee_evaluation", "evaluations", []string{"status", "evaluation_type"}, intent.Parameters)
	case "LIST_CERTIFICATIONS":
		result = e.executeSimpleListByTable(ctx, "employee_certifications", "employee_certification", "certifications", []string{"certificate_name", "issued_by", "certificate_number", "status"}, intent.Parameters)
	case "LIST_EDUCATION_HISTORY":
		result = e.executeSimpleListByTable(ctx, "employee_education_histories", "employee_education_history", "education_histories", []string{"institution_name", "degree", "field_of_study"}, intent.Parameters)
	case "LIST_EMPLOYEE_ASSETS":
		result = e.executeSimpleListByTable(ctx, "employee_assets", "employee_asset", "employee_assets", []string{"asset_name", "asset_code", "status"}, intent.Parameters)
	case "LIST_LEAVE_TYPES":
		result = e.executeSimpleListByTable(ctx, "leave_types", "leave_type", "leave_types", []string{"name", "code", "description"}, intent.Parameters)

	// ==========================================
	//  Sales Module
	// ==========================================
	case "CREATE_SALES_QUOTATION":
		result = e.executeCreateSalesQuotation(ctx, intent.Parameters, currentUserID, resolvedEntities)
	case "LIST_SALES_QUOTATIONS":
		result = e.executeListSalesQuotations(ctx, intent.Parameters)
	case "APPROVE_SALES_QUOTATION":
		result = e.executeApproveSalesQuotation(ctx, intent.Parameters, currentUserID)
	case "CREATE_SALES_ORDER":
		result = e.executeCreateSalesOrder(ctx, intent.Parameters, currentUserID)
	case "LIST_SALES_ORDERS":
		result = e.executeListSalesOrders(ctx, intent.Parameters)
	case "QUERY_SALES_ORDER":
		result = e.executeQuerySalesOrder(ctx, intent.Parameters)
	case "LIST_DELIVERY_ORDERS":
		result = e.executeListDeliveryOrders(ctx, intent.Parameters)
	case "CREATE_DELIVERY_ORDER":
		result = e.executeCreateDeliveryOrder(ctx, intent.Parameters, currentUserID)
	case "LIST_SALES_INVOICES":
		result = e.executeListSalesInvoices(ctx, intent.Parameters)
	case "CREATE_SALES_INVOICE":
		result = e.executeCreateSalesInvoice(ctx, intent.Parameters, currentUserID)
	case "LIST_SALES_VISITS":
		result = e.executeSimpleListByTable(ctx, "sales_visits", "sales_visit", "sales_visits", []string{"code", "contact_person", "status", "purpose"}, intent.Parameters)
	case "LIST_SALES_ESTIMATIONS":
		result = e.executeListSalesQuotations(ctx, intent.Parameters)

	// ==========================================
	//  Purchase Module
	// ==========================================
	case "LIST_PURCHASE_REQUISITIONS":
		result = e.executeListPurchaseRequisitions(ctx, intent.Parameters)
	case "CREATE_PURCHASE_REQUISITION":
		result = e.executeCreatePurchaseRequisition(ctx, intent.Parameters)
	case "LIST_PURCHASE_ORDERS":
		result = e.executeListPurchaseOrders(ctx, intent.Parameters)
	case "CREATE_PURCHASE_ORDER":
		result = e.executeCreatePurchaseOrder(ctx, intent.Parameters)
	case "APPROVE_PURCHASE_ORDER":
		result = e.executeApprovePurchaseOrder(ctx, intent.Parameters)
	case "LIST_GOODS_RECEIPTS":
		result = e.executeListGoodsReceipts(ctx, intent.Parameters)
	case "LIST_SUPPLIER_INVOICES":
		result = e.executeListSupplierInvoices(ctx, intent.Parameters)

	// ==========================================
	//  Stock / Inventory Module
	// ==========================================
	case "QUERY_STOCK", "LIST_INVENTORY":
		result = e.executeQueryStock(ctx, intent.Parameters, resolvedEntities)
	case "LIST_STOCK_MOVEMENTS":
		result = e.executeSimpleListByTable(ctx, "stock_movements", "stock_movement", "stock_movements", []string{"reference_number", "reference_type", "movement_type", "note"}, intent.Parameters)
	case "LIST_STOCK_OPNAME", "CREATE_STOCK_OPNAME":
		if intent.IntentCode == "LIST_STOCK_OPNAME" {
			result = e.executeSimpleListByTable(ctx, "stock_opnames", "stock_opname", "stock_opnames", []string{"opname_number", "status", "notes"}, intent.Parameters)
		} else {
			result = e.executeSimpleCreateByTable(ctx, "stock_opnames", "stock_opname", []string{"opname_number", "warehouse_id", "opname_date", "status", "notes", "created_by"}, []string{"warehouse_id", "opname_date"}, intent.Parameters)
		}

	// ==========================================
	//  Finance Module
	// ==========================================
	case "LIST_COA", "QUERY_COA":
		result = e.executeListCOA(ctx, intent.Parameters)
	case "LIST_JOURNALS":
		result = e.executeListJournals(ctx, intent.Parameters)
	case "CREATE_JOURNAL":
		result = e.executeCreateJournal(ctx, intent.Parameters)
	case "LIST_BANK_ACCOUNTS":
		result = e.executeListBankAccounts(ctx, intent.Parameters)
	case "LIST_PAYMENTS":
		result = e.executeListFinancePayments(ctx, intent.Parameters)
	case "LIST_TAX_INVOICES":
		result = e.executeListTaxInvoices(ctx, intent.Parameters)
	case "LIST_BUDGETS":
		result = e.executeListBudgets(ctx, intent.Parameters)
	case "LIST_CASH_BANK":
		result = e.executeListCashBank(ctx, intent.Parameters)
	case "LIST_ASSETS":
		result = e.executeListAssets(ctx, intent.Parameters)
	case "LIST_SALARY":
		result = e.executeListSalary(ctx, intent.Parameters)

	// ==========================================
	//  Master Data Module
	// ==========================================
	case "LIST_SUPPLIERS", "QUERY_SUPPLIER":
		if intent.IntentCode == "LIST_SUPPLIERS" {
			result = e.executeSimpleListByTable(ctx, "suppliers", "supplier", "suppliers", []string{"name", "code"}, intent.Parameters)
		} else {
			result = e.executeSimpleQueryByTable(ctx, "suppliers", "supplier", []string{"name", "code"}, intent.Parameters)
		}
	case "LIST_PRODUCTS", "QUERY_PRODUCT":
		// Smart reroute: product queries about stock levels → QUERY_STOCK
		if lowStock, ok := intent.Parameters["low_stock"].(bool); ok && lowStock {
			result = e.executeQueryStock(ctx, intent.Parameters, resolvedEntities)
		} else if search := getStringParam(intent.Parameters, "search"); search != "" && isStockFilterTerm(search) {
			intent.Parameters["low_stock"] = true
			result = e.executeQueryStock(ctx, intent.Parameters, resolvedEntities)
		} else {
			if intent.IntentCode == "LIST_PRODUCTS" {
				result = e.executeSimpleListByTable(ctx, "products", "product", "products", []string{"name", "code", "sku"}, intent.Parameters)
			} else {
				result = e.executeSimpleQueryByTable(ctx, "products", "product", []string{"name", "code", "sku"}, intent.Parameters)
			}
		}
	case "CREATE_PRODUCT":
		result = e.executeSimpleCreateByTable(ctx, "products", "product", []string{"code", "sku", "name", "description", "base_unit_id", "purchase_unit_id", "sales_unit_id", "product_category_id", "product_brand_id", "minimum_stock", "maximum_stock", "price", "sale_price", "purchase_price", "is_active", "created_by"}, []string{"name"}, intent.Parameters)
	case "LIST_WAREHOUSES":
		result = e.executeSimpleListByTable(ctx, "warehouses", "warehouse", "warehouses", []string{"name", "code"}, intent.Parameters)
	case "LIST_PRODUCT_CATEGORIES":
		result = e.executeSimpleListByTable(ctx, "product_categories", "product_category", "categories", []string{"name", "code"}, intent.Parameters)
	case "LIST_PRODUCT_BRANDS":
		result = e.executeSimpleListByTable(ctx, "product_brands", "product_brand", "brands", []string{"name", "code"}, intent.Parameters)
	case "LIST_PAYMENT_TERMS":
		result = e.executeSimpleListByTable(ctx, "payment_terms", "payment_term", "payment_terms", []string{"name", "code"}, intent.Parameters)
	case "LIST_COURIER_AGENCIES":
		result = e.executeSimpleListByTable(ctx, "courier_agencies", "courier_agency", "courier_agencies", []string{"name", "code"}, intent.Parameters)

	// ==========================================
	//  Organization Module
	// ==========================================
	case "LIST_DIVISIONS":
		result = e.executeSimpleListByTable(ctx, "divisions", "division", "divisions", []string{"name", "code"}, intent.Parameters)
	case "LIST_JOB_POSITIONS":
		result = e.executeSimpleListByTable(ctx, "job_positions", "job_position", "job_positions", []string{"name", "code"}, intent.Parameters)
	case "LIST_BUSINESS_UNITS":
		result = e.executeSimpleListByTable(ctx, "business_units", "business_unit", "business_units", []string{"name", "code"}, intent.Parameters)
	case "LIST_AREAS":
		result = e.executeSimpleListByTable(ctx, "areas", "area", "areas", []string{"name", "code"}, intent.Parameters)

	// ==========================================
	//  Geographic Module
	// ==========================================
	case "LIST_PROVINCES":
		result = e.executeSimpleListByTable(ctx, "provinces", "province", "provinces", []string{"name", "code"}, intent.Parameters)
	case "LIST_CITIES":
		result = e.executeSimpleListByTable(ctx, "cities", "city", "cities", []string{"name", "code"}, intent.Parameters)
	case "LIST_DISTRICTS":
		result = e.executeSimpleListByTable(ctx, "districts", "district", "districts", []string{"name", "code"}, intent.Parameters)

	// ==========================================
	//  Reports Module
	// ==========================================
	case "GENERATE_REPORT":
		result = e.executeGenerateReport(ctx, intent.Parameters)

	// ==========================================
	//  User Management Module
	// ==========================================
	case "LIST_USERS":
		result = e.executeSimpleListByTable(ctx, "users", "user", "users", []string{"name", "email"}, intent.Parameters)
	case "LIST_ROLES":
		result = e.executeSimpleListByTable(ctx, "roles", "role", "roles", []string{"name", "code"}, intent.Parameters)

	// ==========================================
	//  General
	// ==========================================
	case "GENERAL_CHAT":
		result = &ActionResult{
			Success: true,
			Message: "general_chat",
			Action:  "QUERY",
		}

	default:
		result = &ActionResult{
			Success:      false,
			Message:      fmt.Sprintf("Unknown intent code: %s", intent.IntentCode),
			Action:       intent.ActionType,
			ErrorCode:    "UNKNOWN_INTENT",
			ErrorMessage: fmt.Sprintf("Intent '%s' is not recognized. Please try a different request.", intent.IntentCode),
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result
}

// notImplementedResult returns a structured result for intents that are recognized but not yet wired to backend usecases
func (e *ActionExecutor) notImplementedResult(intent *IntentResult, guidance string) *ActionResult {
	return &ActionResult{
		Success:    true,
		Message:    guidance,
		Action:     intent.ActionType,
		EntityType: intent.Module,
		ErrorCode:  "NOT_IMPLEMENTED",
	}
}

// BuildActionPreview creates a human-readable preview of what will happen before execution
func (e *ActionExecutor) BuildActionPreview(intent *IntentResult, resolvedEntities map[string]*ResolvedEntity) map[string]interface{} {
	preview := map[string]interface{}{
		"intent":      intent.IntentCode,
		"action_type": intent.ActionType,
		"module":      intent.Module,
		"parameters":  intent.Parameters,
	}

	if len(resolvedEntities) > 0 {
		entities := make(map[string]string)
		for key, entity := range resolvedEntities {
			display := entity.DisplayName
			if entity.Code != "" {
				display = fmt.Sprintf("%s (%s)", entity.DisplayName, entity.Code)
			}
			entities[key] = display
		}
		preview["resolved_entities"] = entities
	}

	return preview
}

// --- HRD Action Implementations ---

func (e *ActionExecutor) executeCreateHoliday(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.HolidayUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Holiday service is not available"}
	}

	if shouldAutoCreateIndonesiaHolidays(params) {
		return e.executeCreateIndonesiaHolidayBatch(ctx, params)
	}

	req := &hrdDTO.CreateHolidayRequest{
		Name:     getStringParam(params, "name"),
		Date:     getStringParam(params, "date"),
		Type:     getStringParam(params, "type"),
		IsActive: true,
	}

	if req.Type == "" {
		req.Type = "NATIONAL"
	}

	if desc := getStringParam(params, "description"); desc != "" {
		req.Description = desc
	}

	if isCollective, ok := params["is_collective_leave"].(bool); ok {
		req.IsCollectiveLeave = isCollective
	}
	if cutsAnnual, ok := params["cuts_annual_leave"].(bool); ok {
		req.CutsAnnualLeave = cutsAnnual
	}

	resp, err := e.deps.HolidayUsecase.Create(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "CREATE",
			EntityType:   "holiday",
			ErrorCode:    "CREATE_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success:    true,
		Data:       resp,
		Message:    fmt.Sprintf("Holiday '%s' created successfully on %s", req.Name, req.Date),
		EntityType: "holiday",
		EntityID:   resp.ID,
		Action:     "CREATE",
	}
}

type indonesiaHolidayAPIItem struct {
	Date      string `json:"date"`
	LocalName string `json:"localName"`
	Name      string `json:"name"`
}

func shouldAutoCreateIndonesiaHolidays(params map[string]interface{}) bool {
	source := strings.ToUpper(strings.TrimSpace(getStringParam(params, "holiday_source")))
	countryCode := strings.ToUpper(strings.TrimSpace(getStringParam(params, "country_code")))
	if source == "PUBLIC_API" && countryCode == "ID" {
		return true
	}

	country := strings.ToLower(strings.TrimSpace(getStringParam(params, "country")))
	return country == "id" || country == "indonesia"
}

func (e *ActionExecutor) executeCreateIndonesiaHolidayBatch(ctx context.Context, params map[string]interface{}) *ActionResult {
	year := getIntParam(params, "year")
	if year < 2000 || year > 2100 {
		year = apptime.Now().Year()
	}

	apiItems, err := fetchIndonesiaPublicHolidays(ctx, year)
	if err != nil {
		return holidayCreateErrorResult(fmt.Sprintf("failed to fetch Indonesia holidays for %d: %v", year, err))
	}

	if len(apiItems) == 0 {
		return holidayCreateErrorResult(fmt.Sprintf("no Indonesia public holidays found for %d", year))
	}

	existing, err := e.deps.HolidayUsecase.GetByYear(ctx, year)
	if err != nil {
		return holidayCreateErrorResult(fmt.Sprintf("failed to check existing holidays: %v", err))
	}

	createReqs := buildIndonesiaHolidayCreateRequests(apiItems, existing)

	if len(createReqs) == 0 {
		return holidayNoopResult(year, len(apiItems))
	}

	created, err := e.deps.HolidayUsecase.CreateBatch(ctx, createReqs)
	if err != nil {
		return holidayCreateErrorResult(err.Error())
	}

	return &ActionResult{
		Success:    true,
		Action:     "CREATE",
		EntityType: "holiday",
		Message:    fmt.Sprintf("Created %d Indonesia holidays for year %d.", len(created), year),
		Data: map[string]interface{}{
			"year":          year,
			"created_count": len(created),
			"created":       created,
		},
	}
}

func holidayCreateErrorResult(errMsg string) *ActionResult {
	return &ActionResult{
		Success:      false,
		Action:       "CREATE",
		EntityType:   "holiday",
		ErrorCode:    "CREATE_FAILED",
		ErrorMessage: errMsg,
	}
}

func holidayNoopResult(year int, sourceCount int) *ActionResult {
	return &ActionResult{
		Success:    true,
		Action:     "CREATE",
		EntityType: "holiday",
		Message:    fmt.Sprintf("No new holidays to create. All Indonesia public holidays for %d already exist.", year),
		Data: map[string]interface{}{
			"year":          year,
			"created_count": 0,
			"skipped_count": sourceCount,
		},
	}
}

func buildIndonesiaHolidayCreateRequests(apiItems []indonesiaHolidayAPIItem, existing []hrdDTO.HolidayResponse) []hrdDTO.CreateHolidayRequest {
	existingByDate := make(map[string]struct{}, len(existing))
	for _, h := range existing {
		existingByDate[h.Date] = struct{}{}
	}

	createReqs := make([]hrdDTO.CreateHolidayRequest, 0, len(apiItems))
	for _, item := range apiItems {
		if req, ok := buildHolidayCreateRequestFromAPIItem(item, existingByDate); ok {
			createReqs = append(createReqs, req)
		}
	}

	return createReqs
}

func buildHolidayCreateRequestFromAPIItem(item indonesiaHolidayAPIItem, existingByDate map[string]struct{}) (hrdDTO.CreateHolidayRequest, bool) {
	date := strings.TrimSpace(item.Date)
	if date == "" {
		return hrdDTO.CreateHolidayRequest{}, false
	}
	if _, exists := existingByDate[date]; exists {
		return hrdDTO.CreateHolidayRequest{}, false
	}

	name := strings.TrimSpace(item.LocalName)
	if name == "" {
		name = strings.TrimSpace(item.Name)
	}
	if name == "" {
		return hrdDTO.CreateHolidayRequest{}, false
	}

	description := "Imported from Indonesia public holiday dataset"
	if item.Name != "" && item.Name != name {
		description = fmt.Sprintf("%s (%s)", description, item.Name)
	}

	return hrdDTO.CreateHolidayRequest{
		Date:              date,
		Name:              name,
		Description:       description,
		Type:              "NATIONAL",
		IsCollectiveLeave: false,
		CutsAnnualLeave:   false,
		IsRecurring:       false,
		IsActive:          true,
	}, true
}

func fetchIndonesiaPublicHolidays(ctx context.Context, year int) ([]indonesiaHolidayAPIItem, error) {
	url := fmt.Sprintf("https://date.nager.at/api/v3/PublicHolidays/%d/ID", year)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("holiday source API returned status %d", resp.StatusCode)
	}

	var out []indonesiaHolidayAPIItem
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (e *ActionExecutor) executeListHolidays(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.HolidayUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Holiday service is not available"}
	}

	req := &hrdDTO.ListHolidaysRequest{
		Page:    1,
		PerPage: 20,
	}

	if year := getIntParam(params, "year"); year > 0 {
		req.Year = year
	}
	if search := getStringParam(params, "search"); search != "" {
		req.Search = search
	}
	if hType := getStringParam(params, "type"); hType != "" {
		req.Type = hType
	}

	holidays, pagination, err := e.deps.HolidayUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "holiday",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"holidays":   holidays,
			"pagination": pagination,
		},
		Message:    fmt.Sprintf("Found %d holidays", len(holidays)),
		EntityType: "holiday",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeCreateLeaveRequest(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.LeaveRequestUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: leaveRequestServiceUnavailableMessage}
	}

	// Convert params to JSON then unmarshal to the DTO for flexibility
	paramJSON, err := json.Marshal(params)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: invalidParamsMessage}
	}

	var req hrdDTO.CreateLeaveRequestDTO
	if err := json.Unmarshal(paramJSON, &req); err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: "Failed to parse leave request parameters"}
	}

	resp, err := e.deps.LeaveRequestUsecase.Create(ctx, &req, currentUserID)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "CREATE",
			EntityType:   "leave_request",
			ErrorCode:    "CREATE_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success:    true,
		Data:       resp,
		Message:    "Leave request created successfully",
		EntityType: "leave_request",
		EntityID:   resp.ID,
		Action:     "CREATE",
	}
}

func (e *ActionExecutor) executeListLeaveRequests(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.LeaveRequestUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: leaveRequestServiceUnavailableMessage}
	}

	filters := &hrdDTO.LeaveRequestListFilterDTO{
		Page:    1,
		PerPage: 20,
	}

	if status := getStringParam(params, "status"); status != "" {
		filters.Status = &status
	}

	results, total, err := e.deps.LeaveRequestUsecase.List(ctx, filters, currentUserID)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "leave_request",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"leave_requests": results,
			"total":          total,
		},
		Message:    fmt.Sprintf("Found %d leave requests", total),
		EntityType: "leave_request",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeApproveLeaveRequest(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.LeaveRequestUsecase == nil {
		return &ActionResult{Success: false, Action: "UPDATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: leaveRequestServiceUnavailableMessage}
	}

	leaveRequestID := strings.TrimSpace(getStringParam(params, "leave_request_id"))
	if leaveRequestID == "" {
		leaveRequestID = strings.TrimSpace(getStringParam(params, "id"))
	}
	if leaveRequestID == "" {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "leave_request", ErrorCode: "MISSING_PARAMS", ErrorMessage: "Please provide leave_request_id or id"}
	}

	resp, err := e.deps.LeaveRequestUsecase.Approve(ctx, leaveRequestID, &hrdDTO.ApproveLeaveRequestDTO{}, currentUserID)
	if err != nil {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "leave_request", ErrorCode: "UPDATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{
		Success:    true,
		Data:       resp,
		Message:    "Leave request approved successfully",
		EntityType: "leave_request",
		EntityID:   resp.ID,
		Action:     "UPDATE",
	}
}

func (e *ActionExecutor) executeRejectLeaveRequest(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.LeaveRequestUsecase == nil {
		return &ActionResult{Success: false, Action: "UPDATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: leaveRequestServiceUnavailableMessage}
	}

	leaveRequestID := strings.TrimSpace(getStringParam(params, "leave_request_id"))
	if leaveRequestID == "" {
		leaveRequestID = strings.TrimSpace(getStringParam(params, "id"))
	}
	if leaveRequestID == "" {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "leave_request", ErrorCode: "MISSING_PARAMS", ErrorMessage: "Please provide leave_request_id or id"}
	}

	rejectionNote := strings.TrimSpace(getStringParam(params, "rejection_note"))
	if rejectionNote == "" {
		rejectionNote = strings.TrimSpace(getStringParam(params, "reason"))
	}
	if len(rejectionNote) < 10 {
		rejectionNote = "Rejected by AI assistant due to policy or manager decision."
	}

	resp, err := e.deps.LeaveRequestUsecase.Reject(ctx, leaveRequestID, &hrdDTO.RejectLeaveRequestDTO{RejectionNote: rejectionNote}, currentUserID)
	if err != nil {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "leave_request", ErrorCode: "UPDATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{
		Success:    true,
		Data:       resp,
		Message:    "Leave request rejected successfully",
		EntityType: "leave_request",
		EntityID:   resp.ID,
		Action:     "UPDATE",
	}
}

func (e *ActionExecutor) executeQueryAttendance(ctx context.Context, params map[string]interface{}, resolvedEntities map[string]*ResolvedEntity) *ActionResult {
	if e.deps.AttendanceUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Attendance service is not available"}
	}

	req := &hrdDTO.ListAttendanceRecordsRequest{
		Page:    1,
		PerPage: 20,
	}

	if emp, ok := resolvedEntities["employee"]; ok {
		req.EmployeeID = emp.ID
	}

	records, pagination, err := e.deps.AttendanceUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "attendance",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"records":    records,
			"pagination": pagination,
		},
		Message:    fmt.Sprintf("Found %d attendance records", len(records)),
		EntityType: "attendance",
		Action:     "QUERY",
	}
}

// --- Sales Action Implementations ---

func (e *ActionExecutor) executeCreateSalesQuotation(ctx context.Context, params map[string]interface{}, currentUserID string, resolvedEntities map[string]*ResolvedEntity) *ActionResult {
	if e.deps.SalesQuotationUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: salesQuotationServiceUnavailableMessage}
	}

	req, errResult := e.buildSalesQuotationRequest(ctx, params, currentUserID, resolvedEntities)
	if errResult != nil {
		return errResult
	}

	resp, err := e.deps.SalesQuotationUsecase.Create(ctx, req, &currentUserID)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "CREATE",
			EntityType:   "sales_quotation",
			ErrorCode:    "CREATE_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success:    true,
		Data:       resp,
		Message:    fmt.Sprintf("Sales quotation created for %s", req.CustomerName),
		EntityType: "sales_quotation",
		EntityID:   resp.ID,
		Action:     "CREATE",
	}
}

func (e *ActionExecutor) buildSalesQuotationRequest(ctx context.Context, params map[string]interface{}, currentUserID string, resolvedEntities map[string]*ResolvedEntity) (*salesDTO.CreateSalesQuotationRequest, *ActionResult) {
	if errResult := e.resolveSalesQuotationReferences(ctx, params); errResult != nil {
		return nil, errResult
	}

	applySalesQuotationDefaults(ctx, e.entityResolver, params, currentUserID, resolvedEntities)
	cleanParams := filterMapByAllowedKeys(params, map[string]bool{
		"quotation_date":   true,
		"valid_until":      true,
		"payment_terms_id": true,
		"sales_rep_id":     true,
		"business_unit_id": true,
		"business_type_id": true,
		"customer_name":    true,
		"customer_contact": true,
		"customer_phone":   true,
		"customer_email":   true,
		"tax_rate":         true,
		"delivery_cost":    true,
		"other_cost":       true,
		"discount_amount":  true,
		"notes":            true,
		"items":            true,
	})

	var req salesDTO.CreateSalesQuotationRequest
	if errResult := marshalParamsToRequest(cleanParams, &req, "sales quotation"); errResult != nil {
		return nil, errResult
	}

	return &req, nil
}

func (e *ActionExecutor) resolveSalesQuotationReferences(ctx context.Context, params map[string]interface{}) *ActionResult {
	if _, hasID := params["payment_terms_id"]; !hasID {
		for _, key := range []string{"payment_terms_name", "payment_terms", "syarat_pembayaran"} {
			name := strings.TrimSpace(getStringParam(params, key))
			if name == "" {
				continue
			}
			ptID, err := e.entityResolver.ResolvePaymentTerms(ctx, name)
			if err != nil {
				return newEntityNotFoundResult("sales_quotation", fmt.Sprintf("Syarat pembayaran '%s' tidak ditemukan di database", name))
			}
			params["payment_terms_id"] = ptID
			break
		}
	}

	if _, hasID := params["business_unit_id"]; !hasID {
		for _, key := range []string{"business_unit_name", "business_unit", "unit_bisnis"} {
			name := strings.TrimSpace(getStringParam(params, key))
			if name == "" {
				continue
			}
			buID, err := e.entityResolver.ResolveBusinessUnit(ctx, name)
			if err != nil {
				return newEntityNotFoundResult("sales_quotation", fmt.Sprintf("Unit bisnis '%s' tidak ditemukan di database", name))
			}
			params["business_unit_id"] = buID
			break
		}
	}

	if items, ok := params["items"].([]interface{}); ok {
		if errResult := e.normalizeSalesQuotationItems(ctx, items); errResult != nil {
			return errResult
		}
		params["items"] = items
	}

	return nil
}

func (e *ActionExecutor) normalizeSalesQuotationItems(ctx context.Context, items []interface{}) *ActionResult {
	for i, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if price := getFloatParam(itemMap, "price"); price > 0 {
			itemMap["price"] = price
		}
		if discount := getFloatParam(itemMap, "discount"); discount > 0 {
			itemMap["discount"] = discount
		}
		if qty := getFloatParam(itemMap, "quantity"); qty > 0 {
			itemMap["quantity"] = qty
		}

		if _, hasProductID := itemMap["product_id"]; !hasProductID {
			productName := strings.TrimSpace(getStringParam(itemMap, "product_name"))
			if productName != "" {
				prodID, prodPrice, err := e.entityResolver.ResolveProductByName(ctx, productName)
				if err != nil {
					return newEntityNotFoundResult("sales_quotation", fmt.Sprintf("Produk '%s' tidak dapat di-resolve: %s", productName, err.Error()))
				}
				itemMap["product_id"] = prodID
				if price, _ := itemMap["price"].(float64); price <= 0 && prodPrice > 0 {
					itemMap["price"] = prodPrice
				}
			}
		}
		if qty, _ := itemMap["quantity"].(float64); qty <= 0 {
			itemMap["quantity"] = float64(1)
		}
		if _, hasDiscount := itemMap["discount"]; !hasDiscount {
			itemMap["discount"] = float64(0)
		}
		items[i] = itemMap
	}
	return nil
}

func applySalesQuotationDefaults(ctx context.Context, resolver *EntityResolver, params map[string]interface{}, currentUserID string, resolvedEntities map[string]*ResolvedEntity) {
	applyQuotationDateDefault(params)
	applySalesRepDefault(ctx, resolver, params, currentUserID)
	applyResolvedCustomerName(params, resolvedEntities)
}

func applyQuotationDateDefault(params map[string]interface{}) {
	today := apptime.Now().Format("2006-01-02")
	qd, ok := params["quotation_date"].(string)
	if !ok || strings.TrimSpace(qd) == "" {
		params["quotation_date"] = today
		return
	}
	parsed, err := time.Parse("2006-01-02", qd)
	if err != nil || parsed.Year() < apptime.Now().Year() {
		params["quotation_date"] = today
	}
}

func applySalesRepDefault(ctx context.Context, resolver *EntityResolver, params map[string]interface{}, currentUserID string) {
	if strings.TrimSpace(getStringParam(params, "sales_rep_id")) != "" {
		return
	}
	empID, err := resolver.ResolveUserToEmployeeID(ctx, currentUserID)
	if err != nil {
		return
	}
	params["sales_rep_id"] = empID
}

func applyResolvedCustomerName(params map[string]interface{}, resolvedEntities map[string]*ResolvedEntity) {
	if customer, ok := resolvedEntities["customer"]; ok && strings.TrimSpace(getStringParam(params, "customer_name")) == "" {
		params["customer_name"] = customer.DisplayName
	}
}

func (e *ActionExecutor) executeListSalesQuotations(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.SalesQuotationUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: salesQuotationServiceUnavailableMessage}
	}

	req := &salesDTO.ListSalesQuotationsRequest{
		Page:    1,
		PerPage: 20,
	}

	if search := getStringParam(params, "search"); search != "" {
		req.Search = search
	}
	if status := getStringParam(params, "status"); status != "" {
		req.Status = status
	}
	if period := getStringParam(params, "period"); period != "" {
		now := apptime.Now()
		switch period {
		case "current_month":
			req.DateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		case "last_month":
			lastMonth := now.AddDate(0, -1, 0)
			req.DateFrom = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		case "current_year":
			req.DateFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		}
	}

	quotations, pagination, err := e.deps.SalesQuotationUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "sales_quotation",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"quotations": quotations,
			"pagination": pagination,
		},
		Message:    fmt.Sprintf("Found %d sales quotations", len(quotations)),
		EntityType: "sales_quotation",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeApproveSalesQuotation(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.SalesQuotationUsecase == nil {
		return &ActionResult{Success: false, Action: "UPDATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: salesQuotationServiceUnavailableMessage}
	}

	quotationID := strings.TrimSpace(getStringParam(params, "quotation_id"))
	if quotationID == "" {
		quotationID = strings.TrimSpace(getStringParam(params, "id"))
	}
	if quotationID == "" {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "sales_quotation", ErrorCode: "MISSING_PARAMS", ErrorMessage: "Please provide quotation_id or id"}
	}

	status := strings.ToLower(strings.TrimSpace(getStringParam(params, "status")))
	if status == "" {
		status = "approved"
	}

	req := &salesDTO.UpdateSalesQuotationStatusRequest{Status: status}
	if status == "rejected" {
		reason := strings.TrimSpace(getStringParam(params, "rejection_reason"))
		if reason == "" {
			reason = strings.TrimSpace(getStringParam(params, "reason"))
		}
		if reason != "" {
			req.RejectionReason = &reason
		}
	}

	resp, err := e.deps.SalesQuotationUsecase.UpdateStatus(ctx, quotationID, req, &currentUserID)
	if err != nil {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "sales_quotation", ErrorCode: "UPDATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{
		Success:    true,
		Data:       resp,
		Message:    fmt.Sprintf("Sales quotation %s updated to %s", resp.Code, strings.ToUpper(status)),
		EntityType: "sales_quotation",
		EntityID:   resp.ID,
		Action:     "UPDATE",
	}
}

func (e *ActionExecutor) executeCreateSalesOrder(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.SalesOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: salesOrderServiceUnavailableMessage}
	}

	req, errResult := e.buildSalesOrderRequest(ctx, params, currentUserID)
	if errResult != nil {
		return errResult
	}

	resp, err := e.deps.SalesOrderUsecase.Create(ctx, req, &currentUserID)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: "sales_order", ErrorCode: "CREATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{Success: true, Data: resp, Message: fmt.Sprintf("Sales order %s created successfully", resp.Code), EntityType: "sales_order", EntityID: resp.ID, Action: "CREATE"}
}

func (e *ActionExecutor) buildSalesOrderRequest(ctx context.Context, params map[string]interface{}, currentUserID string) (*salesDTO.CreateSalesOrderRequest, *ActionResult) {
	applySalesOrderDefaults(ctx, e.entityResolver, params, currentUserID)

	var req salesDTO.CreateSalesOrderRequest
	if errResult := marshalParamsToRequest(params, &req, "sales order"); errResult != nil {
		return nil, errResult
	}

	return &req, nil
}

func applySalesOrderDefaults(ctx context.Context, resolver *EntityResolver, params map[string]interface{}, currentUserID string) {
	resolveCustomerIdentity(ctx, resolver, params)
	resolveSalesRepID(ctx, resolver, params, currentUserID)
	resolvePaymentTermsID(ctx, resolver, params)
	resolveBusinessUnitID(ctx, resolver, params)
	normalizeSalesOrderItems(ctx, resolver, params)
}

func resolveSalesRepID(ctx context.Context, resolver *EntityResolver, params map[string]interface{}, currentUserID string) {
	if strings.TrimSpace(getStringParam(params, "sales_rep_id")) != "" {
		return
	}

	for _, key := range []string{"sales_rep_name", "sales_rep", "sales_name", "employee_name"} {
		salesRepName := strings.TrimSpace(getStringParam(params, key))
		if salesRepName == "" {
			continue
		}

		entity, err := resolver.ResolveEmployee(ctx, salesRepName)
		if err == nil && entity != nil && strings.TrimSpace(entity.ID) != "" {
			params["sales_rep_id"] = entity.ID
			if strings.TrimSpace(getStringParam(params, "sales_rep_name")) == "" && strings.TrimSpace(entity.DisplayName) != "" {
				params["sales_rep_name"] = entity.DisplayName
			}
		}
		return
	}

	if strings.TrimSpace(currentUserID) == "" {
		return
	}
	empID, err := resolver.ResolveUserToEmployeeID(ctx, currentUserID)
	if err == nil && strings.TrimSpace(empID) != "" {
		params["sales_rep_id"] = empID
	}
}

func resolveCustomerIdentity(ctx context.Context, resolver *EntityResolver, params map[string]interface{}) {
	if strings.TrimSpace(getStringParam(params, "customer_id")) != "" {
		return
	}

	customerName := strings.TrimSpace(getStringParam(params, "customer_name"))
	if customerName == "" {
		return
	}

	entity, err := resolver.ResolveCustomer(ctx, customerName)
	if err != nil || entity == nil {
		return
	}

	if strings.TrimSpace(entity.ID) != "" {
		params["customer_id"] = entity.ID
	}
	if strings.TrimSpace(entity.DisplayName) != "" {
		params["customer_name"] = entity.DisplayName
	}
}

func resolvePaymentTermsID(ctx context.Context, resolver *EntityResolver, params map[string]interface{}) {
	if _, hasID := params["payment_terms_id"]; hasID {
		return
	}
	for _, key := range []string{"payment_terms_name", "payment_terms", "syarat_pembayaran"} {
		name := strings.TrimSpace(getStringParam(params, key))
		if name == "" {
			continue
		}
		ptID, err := resolver.ResolvePaymentTerms(ctx, name)
		if err == nil {
			params["payment_terms_id"] = ptID
		}
		return
	}
}

func resolveBusinessUnitID(ctx context.Context, resolver *EntityResolver, params map[string]interface{}) {
	if _, hasID := params["business_unit_id"]; hasID {
		return
	}
	for _, key := range []string{"business_unit_name", "business_unit", "unit_bisnis"} {
		name := strings.TrimSpace(getStringParam(params, key))
		if name == "" {
			continue
		}
		buID, err := resolver.ResolveBusinessUnit(ctx, name)
		if err == nil {
			params["business_unit_id"] = buID
		}
		return
	}

	buID, err := resolver.ResolveDefaultBusinessUnit(ctx)
	if err == nil && strings.TrimSpace(buID) != "" {
		params["business_unit_id"] = buID
	}
}

func normalizeSalesOrderItems(ctx context.Context, resolver *EntityResolver, params map[string]interface{}) {
	rawItems := params["items"]
	if rawItems == nil {
		return
	}

	// Sanitize: the AI may emit items as a JSON-encoded string instead of an array.
	// e.g. "[{\"product_name\":\"X\",\"quantity\":3,\"price\":50000}]"
	// Parse and replace so downstream JSON→struct unmarshal succeeds.
	if itemsStr, ok := rawItems.(string); ok {
		trimmed := strings.TrimSpace(itemsStr)
		var parsed []interface{}
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			params["items"] = parsed
			rawItems = parsed
		} else {
			// Malformed string — remove to prevent unmarshal panic downstream.
			delete(params, "items")
			return
		}
	}

	items, ok := rawItems.([]interface{})
	if !ok {
		return
	}
	for i, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if price := getFloatParam(itemMap, "price"); price > 0 {
			itemMap["price"] = price
		}
		if discount := getFloatParam(itemMap, "discount"); discount > 0 {
			itemMap["discount"] = discount
		}
		if qty := getFloatParam(itemMap, "quantity"); qty > 0 {
			itemMap["quantity"] = qty
		}

		if _, hasProductID := itemMap["product_id"]; !hasProductID {
			productName := strings.TrimSpace(getStringParam(itemMap, "product_name"))
			if productName != "" {
				prodID, _, err := resolver.ResolveProductByName(ctx, productName)
				if err == nil {
					itemMap["product_id"] = prodID
				}
			}
		}
		if _, hasDiscount := itemMap["discount"]; !hasDiscount {
			itemMap["discount"] = float64(0)
		}
		items[i] = itemMap
	}
	params["items"] = items
}

func (e *ActionExecutor) executeCreateDeliveryOrder(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.DeliveryOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Delivery order service is not available"}
	}

	if strings.TrimSpace(getStringParam(params, "delivery_date")) == "" {
		params["delivery_date"] = apptime.Now().Format("2006-01-02")
	}

	paramJSON, err := json.Marshal(params)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: invalidParamsMessage}
	}

	var req salesDTO.CreateDeliveryOrderRequest
	if err := json.Unmarshal(paramJSON, &req); err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: fmt.Sprintf("Failed to parse delivery order parameters: %s", err.Error())}
	}

	resp, err := e.deps.DeliveryOrderUsecase.Create(ctx, &req, &currentUserID)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: "delivery_order", ErrorCode: "CREATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{Success: true, Data: resp, Message: fmt.Sprintf("Delivery order %s created successfully", resp.Code), EntityType: "delivery_order", EntityID: resp.ID, Action: "CREATE"}
}

func (e *ActionExecutor) executeCreateSalesInvoice(ctx context.Context, params map[string]interface{}, currentUserID string) *ActionResult {
	if e.deps.CustomerInvoiceUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Customer invoice service is not available"}
	}

	if strings.TrimSpace(getStringParam(params, "invoice_date")) == "" {
		params["invoice_date"] = apptime.Now().Format("2006-01-02")
	}

	paramJSON, err := json.Marshal(params)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: invalidParamsMessage}
	}

	var req salesDTO.CreateCustomerInvoiceRequest
	if err := json.Unmarshal(paramJSON, &req); err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: fmt.Sprintf("Failed to parse sales invoice parameters: %s", err.Error())}
	}

	resp, err := e.deps.CustomerInvoiceUsecase.Create(ctx, &req, &currentUserID)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: "customer_invoice", ErrorCode: "CREATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{Success: true, Data: resp, Message: fmt.Sprintf("Sales invoice %s created successfully", resp.Code), EntityType: "customer_invoice", EntityID: resp.ID, Action: "CREATE"}
}

func (e *ActionExecutor) executeListSalesOrders(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.SalesOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: salesOrderServiceUnavailableMessage}
	}

	req := &salesDTO.ListSalesOrdersRequest{
		Page:    1,
		PerPage: 20,
	}

	if search := getStringParam(params, "search"); search != "" {
		req.Search = search
	} else if orderNumber := getStringParam(params, "order_number"); orderNumber != "" {
		req.Search = orderNumber
	} else if customerName := getStringParam(params, "customer_name"); customerName != "" {
		req.Search = customerName
	}

	if status := getStringParam(params, "status"); status != "" {
		req.Status = status
	}

	if period := getStringParam(params, "period"); period != "" {
		now := apptime.Now()
		switch period {
		case "current_month":
			req.DateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		case "last_month":
			lastMonth := now.AddDate(0, -1, 0)
			req.DateFrom = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		case "current_year":
			req.DateFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		}
	}

	if page := getIntParam(params, "page"); page > 0 {
		req.Page = page
	}
	if perPage := getIntParam(params, "per_page"); perPage > 0 {
		req.PerPage = perPage
	}
	if unfulfilledOnly, ok := params["unfulfilled_only"].(bool); ok {
		req.UnfulfilledOnly = unfulfilledOnly
	}

	orders, pagination, err := e.deps.SalesOrderUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "sales_order",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"orders":     orders,
			"pagination": pagination,
		},
		Message:    fmt.Sprintf("Found %d sales orders", len(orders)),
		EntityType: "sales_order",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeQuerySalesOrder(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.SalesOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: salesOrderServiceUnavailableMessage}
	}

	if orderID := getStringParam(params, "order_id"); orderID != "" {
		order, err := e.deps.SalesOrderUsecase.GetByID(ctx, orderID)
		if err != nil {
			return &ActionResult{
				Success:      false,
				Action:       "QUERY",
				EntityType:   "sales_order",
				ErrorCode:    "QUERY_FAILED",
				ErrorMessage: err.Error(),
			}
		}

		return &ActionResult{
			Success:    true,
			Data:       order,
			Message:    fmt.Sprintf("Sales order %s found", order.Code),
			EntityType: "sales_order",
			EntityID:   order.ID,
			Action:     "QUERY",
		}
	}

	// Fallback for natural language queries that mention order number/customer
	listParams := map[string]interface{}{}
	if orderNumber := getStringParam(params, "order_number"); orderNumber != "" {
		listParams["search"] = orderNumber
	}
	if search := getStringParam(params, "search"); search != "" {
		listParams["search"] = search
	}
	if customerName := getStringParam(params, "customer_name"); customerName != "" {
		if _, exists := listParams["search"]; !exists {
			listParams["search"] = customerName
		}
	}
	if status := getStringParam(params, "status"); status != "" {
		listParams["status"] = status
	}
	if period := getStringParam(params, "period"); period != "" {
		listParams["period"] = period
	}
	if len(listParams) == 0 {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "sales_order",
			ErrorCode:    "MISSING_PARAMS",
			ErrorMessage: "Please specify order_id, order_number, search, or customer_name",
		}
	}

	result := e.executeListSalesOrders(ctx, listParams)
	if !result.Success {
		return result
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "sales_order",
			ErrorCode:    "INVALID_RESPONSE",
			ErrorMessage: "Unexpected sales order response format",
		}
	}

	ordersRaw, ok := data["orders"]
	if !ok {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "sales_order",
			ErrorCode:    "INVALID_RESPONSE",
			ErrorMessage: "Sales order list is missing in response",
		}
	}

	orders, ok := ordersRaw.([]salesDTO.SalesOrderResponse)
	if !ok {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "sales_order",
			ErrorCode:    "INVALID_RESPONSE",
			ErrorMessage: "Sales order list response has invalid type",
		}
	}

	if len(orders) == 0 {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "sales_order",
			ErrorCode:    "NOT_FOUND",
			ErrorMessage: "Sales order not found",
		}
	}

	order := orders[0]
	return &ActionResult{
		Success:    true,
		Data:       order,
		Message:    fmt.Sprintf("Sales order %s found", order.Code),
		EntityType: "sales_order",
		EntityID:   order.ID,
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeListPurchaseOrders(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.PurchaseOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: purchaseOrderServiceUnavailableMessage}
	}

	page := 1
	perPage := 20
	if p := getIntParam(params, "page"); p > 0 {
		page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		perPage = pp
	}

	listParams := purchaseRepos.PurchaseOrderListParams{
		Search: strings.TrimSpace(getStringParam(params, "search")),
		Status: strings.TrimSpace(getStringParam(params, "status")),
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	// Fallback for extractor schemas that may provide supplier_name or order_number.
	if listParams.Search == "" {
		if orderNumber := strings.TrimSpace(getStringParam(params, "order_number")); orderNumber != "" {
			listParams.Search = orderNumber
		} else if supplierName := strings.TrimSpace(getStringParam(params, "supplier_name")); supplierName != "" {
			listParams.Search = supplierName
		}
	}

	orders, total, err := e.deps.PurchaseOrderUsecase.List(ctx, listParams)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "purchase_order",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"orders": orders,
			"pagination": map[string]interface{}{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": totalPages,
			},
		},
		Message:    fmt.Sprintf("Found %d purchase orders", len(orders)),
		EntityType: "purchase_order",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeListDeliveryOrders(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.DeliveryOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Delivery order service is not available"}
	}

	req := &salesDTO.ListDeliveryOrdersRequest{
		Page:    1,
		PerPage: 20,
	}

	if search := getStringParam(params, "search"); search != "" {
		req.Search = search
	}
	if status := getStringParam(params, "status"); status != "" {
		req.Status = status
	}
	if page := getIntParam(params, "page"); page > 0 {
		req.Page = page
	}
	if perPage := getIntParam(params, "per_page"); perPage > 0 {
		req.PerPage = perPage
	}

	if period := getStringParam(params, "period"); period != "" {
		now := apptime.Now()
		switch period {
		case "current_month":
			req.DateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		case "last_month":
			lastMonth := now.AddDate(0, -1, 0)
			req.DateFrom = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		case "current_year":
			req.DateFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		}
	}

	items, pagination, err := e.deps.DeliveryOrderUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "delivery_order",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"delivery_orders": items,
			"pagination":      pagination,
		},
		Message:    fmt.Sprintf("Found %d delivery orders", len(items)),
		EntityType: "delivery_order",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeListSalesInvoices(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.CustomerInvoiceUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Customer invoice service is not available"}
	}

	req := &salesDTO.ListCustomerInvoicesRequest{
		Page:    1,
		PerPage: 20,
	}

	if search := getStringParam(params, "search"); search != "" {
		req.Search = search
	}
	if status := getStringParam(params, "status"); status != "" {
		req.Status = status
	}
	if page := getIntParam(params, "page"); page > 0 {
		req.Page = page
	}
	if perPage := getIntParam(params, "per_page"); perPage > 0 {
		req.PerPage = perPage
	}

	if period := getStringParam(params, "period"); period != "" {
		now := apptime.Now()
		switch period {
		case "current_month":
			req.DateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		case "last_month":
			lastMonth := now.AddDate(0, -1, 0)
			req.DateFrom = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		case "current_year":
			req.DateFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			req.DateTo = now.Format("2006-01-02")
		}
	}

	invoices, pagination, err := e.deps.CustomerInvoiceUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "customer_invoice",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"invoices":   invoices,
			"pagination": pagination,
		},
		Message:    fmt.Sprintf("Found %d sales invoices", len(invoices)),
		EntityType: "customer_invoice",
		Action:     "QUERY",
	}
}

func resolvePeriodDateRange(period string) (string, string) {
	now := apptime.Now()
	switch period {
	case "current_month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02"), now.Format("2006-01-02")
	case "last_month":
		lastMonth := now.AddDate(0, -1, 0)
		return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02"), time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	case "current_year":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02"), now.Format("2006-01-02")
	default:
		return "", ""
	}
}

func (e *ActionExecutor) executeListPurchaseRequisitions(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.PurchaseRequisitionUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Purchase requisition service is not available"}
	}

	page, perPage := 1, 20
	if p := getIntParam(params, "page"); p > 0 {
		page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		perPage = pp
	}

	req := purchaseRepos.PurchaseRequisitionListParams{
		Search: strings.TrimSpace(getStringParam(params, "search")),
		Status: strings.TrimSpace(getStringParam(params, "status")),
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	items, total, err := e.deps.PurchaseRequisitionUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "purchase_requisition", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"requisitions": items,
			"pagination": map[string]interface{}{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": totalPages,
			},
		},
		Message:    fmt.Sprintf("Found %d purchase requisitions", len(items)),
		EntityType: "purchase_requisition",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeListGoodsReceipts(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.GoodsReceiptUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Goods receipt service is not available"}
	}

	page, perPage := 1, 20
	if p := getIntParam(params, "page"); p > 0 {
		page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		perPage = pp
	}

	req := purchaseRepos.GoodsReceiptListParams{
		Search: strings.TrimSpace(getStringParam(params, "search")),
		Status: strings.TrimSpace(getStringParam(params, "status")),
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	items, total, err := e.deps.GoodsReceiptUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "goods_receipt", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"goods_receipts": items,
			"pagination": map[string]interface{}{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": totalPages,
			},
		},
		Message:    fmt.Sprintf("Found %d goods receipts", len(items)),
		EntityType: "goods_receipt",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeListSupplierInvoices(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.SupplierInvoiceUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Supplier invoice service is not available"}
	}

	page, perPage := 1, 20
	if p := getIntParam(params, "page"); p > 0 {
		page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		perPage = pp
	}

	req := purchaseRepos.SupplierInvoiceListParams{
		Search: strings.TrimSpace(getStringParam(params, "search")),
		Status: strings.TrimSpace(getStringParam(params, "status")),
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	items, total, err := e.deps.SupplierInvoiceUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "supplier_invoice", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"supplier_invoices": items,
			"pagination": map[string]interface{}{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": totalPages,
			},
		},
		Message:    fmt.Sprintf("Found %d supplier invoices", len(items)),
		EntityType: "supplier_invoice",
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeCreatePurchaseRequisition(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.PurchaseRequisitionUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Purchase requisition service is not available"}
	}

	if strings.TrimSpace(getStringParam(params, "request_date")) == "" {
		params["request_date"] = apptime.Now().Format("2006-01-02")
	}

	paramJSON, err := json.Marshal(params)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: invalidParamsMessage}
	}

	var req purchaseDTO.CreatePurchaseRequisitionRequest
	if err := json.Unmarshal(paramJSON, &req); err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: fmt.Sprintf("Failed to parse purchase requisition parameters: %s", err.Error())}
	}

	resp, err := e.deps.PurchaseRequisitionUsecase.Create(ctx, &req)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: "purchase_requisition", ErrorCode: "CREATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{Success: true, Data: resp, Message: fmt.Sprintf("Purchase requisition %s created successfully", resp.Code), EntityType: "purchase_requisition", EntityID: resp.ID, Action: "CREATE"}
}

func (e *ActionExecutor) executeCreatePurchaseOrder(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.PurchaseOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: purchaseOrderServiceUnavailableMessage}
	}

	if strings.TrimSpace(getStringParam(params, "order_date")) == "" {
		params["order_date"] = apptime.Now().Format("2006-01-02")
	}

	paramJSON, err := json.Marshal(params)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: invalidParamsMessage}
	}

	var req purchaseDTO.CreatePurchaseOrderRequest
	if err := json.Unmarshal(paramJSON, &req); err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: fmt.Sprintf("Failed to parse purchase order parameters: %s", err.Error())}
	}

	resp, err := e.deps.PurchaseOrderUsecase.Create(ctx, &req)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: "purchase_order", ErrorCode: "CREATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{Success: true, Data: resp, Message: fmt.Sprintf("Purchase order %s created successfully", resp.Code), EntityType: "purchase_order", EntityID: resp.ID, Action: "CREATE"}
}

func (e *ActionExecutor) executeApprovePurchaseOrder(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.PurchaseOrderUsecase == nil {
		return &ActionResult{Success: false, Action: "UPDATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: purchaseOrderServiceUnavailableMessage}
	}

	purchaseOrderID := strings.TrimSpace(getStringParam(params, "purchase_order_id"))
	if purchaseOrderID == "" {
		purchaseOrderID = strings.TrimSpace(getStringParam(params, "id"))
	}
	if purchaseOrderID == "" {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "purchase_order", ErrorCode: "MISSING_PARAMS", ErrorMessage: "Please provide purchase_order_id or id"}
	}

	resp, err := e.deps.PurchaseOrderUsecase.Approve(ctx, purchaseOrderID)
	if err != nil {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: "purchase_order", ErrorCode: "UPDATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{Success: true, Data: resp, Message: fmt.Sprintf("Purchase order %s approved successfully", resp.Code), EntityType: "purchase_order", EntityID: resp.ID, Action: "UPDATE"}
}

func (e *ActionExecutor) executeCreateJournal(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.JournalUsecase == nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Journal service is not available"}
	}

	if strings.TrimSpace(getStringParam(params, "entry_date")) == "" {
		params["entry_date"] = apptime.Now().Format("2006-01-02")
	}

	paramJSON, err := json.Marshal(params)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: invalidParamsMessage}
	}

	var req financeDTO.CreateJournalEntryRequest
	if err := json.Unmarshal(paramJSON, &req); err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: fmt.Sprintf("Failed to parse journal parameters: %s", err.Error())}
	}

	resp, err := e.deps.JournalUsecase.Create(ctx, &req)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: "journal", ErrorCode: "CREATE_FAILED", ErrorMessage: err.Error()}
	}

	return &ActionResult{Success: true, Data: resp, Message: "Journal entry created successfully", EntityType: "journal", EntityID: resp.ID, Action: "CREATE"}
}

func (e *ActionExecutor) executeGenerateReport(ctx context.Context, params map[string]interface{}) *ActionResult {
	reportType := strings.TrimSpace(getStringParam(params, "report_type"))
	if reportType == "" {
		reportType = strings.TrimSpace(getStringParam(params, "type"))
	}
	if reportType == "" {
		reportType = "summary"
	}

	tables := []string{"sales_orders", "purchase_orders", "customer_invoices", "supplier_invoices", "journal_entries", "products"}
	summary := make(map[string]interface{})

	for _, table := range tables {
		var count int64
		if err := e.entityResolver.db.WithContext(ctx).Table(table).Count(&count).Error; err == nil {
			summary[table] = count
		}
	}

	return &ActionResult{
		Success:    true,
		Action:     "QUERY",
		EntityType: "report",
		Message:    fmt.Sprintf("Generated %s report summary", reportType),
		Data: map[string]interface{}{
			"report_type":  reportType,
			"generated_at": apptime.Now().Format(time.RFC3339),
			"summary":      summary,
		},
	}
}

func (e *ActionExecutor) executeListCOA(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.CoaUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Chart of accounts service is not available"}
	}

	req := &financeDTO.ListChartOfAccountsRequest{Page: 1, PerPage: 20}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}
	req.Search = getStringParam(params, "search")

	items, total, err := e.deps.CoaUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "coa", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"accounts": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d chart of accounts", len(items)), EntityType: "coa", Action: "QUERY"}
}

func (e *ActionExecutor) executeListJournals(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.JournalUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Journal service is not available"}
	}

	req := &financeDTO.ListJournalEntriesRequest{Page: 1, PerPage: 20, Search: getStringParam(params, "search")}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}
	if startDate, endDate := resolvePeriodDateRange(getStringParam(params, "period")); startDate != "" && endDate != "" {
		req.StartDate = &startDate
		req.EndDate = &endDate
	}

	items, total, err := e.deps.JournalUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "journal", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"journals": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d journals", len(items)), EntityType: "journal", Action: "QUERY"}
}

func (e *ActionExecutor) executeListFinancePayments(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.FinancePaymentUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Finance payment service is not available"}
	}

	req := &financeDTO.ListPaymentsRequest{Page: 1, PerPage: 20, Search: getStringParam(params, "search")}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}
	if startDate, endDate := resolvePeriodDateRange(getStringParam(params, "period")); startDate != "" && endDate != "" {
		req.StartDate = &startDate
		req.EndDate = &endDate
	}

	items, total, err := e.deps.FinancePaymentUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "payment", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"payments": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d payments", len(items)), EntityType: "payment", Action: "QUERY"}
}

func (e *ActionExecutor) executeListTaxInvoices(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.TaxInvoiceUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Tax invoice service is not available"}
	}

	req := &financeDTO.ListTaxInvoicesRequest{Page: 1, PerPage: 20, Search: getStringParam(params, "search")}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}
	if startDate, endDate := resolvePeriodDateRange(getStringParam(params, "period")); startDate != "" && endDate != "" {
		req.StartDate = &startDate
		req.EndDate = &endDate
	}

	items, total, err := e.deps.TaxInvoiceUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "tax_invoice", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"tax_invoices": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d tax invoices", len(items)), EntityType: "tax_invoice", Action: "QUERY"}
}

func (e *ActionExecutor) executeListBudgets(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.BudgetUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Budget service is not available"}
	}

	req := &financeDTO.ListBudgetsRequest{Page: 1, PerPage: 20, Search: getStringParam(params, "search")}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}
	if startDate, endDate := resolvePeriodDateRange(getStringParam(params, "period")); startDate != "" && endDate != "" {
		req.StartDate = &startDate
		req.EndDate = &endDate
	}

	items, total, err := e.deps.BudgetUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "budget", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"budgets": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d budgets", len(items)), EntityType: "budget", Action: "QUERY"}
}

func (e *ActionExecutor) executeListCashBank(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.CashBankUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Cash bank service is not available"}
	}

	req := &financeDTO.ListCashBankJournalsRequest{Page: 1, PerPage: 20, Search: getStringParam(params, "search")}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}
	if startDate, endDate := resolvePeriodDateRange(getStringParam(params, "period")); startDate != "" && endDate != "" {
		req.StartDate = &startDate
		req.EndDate = &endDate
	}

	items, total, err := e.deps.CashBankUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "cash_bank", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"cash_bank_journals": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d cash bank journals", len(items)), EntityType: "cash_bank", Action: "QUERY"}
}

func (e *ActionExecutor) executeListAssets(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.AssetUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Asset service is not available"}
	}

	req := &financeDTO.ListAssetsRequest{Page: 1, PerPage: 20, Search: getStringParam(params, "search")}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}
	if startDate, endDate := resolvePeriodDateRange(getStringParam(params, "period")); startDate != "" && endDate != "" {
		req.StartDate = &startDate
		req.EndDate = &endDate
	}

	items, total, err := e.deps.AssetUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "asset", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"assets": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d assets", len(items)), EntityType: "asset", Action: "QUERY"}
}

func (e *ActionExecutor) executeListSalary(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.SalaryUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Salary service is not available"}
	}

	req := &hrdDTO.ListSalaryStructuresRequest{Page: 1, PerPage: 20, Search: getStringParam(params, "search")}
	if p := getIntParam(params, "page"); p > 0 {
		req.Page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		req.PerPage = pp
	}

	items, total, err := e.deps.SalaryUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "salary", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"salary_structures": items, "pagination": map[string]interface{}{"page": req.Page, "per_page": req.PerPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d salary records", len(items)), EntityType: "salary", Action: "QUERY"}
}

func (e *ActionExecutor) executeListBankAccounts(ctx context.Context, params map[string]interface{}) *ActionResult {
	if e.deps.BankAccountUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Bank account service is not available"}
	}

	page, perPage := 1, 20
	if p := getIntParam(params, "page"); p > 0 {
		page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		perPage = pp
	}

	req := coreRepos.BankAccountListParams{
		Search: strings.TrimSpace(getStringParam(params, "search")),
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	items, total, err := e.deps.BankAccountUsecase.List(ctx, req)
	if err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: "bank_account", ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return &ActionResult{Success: true, Data: map[string]interface{}{"bank_accounts": items, "pagination": map[string]interface{}{"page": page, "per_page": perPage, "total": total, "total_pages": totalPages}}, Message: fmt.Sprintf("Found %d bank accounts", len(items)), EntityType: "bank_account", Action: "QUERY"}
}

func (e *ActionExecutor) executeListEmployees(ctx context.Context, params map[string]interface{}) *ActionResult {
	return e.executeSimpleListByTable(ctx, "employees", "employee", "employees", []string{"name", "employee_code", "employee_number"}, params)
}

func (e *ActionExecutor) executeQueryEmployee(ctx context.Context, params map[string]interface{}) *ActionResult {
	return e.executeSimpleQueryByTable(ctx, "employees", "employee", []string{"name", "employee_code", "employee_number"}, params)
}

func (e *ActionExecutor) executeSimpleListByTable(ctx context.Context, table, entityType, dataKey string, searchableColumns []string, params map[string]interface{}) *ActionResult {
	if e.entityResolver == nil || e.entityResolver.db == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: dbServiceUnavailableMessage}
	}

	page, perPage := 1, 20
	if p := getIntParam(params, "page"); p > 0 {
		page = p
	}
	if pp := getIntParam(params, "per_page"); pp > 0 {
		perPage = pp
	}

	query := e.entityResolver.db.WithContext(ctx).Table(table)
	countQuery := e.entityResolver.db.WithContext(ctx).Table(table)

	search := strings.TrimSpace(getStringParam(params, "search"))
	if search != "" && len(searchableColumns) > 0 {
		like := search + "%"
		conds := make([]string, 0, len(searchableColumns))
		args := make([]interface{}, 0, len(searchableColumns))
		for _, col := range searchableColumns {
			conds = append(conds, fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", col))
			args = append(args, like)
		}
		condition := strings.Join(conds, " OR ")
		query = query.Where(condition, args...)
		countQuery = countQuery.Where(condition, args...)
	}

	query = query.Select("*").Limit(perPage).Offset((page - 1) * perPage)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: entityType, ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	rows := make([]map[string]interface{}, 0)
	if err := query.Scan(&rows).Error; err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: entityType, ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return &ActionResult{
		Success: true,
		Data: map[string]interface{}{
			dataKey: rows,
			"pagination": map[string]interface{}{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": totalPages,
			},
		},
		Message:    fmt.Sprintf("Found %d %s records", len(rows), strings.ReplaceAll(entityType, "_", " ")),
		EntityType: entityType,
		Action:     "QUERY",
	}
}

func (e *ActionExecutor) executeSimpleCreateByTable(ctx context.Context, table, entityType string, allowedFields []string, requiredFields []string, params map[string]interface{}) *ActionResult {
	if e.entityResolver == nil || e.entityResolver.db == nil {
		return unavailableActionResult("CREATE", dbServiceUnavailableMessage)
	}

	allowed := sliceToSet(allowedFields)
	record := filterMapByAllowedKeys(params, allowed)
	if missing := missingRequiredFields(record, requiredFields); len(missing) > 0 {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: entityType, ErrorCode: "MISSING_PARAMS", ErrorMessage: fmt.Sprintf("Missing required fields: %s", strings.Join(missing, ", "))}
	}

	applyAuditTimestamps(record, allowed)

	if err := e.entityResolver.db.WithContext(ctx).Table(table).Create(&record).Error; err != nil {
		return &ActionResult{Success: false, Action: "CREATE", EntityType: entityType, ErrorCode: "CREATE_FAILED", ErrorMessage: err.Error()}
	}

	return successTableActionResult("CREATE", entityType, record, fmt.Sprintf("%s created successfully", strings.ReplaceAll(entityType, "_", " ")), fmt.Sprint(record["id"]))
}

type simpleStatusUpdateRequest struct {
	table        string
	entityType   string
	idField      string
	statusField  string
	statusValue  string
	params       map[string]interface{}
	extraUpdates map[string]interface{}
}

func (e *ActionExecutor) executeSimpleStatusUpdateByTable(ctx context.Context, req simpleStatusUpdateRequest) *ActionResult {
	if e.entityResolver == nil || e.entityResolver.db == nil {
		return unavailableActionResult("UPDATE", dbServiceUnavailableMessage)
	}

	entityID := strings.TrimSpace(getStringParam(req.params, req.idField))
	if entityID == "" {
		entityID = strings.TrimSpace(getStringParam(req.params, "id"))
	}
	if entityID == "" {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: req.entityType, ErrorCode: "MISSING_PARAMS", ErrorMessage: fmt.Sprintf("Please provide %s or id", req.idField)}
	}

	updates := map[string]interface{}{req.statusField: req.statusValue}
	for k, v := range req.extraUpdates {
		updates[k] = v
	}

	res := e.entityResolver.db.WithContext(ctx).Table(req.table).Where(fmt.Sprintf("%s = ?", req.idField), entityID).Updates(updates)
	if res.Error != nil {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: req.entityType, ErrorCode: "UPDATE_FAILED", ErrorMessage: res.Error.Error()}
	}
	if res.RowsAffected == 0 {
		return &ActionResult{Success: false, Action: "UPDATE", EntityType: req.entityType, ErrorCode: "NOT_FOUND", ErrorMessage: fmt.Sprintf("%s not found", strings.ReplaceAll(req.entityType, "_", " "))}
	}

	data := map[string]interface{}{"id": entityID, req.statusField: req.statusValue}
	return successTableActionResult("UPDATE", req.entityType, data, fmt.Sprintf("%s updated to %s", strings.ReplaceAll(req.entityType, "_", " "), req.statusValue), entityID)
}

func (e *ActionExecutor) executeSimpleQueryByTable(ctx context.Context, table, entityType string, searchableColumns []string, params map[string]interface{}) *ActionResult {
	if e.entityResolver == nil || e.entityResolver.db == nil {
		return unavailableActionResult("QUERY", dbServiceUnavailableMessage)
	}

	query := e.entityResolver.db.WithContext(ctx).Table(table)
	if id := strings.TrimSpace(getStringParam(params, "id")); id != "" {
		query = query.Where("id = ?", id)
	} else {
		search := strings.TrimSpace(getStringParam(params, "search"))
		if search == "" {
			search = strings.TrimSpace(getStringParam(params, "name"))
		}
		if search == "" {
			return &ActionResult{Success: false, Action: "QUERY", EntityType: entityType, ErrorCode: "MISSING_PARAMS", ErrorMessage: "Please provide id, search, or name parameter"}
		}
		if condition, args := buildSearchCondition(search, searchableColumns); condition != "" {
			query = query.Where(condition, args...)
		}
	}

	row := map[string]interface{}{}
	if err := query.Select("*").Limit(1).Scan(&row).Error; err != nil {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: entityType, ErrorCode: "QUERY_FAILED", ErrorMessage: err.Error()}
	}
	if len(row) == 0 {
		return &ActionResult{Success: false, Action: "QUERY", EntityType: entityType, ErrorCode: "NOT_FOUND", ErrorMessage: fmt.Sprintf("%s not found", strings.ReplaceAll(entityType, "_", " "))}
	}

	entityID := ""
	if idVal, ok := row["id"]; ok {
		entityID = fmt.Sprint(idVal)
	}

	return successTableActionResult("QUERY", entityType, row, fmt.Sprintf("%s found", strings.ReplaceAll(entityType, "_", " ")), entityID)
}

func unavailableActionResult(action, message string) *ActionResult {
	return &ActionResult{Success: false, Action: action, ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: message}
}

func successTableActionResult(action, entityType string, data interface{}, message string, entityID ...string) *ActionResult {
	result := &ActionResult{Success: true, Data: data, Message: message, EntityType: entityType, Action: action}
	if len(entityID) > 0 && strings.TrimSpace(entityID[0]) != "" {
		result.EntityID = entityID[0]
	}
	return result
}

func filterMapByAllowedKeys(source map[string]interface{}, allowed map[string]bool) map[string]interface{} {
	filtered := make(map[string]interface{}, len(source))
	for key, value := range source {
		if allowed[key] {
			filtered[key] = value
		}
	}
	return filtered
}

func sliceToSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func missingRequiredFields(record map[string]interface{}, requiredFields []string) []string {
	missing := make([]string, 0)
	for _, field := range requiredFields {
		val, exists := record[field]
		if !exists || val == nil || strings.TrimSpace(fmt.Sprint(val)) == "" {
			missing = append(missing, field)
		}
	}
	return missing
}

func applyAuditTimestamps(record map[string]interface{}, allowed map[string]bool) {
	if allowed["created_at"] {
		if _, exists := record["created_at"]; !exists {
			record["created_at"] = apptime.Now()
		}
	}
	if allowed["updated_at"] {
		if _, exists := record["updated_at"]; !exists {
			record["updated_at"] = apptime.Now()
		}
	}
}

func buildSearchCondition(search string, searchableColumns []string) (string, []interface{}) {
	if len(searchableColumns) == 0 {
		return "", nil
	}
	like := search + "%"
	conds := make([]string, 0, len(searchableColumns))
	args := make([]interface{}, 0, len(searchableColumns))
	for _, col := range searchableColumns {
		conds = append(conds, fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", col))
		args = append(args, like)
	}
	return strings.Join(conds, " OR "), args
}

func marshalParamsToRequest(params map[string]interface{}, target interface{}, entityLabel string) *ActionResult {
	paramJSON, err := json.Marshal(params)
	if err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: invalidParamsMessage}
	}

	if err := json.Unmarshal(paramJSON, target); err != nil {
		return &ActionResult{Success: false, Action: "CREATE", ErrorCode: "INVALID_PARAMS", ErrorMessage: fmt.Sprintf("Failed to parse %s parameters: %s", entityLabel, err.Error())}
	}

	return nil
}

func newEntityNotFoundResult(entityType, message string) *ActionResult {
	return &ActionResult{Success: false, Action: "CREATE", EntityType: entityType, ErrorCode: "ENTITY_NOT_FOUND", ErrorMessage: message}
}

// stockFilterKeywords maps natural language stock-level terms to the low_stock filter.
// These terms should NOT be used as product name search.
var stockFilterKeywords = []string{
	"kurang", "rendah", "low", "minimum", "habis", "kosong", "out of stock",
	"menipis", "sedikit", "dikit", "kritis", "hampir habis", "empty", "shortage",
	"minim", "tipis", "critical",
}

// isStockFilterTerm checks if a search term is a stock-level filter word, not a product name
func isStockFilterTerm(term string) bool {
	lower := strings.TrimSpace(strings.ToLower(term))
	for _, kw := range stockFilterKeywords {
		if lower == kw || strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// --- Inventory Action Implementations ---

func (e *ActionExecutor) executeQueryStock(ctx context.Context, params map[string]interface{}, resolvedEntities map[string]*ResolvedEntity) *ActionResult {
	if e.deps.InventoryUsecase == nil {
		return &ActionResult{Success: false, Action: "QUERY", ErrorCode: "SERVICE_UNAVAILABLE", ErrorMessage: "Inventory service is not available"}
	}

	req := &inventoryDTO.GetInventoryListRequest{
		Page:    1,
		PerPage: 20,
	}

	if product, ok := resolvedEntities["product"]; ok {
		req.ProductID = product.ID
	}
	if warehouse, ok := resolvedEntities["warehouse"]; ok {
		req.WarehouseID = warehouse.ID
	}

	// Handle search parameter — only use it if it's an actual product name, not a filter word
	if search := getStringParam(params, "search"); search != "" {
		if isStockFilterTerm(search) {
			req.LowStock = true
		} else {
			req.Search = search
		}
	}

	// Explicit low_stock flag from parameter extraction
	if lowStock, ok := params["low_stock"].(bool); ok && lowStock {
		req.LowStock = true
	}

	// Detect stock-level keywords in filter/status params
	if filter := getStringParam(params, "filter"); filter != "" && isStockFilterTerm(filter) {
		req.LowStock = true
	}
	if status := getStringParam(params, "status"); status != "" && isStockFilterTerm(status) {
		req.LowStock = true
	}

	resp, err := e.deps.InventoryUsecase.GetStockList(ctx, req)
	if err != nil {
		return &ActionResult{
			Success:      false,
			Action:       "QUERY",
			EntityType:   "inventory",
			ErrorCode:    "QUERY_FAILED",
			ErrorMessage: err.Error(),
		}
	}

	// Post-query filtering: if LowStock requested, filter items by status
	if req.LowStock && resp != nil && len(resp.Data) > 0 {
		var filtered []inventoryDTO.InventoryStockItem
		for _, item := range resp.Data {
			if item.Status == "low_stock" || item.Status == "out_of_stock" {
				filtered = append(filtered, item)
			}
		}
		resp.Data = filtered
		resp.Meta.Total = int64(len(filtered))
	}

	// Build a concise summary for the LLM to avoid confusion with large JSON
	summary := e.buildStockSummary(resp)

	return &ActionResult{
		Success:    true,
		Data:       summary,
		Message:    fmt.Sprintf("Found %d stock items", len(resp.Data)),
		EntityType: "inventory",
		Action:     "QUERY",
	}
}

// buildStockSummary creates a structured, LLM-friendly summary of stock data
func (e *ActionExecutor) buildStockSummary(resp *inventoryDTO.GetInventoryListResponse) map[string]interface{} {
	if resp == nil || len(resp.Data) == 0 {
		return map[string]interface{}{
			"total_items": 0,
			"items":       []interface{}{},
			"message":     "Tidak ada data stok ditemukan",
		}
	}

	items := make([]map[string]interface{}, 0, len(resp.Data))
	for _, item := range resp.Data {
		items = append(items, map[string]interface{}{
			"product_name": item.ProductName,
			"product_code": item.ProductCode,
			"warehouse":    item.WarehouseName,
			"available":    item.Available,
			"on_hand":      item.OnHand,
			"reserved":     item.Reserved,
			"min_stock":    item.MinStock,
			"max_stock":    item.MaxStock,
			"unit":         item.UomName,
			"status":       item.Status,
		})
	}

	return map[string]interface{}{
		"total_items": len(resp.Data),
		"page":        resp.Meta.Page,
		"per_page":    resp.Meta.PerPage,
		"items":       items,
	}
}

// --- Helper functions for parameter extraction ---

func getStringParam(params map[string]interface{}, key string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return ""
}

func getIntParam(params map[string]interface{}, key string) int {
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	if val, ok := params[key].(int); ok {
		return val
	}
	return 0
}

func getFloatParam(params map[string]interface{}, key string) float64 {
	if val, ok := params[key].(float64); ok {
		return val
	}
	if val, ok := params[key].(int); ok {
		return float64(val)
	}
	if val, ok := params[key].(int64); ok {
		return float64(val)
	}
	if val, ok := params[key].(string); ok {
		if parsed, ok := parseAmountString(val); ok {
			return parsed
		}
	}
	return 0
}

func parseAmountString(raw string) (float64, bool) {
	text := strings.TrimSpace(strings.ToLower(raw))
	if text == "" {
		return 0, false
	}

	text = strings.ReplaceAll(text, "rp", "")
	text = strings.ReplaceAll(text, "idr", "")
	text = strings.ReplaceAll(text, " ", "")

	multiplier := 1.0
	switch {
	case strings.HasSuffix(text, "miliar"):
		multiplier = 1_000_000_000
		text = strings.TrimSuffix(text, "miliar")
	case strings.HasSuffix(text, "milyar"):
		multiplier = 1_000_000_000
		text = strings.TrimSuffix(text, "milyar")
	case strings.HasSuffix(text, "bio"):
		multiplier = 1_000_000_000
		text = strings.TrimSuffix(text, "bio")
	case strings.HasSuffix(text, "b"):
		multiplier = 1_000_000_000
		text = strings.TrimSuffix(text, "b")
	case strings.HasSuffix(text, "juta"):
		multiplier = 1_000_000
		text = strings.TrimSuffix(text, "juta")
	case strings.HasSuffix(text, "jt"):
		multiplier = 1_000_000
		text = strings.TrimSuffix(text, "jt")
	case strings.HasSuffix(text, "m"):
		multiplier = 1_000_000
		text = strings.TrimSuffix(text, "m")
	case strings.HasSuffix(text, "ribu"):
		multiplier = 1_000
		text = strings.TrimSuffix(text, "ribu")
	case strings.HasSuffix(text, "rb"):
		multiplier = 1_000
		text = strings.TrimSuffix(text, "rb")
	case strings.HasSuffix(text, "k"):
		multiplier = 1_000
		text = strings.TrimSuffix(text, "k")
	}

	cleaned := regexp.MustCompile(`[^0-9\.,]`).ReplaceAllString(text, "")
	if cleaned == "" {
		return 0, false
	}

	if strings.Contains(cleaned, ".") && strings.Contains(cleaned, ",") {
		lastDot := strings.LastIndex(cleaned, ".")
		lastComma := strings.LastIndex(cleaned, ",")
		if lastDot > lastComma {
			cleaned = strings.ReplaceAll(cleaned, ",", "")
		} else {
			cleaned = strings.ReplaceAll(cleaned, ".", "")
			cleaned = strings.ReplaceAll(cleaned, ",", ".")
		}
	} else if strings.Count(cleaned, ",") > 0 {
		if strings.Count(cleaned, ",") > 1 {
			cleaned = strings.ReplaceAll(cleaned, ",", "")
		} else {
			idx := strings.LastIndex(cleaned, ",")
			digitsAfter := len(cleaned) - idx - 1
			if digitsAfter == 3 && multiplier == 1 {
				cleaned = strings.ReplaceAll(cleaned, ",", "")
			} else {
				cleaned = strings.ReplaceAll(cleaned, ",", ".")
			}
		}
	} else if strings.Count(cleaned, ".") > 0 {
		if strings.Count(cleaned, ".") > 1 {
			cleaned = strings.ReplaceAll(cleaned, ".", "")
		} else {
			idx := strings.LastIndex(cleaned, ".")
			digitsAfter := len(cleaned) - idx - 1
			if digitsAfter == 3 && multiplier == 1 {
				cleaned = strings.ReplaceAll(cleaned, ".", "")
			}
		}
	}

	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, false
	}

	return value * multiplier, true
}
