package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Layer 3: Request Validator — backend validation without any LLM calls
// Validates required fields, entity existence, scope, and business rules

// ValidationResult holds the outcome of server-side parameter validation
type ValidationResult struct {
	Valid            bool                       `json:"valid"`
	Errors           []ValidationError          `json:"errors,omitempty"`
	ResolvedEntities map[string]*ResolvedEntity `json:"resolved_entities,omitempty"`
	SanitizedParams  map[string]interface{}     `json:"sanitized_params,omitempty"`
}

// ValidationError represents a single field-level validation error
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RequestValidator validates parameters and resolves entities without LLM
type RequestValidator struct {
	db             *gorm.DB
	entityResolver *EntityResolver
}

// NewRequestValidator creates a new RequestValidator
func NewRequestValidator(db *gorm.DB, entityResolver *EntityResolver) *RequestValidator {
	return &RequestValidator{
		db:             db,
		entityResolver: entityResolver,
	}
}

// Validate performs server-side validation of extracted parameters for a given intent
func (v *RequestValidator) Validate(ctx context.Context, intent *IntentResult, params map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:            true,
		ResolvedEntities: make(map[string]*ResolvedEntity),
		SanitizedParams:  make(map[string]interface{}),
	}

	// Copy params to sanitized
	for k, val := range params {
		result.SanitizedParams[k] = val
	}

	// Run intent-specific validation rules
	switch {
	// Holiday creation
	case intent.IntentCode == "CREATE_HOLIDAY":
		if isIndonesiaBulkHolidayMode(params) {
			if year := getIntParam(params, "year"); year < 2000 || year > 2100 {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "year",
					Code:    "INVALID_RANGE",
					Message: "Field 'year' harus di antara 2000-2100 untuk create holiday otomatis",
				})
			}
		} else {
			v.validateRequired(result, params, "name", "date")
			v.validateDateFormat(result, params, "date")
		}

	// Leave request
	case intent.IntentCode == "CREATE_LEAVE_REQUEST":
		v.validateRequired(result, params, "start_date", "end_date")
		v.validateDateFormat(result, params, "start_date", "end_date")
		v.validateDateRange(result, params, "start_date", "end_date")

	// Sales quotation — requires customer_name, items with product/qty/price
	// quotation_date and sales_rep_id are auto-filled if missing
	case intent.IntentCode == "CREATE_SALES_QUOTATION":
		v.validateRequired(result, params, "customer_name")
		v.validateSalesQuotationFields(result, params)

	// Sales order — if no sales_quotation_id, require customer and core form fields.
	// sales_rep must be explicitly provided by user to avoid wrong assignment.
	case intent.IntentCode == "CREATE_SALES_ORDER":
		v.validateSalesOrderFields(result, params)

	// Sales target creation — area is required while year/total can be defaulted
	case intent.IntentCode == "CREATE_SALES_TARGET":
		v.validateSalesTargetFields(result, params)

	// Purchase order
	case intent.IntentCode == "CREATE_PURCHASE_ORDER":
		v.validateRequired(result, params, "supplier_name")

	// Stock queries
	case strings.HasPrefix(intent.IntentCode, "QUERY_STOCK") || intent.IntentCode == "LIST_INVENTORY":
		// No required fields, but resolve entities if provided

	// Generic list/query — no required fields
	case strings.HasPrefix(intent.IntentCode, "LIST_") || strings.HasPrefix(intent.IntentCode, "QUERY_"):
		// Nothing required

	// Any CREATE action should have at least one meaningful parameter
	case intent.ActionType == "CREATE" && len(params) == 0:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "_",
			Code:    "MISSING_PARAMS",
			Message: "No parameters provided for creation action",
		})
	}

	// Resolve entity references (customer, employee, product, warehouse)
	v.resolveEntities(ctx, result, params, intent.IntentCode)

	return result
}

func isIndonesiaBulkHolidayMode(params map[string]interface{}) bool {
	if strings.EqualFold(getStringParam(params, "holiday_source"), "PUBLIC_API") && strings.EqualFold(getStringParam(params, "country_code"), "ID") {
		return true
	}

	lowerCountry := strings.ToLower(getStringParam(params, "country"))
	return lowerCountry == "id" || lowerCountry == "indonesia"
}

// validateRequired checks that required string fields are present and non-empty
func (v *RequestValidator) validateRequired(result *ValidationResult, params map[string]interface{}, fields ...string) {
	for _, field := range fields {
		val, ok := params[field]
		if !ok || val == nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Code:    "REQUIRED",
				Message: fmt.Sprintf("Field '%s' is required", field),
			})
			continue
		}
		if str, ok := val.(string); ok && strings.TrimSpace(str) == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Code:    "REQUIRED",
				Message: fmt.Sprintf("Field '%s' cannot be empty", field),
			})
		}
	}
}

// validateDateFormat validates that date strings match YYYY-MM-DD
func (v *RequestValidator) validateDateFormat(result *ValidationResult, params map[string]interface{}, fields ...string) {
	for _, field := range fields {
		val, ok := params[field].(string)
		if !ok || val == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", val); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Code:    "INVALID_FORMAT",
				Message: fmt.Sprintf("Field '%s' must be a valid date (YYYY-MM-DD), got '%s'", field, val),
			})
		}
	}
}

// validateDateRange validates that start_date is before or equal to end_date
func (v *RequestValidator) validateDateRange(result *ValidationResult, params map[string]interface{}, startField, endField string) {
	startStr, ok1 := params[startField].(string)
	endStr, ok2 := params[endField].(string)
	if !ok1 || !ok2 || startStr == "" || endStr == "" {
		return
	}

	start, err1 := time.Parse("2006-01-02", startStr)
	end, err2 := time.Parse("2006-01-02", endStr)
	if err1 != nil || err2 != nil {
		return
	}

	if start.After(end) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   startField,
			Code:    "INVALID_RANGE",
			Message: fmt.Sprintf("'%s' (%s) must be before or equal to '%s' (%s)", startField, startStr, endField, endStr),
		})
	}
}

// validateSalesQuotationFields validates SQ-specific required fields beyond customer_name.
// Informs the user which mandatory fields are missing so they can provide complete data.
func (v *RequestValidator) validateSalesQuotationFields(result *ValidationResult, params map[string]interface{}) {
	var missing []string

	// items is required with at least one product
	items, hasItems := params["items"]
	if !hasItems || items == nil {
		missing = append(missing, "items (daftar produk minimal 1, contoh: 'produk X qty 10 harga 50000')")
	} else if itemSlice, ok := items.([]interface{}); ok && len(itemSlice) == 0 {
		missing = append(missing, "items (daftar produk minimal 1)")
	}

	// payment_terms_id or payment_terms_name (or raw LLM variants) — one must be specified
	if _, ok := params["payment_terms_id"].(string); !ok {
		if _, ok := params["payment_terms_name"].(string); !ok {
			// Also check raw LLM output variants before marking as missing
			if _, ok := params["payment_terms"].(string); !ok {
				if _, ok := params["syarat_pembayaran"].(string); !ok {
					missing = append(missing, "payment_terms (syarat pembayaran, contoh: 'Net 30', 'COD', 'Net 60')")
				}
			}
		}
	}

	// business_unit_id or business_unit_name (or raw LLM variants) — one must be specified
	if _, ok := params["business_unit_id"].(string); !ok {
		if _, ok := params["business_unit_name"].(string); !ok {
			// Also check raw LLM output variants before marking as missing
			if _, ok := params["business_unit"].(string); !ok {
				if _, ok := params["unit_bisnis"].(string); !ok {
					missing = append(missing, "business_unit (unit bisnis, contoh: nama divisi/cabang)")
				}
			}
		}
	}

	if len(missing) > 0 {
		result.Valid = false
		for _, m := range missing {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "sales_quotation",
				Code:    "REQUIRED",
				Message: fmt.Sprintf("Field '%s' diperlukan untuk membuat Sales Quotation", m),
			})
		}
	}
}

// validateSalesOrderFields validates required fields for CREATE_SALES_ORDER.
// If sales_quotation_id is provided, downstream logic can derive most values from quotation.
func (v *RequestValidator) validateSalesOrderFields(result *ValidationResult, params map[string]interface{}) {
	quotationID := strings.TrimSpace(getStringParam(params, "sales_quotation_id"))
	if quotationID != "" {
		return
	}

	customerName := strings.TrimSpace(getStringParam(params, "customer_name"))
	customerID := strings.TrimSpace(getStringParam(params, "customer_id"))
	if customerName == "" && customerID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "customer_name",
			Code:    "REQUIRED",
			Message: "Field 'customer_name' atau 'customer_id' diperlukan untuk membuat Sales Order",
		})
	}

	orderDate := strings.TrimSpace(getStringParam(params, "order_date"))
	if orderDate == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "order_date",
			Code:    "REQUIRED",
			Message: "Field 'order_date' diperlukan untuk membuat Sales Order",
		})
	} else if _, err := time.Parse("2006-01-02", orderDate); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "order_date",
			Code:    "INVALID_FORMAT",
			Message: "Field 'order_date' harus berformat YYYY-MM-DD",
		})
	}

	// NOTE:
	// payment_terms, business_unit, and sales_rep may be filled by execution defaults
	// (name→ID resolution, current-user sales rep fallback, default active business unit).
	// We intentionally avoid hard-blocking at this pre-confirmation stage so pending
	// actions can still be created and resolved during execution.

	itemsRaw, hasItems := params["items"]
	if !hasItems || itemsRaw == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "items",
			Code:    "REQUIRED",
			Message: "Field 'items' diperlukan untuk membuat Sales Order (minimal 1 item)",
		})
		return
	}

	items, ok := itemsRaw.([]interface{})
	if !ok || len(items) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "items",
			Code:    "REQUIRED",
			Message: "Field 'items' diperlukan untuk membuat Sales Order (minimal 1 item)",
		})
		return
	}

	for idx, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "items",
				Code:    "INVALID_FORMAT",
				Message: fmt.Sprintf("Item ke-%d harus berupa object dengan product_name, quantity, dan price", idx+1),
			})
			continue
		}

		productName := strings.TrimSpace(getStringParam(item, "product_name"))
		productID := strings.TrimSpace(getStringParam(item, "product_id"))
		if productName == "" && productID == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "items.product",
				Code:    "REQUIRED",
				Message: fmt.Sprintf("Item ke-%d wajib memiliki product_name atau product_id", idx+1),
			})
		}

		if getFloatParam(item, "quantity") <= 0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "items.quantity",
				Code:    "REQUIRED",
				Message: fmt.Sprintf("Item ke-%d wajib memiliki quantity > 0", idx+1),
			})
		}

		if getFloatParam(item, "price") <= 0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "items.price",
				Code:    "REQUIRED",
				Message: fmt.Sprintf("Item ke-%d wajib memiliki price > 0", idx+1),
			})
		}
	}
}

// validateSalesTargetFields validates required fields for CREATE_SALES_TARGET.
// We require area for operational clarity, while year/total_target can be defaulted in executor.
func (v *RequestValidator) validateSalesTargetFields(result *ValidationResult, params map[string]interface{}) {
	areaName := strings.TrimSpace(getStringParam(params, "area_name"))
	areaID := strings.TrimSpace(getStringParam(params, "area_id"))
	if areaName == "" && areaID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "area_name",
			Code:    "REQUIRED",
			Message: "Field 'area_name' atau 'area_id' diperlukan untuk membuat sales target",
		})
	}
}

// resolveEntities resolves natural language entity names to DB IDs
func (v *RequestValidator) resolveEntities(ctx context.Context, result *ValidationResult, params map[string]interface{}, intentCode string) {
	// Resolve employee
	if empName, ok := params["employee_name"].(string); ok && empName != "" {
		entity, err := v.entityResolver.ResolveEmployee(ctx, empName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "employee_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Employee '%s' not found", empName),
			})
			// Non-blocking: still valid for queries, but log the warning
		} else {
			result.ResolvedEntities["employee"] = entity
		}
	}

	// Resolve sales representative by name
	if salesRepName, ok := params["sales_rep_name"].(string); ok && strings.TrimSpace(salesRepName) != "" {
		entity, err := v.entityResolver.ResolveEmployee(ctx, salesRepName)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "sales_rep_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Sales rep '%s' not found", salesRepName),
			})
		} else {
			result.ResolvedEntities["sales_rep"] = entity
		}
	}

	// Resolve product
	if prodName, ok := params["product_name"].(string); ok && prodName != "" {
		entity, err := v.entityResolver.ResolveProduct(ctx, prodName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "product_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Product '%s' not found", prodName),
			})
		} else {
			result.ResolvedEntities["product"] = entity
		}
	}

	// Resolve warehouse
	if whName, ok := params["warehouse_name"].(string); ok && whName != "" {
		entity, err := v.entityResolver.ResolveWarehouse(ctx, whName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "warehouse_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Warehouse '%s' not found", whName),
			})
		} else {
			result.ResolvedEntities["warehouse"] = entity
		}
	}

	// Resolve customer (from sales quotations — no FK table, just name match)
	if custName, ok := params["customer_name"].(string); ok && custName != "" {
		entity, err := v.entityResolver.ResolveCustomer(ctx, custName)
		if err != nil {
			if shouldRequireExistingCustomerForIntent(intentCode) || strings.Contains(err.Error(), "ENTITY_AMBIGUOUS") {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "customer_name",
					Code:    "ENTITY_NOT_FOUND",
					Message: err.Error(),
				})
			} else {
				// For non-strict intents, allow free-text customer name.
				result.ResolvedEntities["customer"] = &ResolvedEntity{
					DisplayName: custName,
					EntityType:  "customer",
				}
			}
		} else {
			result.ResolvedEntities["customer"] = entity
		}
	}

	// Resolve supplier
	if suppName, ok := params["supplier_name"].(string); ok && suppName != "" {
		entity, err := v.entityResolver.ResolveSupplier(ctx, suppName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "supplier_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Supplier '%s' not found", suppName),
			})
		} else {
			result.ResolvedEntities["supplier"] = entity
		}
	}

	// Resolve bank
	if bankName, ok := params["bank_name"].(string); ok && strings.TrimSpace(bankName) != "" {
		entity, err := v.entityResolver.ResolveBank(ctx, bankName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "bank_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Bank '%s' not found", bankName),
			})
		} else {
			result.ResolvedEntities["bank"] = entity
		}
	}

	// Resolve company
	if companyName, ok := params["company_name"].(string); ok && strings.TrimSpace(companyName) != "" {
		entity, err := v.entityResolver.ResolveCompany(ctx, companyName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "company_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Company '%s' not found", companyName),
			})
		} else {
			result.ResolvedEntities["company"] = entity
		}
	}

	// Resolve division
	if divisionName, ok := params["division_name"].(string); ok && strings.TrimSpace(divisionName) != "" {
		entity, err := v.entityResolver.ResolveDivision(ctx, divisionName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "division_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Division '%s' not found", divisionName),
			})
		} else {
			result.ResolvedEntities["division"] = entity
		}
	}

	// Resolve business type
	if businessTypeName, ok := params["business_type_name"].(string); ok && strings.TrimSpace(businessTypeName) != "" {
		entity, err := v.entityResolver.ResolveBusinessType(ctx, businessTypeName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "business_type_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Business type '%s' not found", businessTypeName),
			})
		} else {
			result.ResolvedEntities["business_type"] = entity
		}
	}

	// Resolve supplier type
	if supplierTypeName, ok := params["supplier_type_name"].(string); ok && strings.TrimSpace(supplierTypeName) != "" {
		entity, err := v.entityResolver.ResolveSupplierType(ctx, supplierTypeName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "supplier_type_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Supplier type '%s' not found", supplierTypeName),
			})
		} else {
			result.ResolvedEntities["supplier_type"] = entity
		}
	}

	// Resolve customer type
	if customerTypeName, ok := params["customer_type_name"].(string); ok && strings.TrimSpace(customerTypeName) != "" {
		entity, err := v.entityResolver.ResolveCustomerType(ctx, customerTypeName)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "customer_type_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Customer type '%s' not found", customerTypeName),
			})
		} else {
			result.ResolvedEntities["customer_type"] = entity
		}
	}

	// Resolve area
	if areaName, ok := params["area_name"].(string); ok && strings.TrimSpace(areaName) != "" {
		entity, err := v.entityResolver.ResolveArea(ctx, areaName)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "area_name",
				Code:    "ENTITY_NOT_FOUND",
				Message: fmt.Sprintf("Area '%s' tidak ditemukan", areaName),
			})
		} else {
			result.ResolvedEntities["area"] = entity
		}
	}
}

func shouldRequireExistingCustomerForIntent(intentCode string) bool {
	intentCode = strings.ToUpper(strings.TrimSpace(intentCode))
	strictIntents := map[string]bool{
		"CREATE_SALES_ORDER":   true,
		"CREATE_SALES_INVOICE": true,
	}
	return strictIntents[intentCode]
}
