package usecase

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// ResolvedEntity holds a resolved entity ID with its display name
type ResolvedEntity struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	EntityType  string `json:"entity_type"`
	Code        string `json:"code,omitempty"`
}

// FormDataOptions holds available options for form field dropdowns
type FormDataOptions struct {
	Customers     []string
	Products      []FormDataProduct
	PaymentTerms  []FormDataOption
	BusinessUnits []FormDataOption
}

// FormDataProduct represents a product option for form selection
type FormDataProduct struct {
	Name string
	SKU  string
}

// FormDataOption represents a generic named option
type FormDataOption struct {
	Name string
	Code string
}

var nonAlphaNumPattern = regexp.MustCompile(`[^a-z0-9]+`)

// EntityResolver resolves natural language entity references to database IDs
type EntityResolver struct {
	db *gorm.DB
}

// NewEntityResolver creates a new EntityResolver
func NewEntityResolver(db *gorm.DB) *EntityResolver {
	return &EntityResolver{db: db}
}

// ResolveEmployee resolves employee name/code to employee ID
func (r *EntityResolver) ResolveEmployee(ctx context.Context, nameOrCode string) (*ResolvedEntity, error) {
	var result struct {
		ID           string
		EmployeeCode string
		Name         string
	}

	query := r.db.WithContext(ctx).Table("employees").
		Select("id, employee_code, name").
		Where("deleted_at IS NULL")

	// Try exact code match first, then prefix name search
	err := query.Where("LOWER(employee_code) = LOWER(?)", nameOrCode).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: employee query error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "employee",
			Code:        result.EmployeeCode,
		}, nil
	}

	// Try name search with prefix matching for index usage
	searchTerm := strings.TrimSpace(nameOrCode)
	err = r.db.WithContext(ctx).Table("employees").
		Select("id, employee_code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: employee search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "employee",
			Code:        result.EmployeeCode,
		}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: employee '%s' not found", nameOrCode)
}

// ResolveProduct resolves product name/code/SKU to product ID
func (r *EntityResolver) ResolveProduct(ctx context.Context, nameOrCode string) (*ResolvedEntity, error) {
	var result struct {
		ID   string
		SKU  string
		Name string
	}

	searchTerm := strings.TrimSpace(nameOrCode)

	// Try exact code/SKU match first
	err := r.db.WithContext(ctx).Table("products").
		Select("id, sku, name").
		Where("deleted_at IS NULL").
		Where("LOWER(sku) = LOWER(?) OR LOWER(code) = LOWER(?)", searchTerm, searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product code query error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "product",
			Code:        result.SKU,
		}, nil
	}

	// Try name prefix search
	err = r.db.WithContext(ctx).Table("products").
		Select("id, sku, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "product",
			Code:        result.SKU,
		}, nil
	}

	// Try normalized search to handle user variants like:
	// "paracetamol 500mg" -> "Paracetamol 500 mg Tablet"
	normalizedSearch := normalizeAlphaNum(searchTerm)
	if normalizedSearch != "" {
		err = r.db.WithContext(ctx).Table("products").
			Select("id, sku, name").
			Where("deleted_at IS NULL").
			Where("regexp_replace(LOWER(name), '[^a-z0-9]+', '', 'g') LIKE ?", normalizedSearch+"%").
			Order("LENGTH(name) ASC").
			Limit(1).Scan(&result).Error
		if err != nil {
			return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product normalized search error: %w", err)
		}
		if result.ID != "" {
			return &ResolvedEntity{
				ID:          result.ID,
				DisplayName: result.Name,
				EntityType:  "product",
				Code:        result.SKU,
			}, nil
		}

		err = r.db.WithContext(ctx).Table("products").
			Select("id, sku, name").
			Where("deleted_at IS NULL").
			Where("regexp_replace(LOWER(name), '[^a-z0-9]+', '', 'g') LIKE ?", "%"+normalizedSearch+"%").
			Order("LENGTH(name) ASC").
			Limit(1).Scan(&result).Error
		if err != nil {
			return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product fuzzy search error: %w", err)
		}
		if result.ID != "" {
			return &ResolvedEntity{
				ID:          result.ID,
				DisplayName: result.Name,
				EntityType:  "product",
				Code:        result.SKU,
			}, nil
		}
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: product '%s' not found", nameOrCode)
}

// ResolveWarehouse resolves warehouse name/code to warehouse ID
func (r *EntityResolver) ResolveWarehouse(ctx context.Context, nameOrCode string) (*ResolvedEntity, error) {
	var result struct {
		ID   string
		Code string
		Name string
	}

	searchTerm := strings.TrimSpace(nameOrCode)

	// Try exact code match first
	err := r.db.WithContext(ctx).Table("warehouses").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(code) = LOWER(?)", searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: warehouse code query error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "warehouse",
			Code:        result.Code,
		}, nil
	}

	// Try name prefix search
	err = r.db.WithContext(ctx).Table("warehouses").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: warehouse search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "warehouse",
			Code:        result.Code,
		}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: warehouse '%s' not found", nameOrCode)
}

// ResolveCustomer resolves customer name from sales documents
func (r *EntityResolver) ResolveCustomer(ctx context.Context, name string) (*ResolvedEntity, error) {
	searchTerm := strings.TrimSpace(name)
	if searchTerm == "" {
		return nil, fmt.Errorf("ENTITY_NOT_FOUND: customer '%s' not found", name)
	}

	var master struct {
		ID   string
		Code string
		Name string
	}
	var masters []struct {
		ID   string
		Code string
		Name string
	}

	// 1) Prefer master customer table for canonical ID + name.
	// Try exact code first because it is expected to be unique and deterministic.
	err := r.db.WithContext(ctx).Table("customers").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(code) = LOWER(?)", searchTerm).
		Limit(1).Scan(&master).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: customer master code query error: %w", err)
	}
	if master.ID != "" {
		return &ResolvedEntity{ID: master.ID, DisplayName: master.Name, EntityType: "customer", Code: master.Code}, nil
	}

	// Exact name match can be ambiguous when multiple legal entities share a base name.
	err = r.db.WithContext(ctx).Table("customers").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) = LOWER(?)", searchTerm).
		Order("name ASC").
		Limit(3).Find(&masters).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: customer master exact name query error: %w", err)
	}
	if len(masters) == 1 {
		match := masters[0]
		return &ResolvedEntity{ID: match.ID, DisplayName: match.Name, EntityType: "customer", Code: match.Code}, nil
	}
	if len(masters) > 1 {
		return nil, fmt.Errorf("ENTITY_AMBIGUOUS: customer '%s' is ambiguous. Candidates: %s", searchTerm, formatCustomerCandidates(masters))
	}

	err = r.db.WithContext(ctx).Table("customers").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", "%"+searchTerm+"%").
		Order("name ASC").
		Limit(3).Find(&masters).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: customer master fuzzy search error: %w", err)
	}
	if len(masters) == 1 {
		match := masters[0]
		return &ResolvedEntity{ID: match.ID, DisplayName: match.Name, EntityType: "customer", Code: match.Code}, nil
	}
	if len(masters) > 1 {
		return nil, fmt.Errorf("ENTITY_AMBIGUOUS: customer '%s' matches multiple records. Candidates: %s", searchTerm, formatCustomerCandidates(masters))
	}

	// 2) Fallback to transactional tables for compatibility with legacy data.
	var transactional struct {
		CustomerName string
	}

	err = r.db.WithContext(ctx).Table("sales_quotations").
		Select("customer_name").
		Where("deleted_at IS NULL").
		Where("LOWER(customer_name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&transactional).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: customer quotation search error: %w", err)
	}
	if transactional.CustomerName != "" {
		return &ResolvedEntity{ID: "", DisplayName: transactional.CustomerName, EntityType: "customer"}, nil
	}

	err = r.db.WithContext(ctx).Table("sales_estimations").
		Select("customer_name").
		Where("deleted_at IS NULL").
		Where("LOWER(customer_name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&transactional).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: customer estimation search error: %w", err)
	}
	if transactional.CustomerName != "" {
		return &ResolvedEntity{ID: "", DisplayName: transactional.CustomerName, EntityType: "customer"}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: customer '%s' not found", name)
}

func formatCustomerCandidates(candidates []struct {
	ID   string
	Code string
	Name string
}) string {
	options := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		label := strings.TrimSpace(candidate.Name)
		if candidate.Code != "" {
			label = fmt.Sprintf("%s (%s)", label, candidate.Code)
		}
		options = append(options, label)
	}

	if len(options) == 0 {
		return ""
	}

	return strings.Join(options, ", ")
}

// ResolveSupplier resolves supplier name/code to supplier ID
func (r *EntityResolver) ResolveSupplier(ctx context.Context, nameOrCode string) (*ResolvedEntity, error) {
	var result struct {
		ID   string
		Code string
		Name string
	}

	searchTerm := strings.TrimSpace(nameOrCode)

	// Try exact code match first
	err := r.db.WithContext(ctx).Table("suppliers").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(code) = LOWER(?)", searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: supplier code query error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "supplier",
			Code:        result.Code,
		}, nil
	}

	// Try name prefix search
	err = r.db.WithContext(ctx).Table("suppliers").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: supplier search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "supplier",
			Code:        result.Code,
		}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: supplier '%s' not found", nameOrCode)
}

// ResolveArea resolves area name/code to area ID
func (r *EntityResolver) ResolveArea(ctx context.Context, nameOrCode string) (*ResolvedEntity, error) {
	var result struct {
		ID   string
		Code string
		Name string
	}

	searchTerm := sanitizeAreaSearch(strings.TrimSpace(nameOrCode))
	if searchTerm == "" {
		return nil, fmt.Errorf("ENTITY_NOT_FOUND: area '%s' not found", nameOrCode)
	}

	// Try exact code/name match first
	err := r.db.WithContext(ctx).Table("areas").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(code) = LOWER(?) OR LOWER(name) = LOWER(?)", searchTerm, searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: area query error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "area",
			Code:        result.Code,
		}, nil
	}

	// Fallback to prefix search
	err = r.db.WithContext(ctx).Table("areas").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: area search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "area",
			Code:        result.Code,
		}, nil
	}

	// Last attempt using contains for phrases like "bali full"
	err = r.db.WithContext(ctx).Table("areas").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", "%"+searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: area contains search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{
			ID:          result.ID,
			DisplayName: result.Name,
			EntityType:  "area",
			Code:        result.Code,
		}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: area '%s' not found", nameOrCode)
}

func sanitizeAreaSearch(v string) string {
	lower := strings.ToLower(strings.TrimSpace(v))
	replacements := []string{" area ", " full", " seluruh", " semua", "region", "wilayah"}
	for _, token := range replacements {
		lower = strings.ReplaceAll(lower, token, " ")
	}
	return strings.TrimSpace(lower)
}

// ResolveUserToEmployeeID looks up the employee record linked to a user account.
// Returns the employee ID (UUID) that corresponds to the given user ID.
func (r *EntityResolver) ResolveUserToEmployeeID(ctx context.Context, userID string) (string, error) {
	var result struct {
		ID string
	}
	err := r.db.WithContext(ctx).Table("employees").
		Select("id").
		Where("deleted_at IS NULL AND user_id = ?", userID).
		Limit(1).Scan(&result).Error
	if err != nil {
		return "", fmt.Errorf("ENTITY_RESOLUTION_FAILED: employee lookup by user_id error: %w", err)
	}
	if result.ID != "" {
		return result.ID, nil
	}
	return "", fmt.Errorf("ENTITY_NOT_FOUND: no employee record linked to user '%s'", userID)
}

// ResolveEntitiesFromParameters resolves all entity references in intent parameters
func (r *EntityResolver) ResolveEntitiesFromParameters(ctx context.Context, params map[string]interface{}) (map[string]*ResolvedEntity, error) {
	resolved := make(map[string]*ResolvedEntity)

	// Resolve employee references
	if empName, ok := params["employee_name"].(string); ok && empName != "" {
		entity, err := r.ResolveEmployee(ctx, empName)
		if err != nil {
			return resolved, err
		}
		resolved["employee"] = entity
	}

	// Resolve product references
	if prodName, ok := params["product_name"].(string); ok && prodName != "" {
		entity, err := r.ResolveProduct(ctx, prodName)
		if err != nil {
			return resolved, err
		}
		resolved["product"] = entity
	}

	// Resolve warehouse references
	if whName, ok := params["warehouse_name"].(string); ok && whName != "" {
		entity, err := r.ResolveWarehouse(ctx, whName)
		if err != nil {
			return resolved, err
		}
		resolved["warehouse"] = entity
	}

	// Resolve customer references
	if custName, ok := params["customer_name"].(string); ok && custName != "" {
		entity, err := r.ResolveCustomer(ctx, custName)
		if err != nil {
			return resolved, err
		}
		resolved["customer"] = entity
	}

	// Resolve area references
	if areaName, ok := params["area_name"].(string); ok && areaName != "" {
		entity, err := r.ResolveArea(ctx, areaName)
		if err != nil {
			return resolved, err
		}
		resolved["area"] = entity
	}

	// Resolve bank references
	if bankName, ok := params["bank_name"].(string); ok && bankName != "" {
		entity, err := r.ResolveBank(ctx, bankName)
		if err != nil {
			return resolved, err
		}
		resolved["bank"] = entity
	}

	// Resolve company references
	if companyName, ok := params["company_name"].(string); ok && companyName != "" {
		entity, err := r.ResolveCompany(ctx, companyName)
		if err != nil {
			return resolved, err
		}
		resolved["company"] = entity
	}

	// Resolve division references
	if divisionName, ok := params["division_name"].(string); ok && divisionName != "" {
		entity, err := r.ResolveDivision(ctx, divisionName)
		if err != nil {
			return resolved, err
		}
		resolved["division"] = entity
	}

	// Resolve business type references
	if btName, ok := params["business_type_name"].(string); ok && btName != "" {
		entity, err := r.ResolveBusinessType(ctx, btName)
		if err != nil {
			return resolved, err
		}
		resolved["business_type"] = entity
	}

	// Resolve supplier type references
	if stName, ok := params["supplier_type_name"].(string); ok && stName != "" {
		entity, err := r.ResolveSupplierType(ctx, stName)
		if err != nil {
			return resolved, err
		}
		resolved["supplier_type"] = entity
	}

	// Resolve customer type references
	if ctName, ok := params["customer_type_name"].(string); ok && ctName != "" {
		entity, err := r.ResolveCustomerType(ctx, ctName)
		if err != nil {
			return resolved, err
		}
		resolved["customer_type"] = entity
	}

	return resolved, nil
}

func (r *EntityResolver) ResolveBank(ctx context.Context, nameOrCode string) (*ResolvedEntity, error) {
	var result struct {
		ID   string
		Code string
		Name string
	}

	searchTerm := strings.TrimSpace(nameOrCode)
	err := r.db.WithContext(ctx).Table("banks").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(code) = LOWER(?) OR LOWER(name) = LOWER(?)", searchTerm, searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: bank query error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{ID: result.ID, DisplayName: result.Name, EntityType: "bank", Code: result.Code}, nil
	}

	err = r.db.WithContext(ctx).Table("banks").
		Select("id, code, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: bank search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{ID: result.ID, DisplayName: result.Name, EntityType: "bank", Code: result.Code}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: bank '%s' not found", nameOrCode)
}

func (r *EntityResolver) ResolveCompany(ctx context.Context, name string) (*ResolvedEntity, error) {
	var result struct {
		ID   string
		Name string
	}

	searchTerm := strings.TrimSpace(name)
	err := r.db.WithContext(ctx).Table("companies").
		Select("id, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) = LOWER(?)", searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: company query error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{ID: result.ID, DisplayName: result.Name, EntityType: "company"}, nil
	}

	err = r.db.WithContext(ctx).Table("companies").
		Select("id, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: company search error: %w", err)
	}
	if result.ID != "" {
		return &ResolvedEntity{ID: result.ID, DisplayName: result.Name, EntityType: "company"}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: company '%s' not found", name)
}

func (r *EntityResolver) ResolveDivision(ctx context.Context, name string) (*ResolvedEntity, error) {
	return r.resolveNameOnlyEntity(ctx, "divisions", "division", name)
}

func (r *EntityResolver) ResolveBusinessType(ctx context.Context, name string) (*ResolvedEntity, error) {
	return r.resolveNameOnlyEntity(ctx, "business_types", "business_type", name)
}

func (r *EntityResolver) ResolveSupplierType(ctx context.Context, name string) (*ResolvedEntity, error) {
	return r.resolveNameOnlyEntity(ctx, "supplier_types", "supplier_type", name)
}

func (r *EntityResolver) ResolveCustomerType(ctx context.Context, name string) (*ResolvedEntity, error) {
	return r.resolveNameOnlyEntity(ctx, "customer_types", "customer_type", name)
}

func (r *EntityResolver) resolveNameOnlyEntity(ctx context.Context, table, entityType, name string) (*ResolvedEntity, error) {
	var result struct {
		ID   string
		Name string
	}

	searchTerm := strings.TrimSpace(name)
	err := r.db.WithContext(ctx).Table(table).
		Select("id, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) = LOWER(?)", searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: %s query error: %w", entityType, err)
	}
	if result.ID != "" {
		return &ResolvedEntity{ID: result.ID, DisplayName: result.Name, EntityType: entityType}, nil
	}

	err = r.db.WithContext(ctx).Table(table).
		Select("id, name").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("ENTITY_RESOLUTION_FAILED: %s search error: %w", entityType, err)
	}
	if result.ID != "" {
		return &ResolvedEntity{ID: result.ID, DisplayName: result.Name, EntityType: entityType}, nil
	}

	return nil, fmt.Errorf("ENTITY_NOT_FOUND: %s '%s' not found", entityType, name)
}

// ResolvePaymentTerms resolves a payment terms name to its database UUID
func (r *EntityResolver) ResolvePaymentTerms(ctx context.Context, name string) (string, error) {
	var terms []struct {
		ID   string
		Name string
		Code string
	}
	searchTerm := strings.TrimSpace(name)
	if searchTerm == "" {
		return "", fmt.Errorf("ENTITY_NOT_FOUND: payment terms '%s' not found", name)
	}

	err := r.db.WithContext(ctx).Table("payment_terms").
		Select("id, name, code").
		Where("deleted_at IS NULL AND is_active = true").
		Find(&terms).Error
	if err != nil {
		return "", fmt.Errorf("ENTITY_RESOLUTION_FAILED: payment terms query error: %w", err)
	}
	if len(terms) == 0 {
		return "", fmt.Errorf("ENTITY_NOT_FOUND: payment terms '%s' not found", name)
	}

	aliases := paymentTermsAliases(searchTerm)
	requestIsCOD := isCODAlias(searchTerm)

	type scoredTerm struct {
		id    string
		score int
	}
	scored := make([]scoredTerm, 0, len(terms))

	for _, term := range terms {
		nameNorm := normalizeLookupToken(term.Name)
		codeNorm := normalizeLookupToken(term.Code)
		bestScore := 0

		for _, alias := range aliases {
			score := 0
			switch {
			case alias != "" && codeNorm == alias:
				score = 100
			case alias != "" && nameNorm == alias:
				score = 95
			case alias != "" && strings.HasPrefix(nameNorm, alias):
				score = 80
			case alias != "" && strings.Contains(nameNorm, alias):
				score = 60
			}
			if score > bestScore {
				bestScore = score
			}
		}

		if bestScore > 0 {
			scored = append(scored, scoredTerm{id: term.ID, score: bestScore})
		}
	}

	if len(scored) == 0 {
		return "", fmt.Errorf("ENTITY_NOT_FOUND: payment terms '%s' not found", name)
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].id < scored[j].id
		}
		return scored[i].score > scored[j].score
	})

	// For COD aliases, require a high-confidence exact code/name mapping.
	if requestIsCOD && scored[0].score < 95 {
		return "", fmt.Errorf("ENTITY_NOT_FOUND: payment terms '%s' not found", name)
	}

	return scored[0].id, nil
}

func paymentTermsAliases(raw string) []string {
	normalized := normalizeLookupToken(raw)
	if normalized == "" {
		return nil
	}

	set := map[string]struct{}{normalized: {}}

	if isCODAlias(raw) {
		set["cod"] = struct{}{}
		set["cashondelivery"] = struct{}{}
		set["tunai"] = struct{}{}
	}

	aliases := make([]string, 0, len(set))
	for alias := range set {
		aliases = append(aliases, alias)
	}

	return aliases
}

func isCODAlias(raw string) bool {
	norm := normalizeLookupToken(raw)
	switch norm {
	case "cod", "cashondelivery", "cashdelivery", "cash", "tunai":
		return true
	default:
		return false
	}
}

func normalizeLookupToken(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if lower == "" {
		return ""
	}
	re := regexp.MustCompile(`[^a-z0-9]+`)
	return re.ReplaceAllString(lower, "")
}

// ResolveBusinessUnit resolves a business unit name to its database UUID
func (r *EntityResolver) ResolveBusinessUnit(ctx context.Context, name string) (string, error) {
	var result struct {
		ID   string
		Name string
	}
	searchTerm := strings.TrimSpace(name)

	// Try exact name match (case-insensitive)
	err := r.db.WithContext(ctx).Table("business_units").
		Select("id, name").
		Where("deleted_at IS NULL AND is_active = true").
		Where("LOWER(name) = LOWER(?)", searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return "", fmt.Errorf("ENTITY_RESOLUTION_FAILED: business unit query error: %w", err)
	}
	if result.ID != "" {
		return result.ID, nil
	}

	// Try prefix search
	err = r.db.WithContext(ctx).Table("business_units").
		Select("id, name").
		Where("deleted_at IS NULL AND is_active = true").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return "", fmt.Errorf("ENTITY_RESOLUTION_FAILED: business unit search error: %w", err)
	}
	if result.ID != "" {
		return result.ID, nil
	}

	// Try contains search (e.g. "Retail" matching "Unit A - Retail")
	err = r.db.WithContext(ctx).Table("business_units").
		Select("id, name").
		Where("deleted_at IS NULL AND is_active = true").
		Where("LOWER(name) LIKE LOWER(?)", "%"+searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return "", fmt.Errorf("ENTITY_RESOLUTION_FAILED: business unit fuzzy search error: %w", err)
	}
	if result.ID != "" {
		return result.ID, nil
	}

	return "", fmt.Errorf("ENTITY_NOT_FOUND: business unit '%s' not found", name)
}

// ResolveDefaultBusinessUnit returns the first active business unit as a fallback.
func (r *EntityResolver) ResolveDefaultBusinessUnit(ctx context.Context) (string, error) {
	var result struct {
		ID string
	}

	err := r.db.WithContext(ctx).Table("business_units").
		Select("id").
		Where("deleted_at IS NULL AND is_active = true").
		Order("name ASC").
		Limit(1).Scan(&result).Error
	if err != nil {
		return "", fmt.Errorf("ENTITY_RESOLUTION_FAILED: default business unit query error: %w", err)
	}
	if strings.TrimSpace(result.ID) == "" {
		return "", fmt.Errorf("ENTITY_NOT_FOUND: no active business unit available")
	}

	return result.ID, nil
}

// ResolveProductByName resolves a product name to its database UUID and selling price
func (r *EntityResolver) ResolveProductByName(ctx context.Context, name string) (id string, price float64, err error) {
	var result struct {
		ID           string
		Name         string
		SellingPrice float64
	}
	searchTerm := strings.TrimSpace(name)

	// Try exact name match (case-insensitive)
	err = r.db.WithContext(ctx).Table("products").
		Select("id, name, selling_price").
		Where("deleted_at IS NULL").
		Where("LOWER(name) = LOWER(?)", searchTerm).
		Limit(1).Scan(&result).Error
	if err != nil {
		return "", 0, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product query error: %w", err)
	}
	if result.ID != "" {
		return result.ID, result.SellingPrice, nil
	}

	// Try prefix search for partial names
	err = r.db.WithContext(ctx).Table("products").
		Select("id, name, selling_price").
		Where("deleted_at IS NULL").
		Where("LOWER(name) LIKE LOWER(?)", searchTerm+"%").
		Limit(1).Scan(&result).Error
	if err != nil {
		return "", 0, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product search error: %w", err)
	}
	if result.ID != "" {
		return result.ID, result.SellingPrice, nil
	}

	// Try normalized prefix match to support common medicine naming variations.
	normalizedSearch := normalizeAlphaNum(searchTerm)
	if normalizedSearch != "" {
		err = r.db.WithContext(ctx).Table("products").
			Select("id, name, selling_price").
			Where("deleted_at IS NULL").
			Where("regexp_replace(LOWER(name), '[^a-z0-9]+', '', 'g') LIKE ?", normalizedSearch+"%").
			Order("LENGTH(name) ASC").
			Limit(1).Scan(&result).Error
		if err != nil {
			return "", 0, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product normalized search error: %w", err)
		}
		if result.ID != "" {
			return result.ID, result.SellingPrice, nil
		}

		// Last fallback: normalized contains search.
		err = r.db.WithContext(ctx).Table("products").
			Select("id, name, selling_price").
			Where("deleted_at IS NULL").
			Where("regexp_replace(LOWER(name), '[^a-z0-9]+', '', 'g') LIKE ?", "%"+normalizedSearch+"%").
			Order("LENGTH(name) ASC").
			Limit(1).Scan(&result).Error
		if err != nil {
			return "", 0, fmt.Errorf("ENTITY_RESOLUTION_FAILED: product fuzzy search error: %w", err)
		}
		if result.ID != "" {
			return result.ID, result.SellingPrice, nil
		}
	}

	return "", 0, fmt.Errorf("ENTITY_NOT_FOUND: product '%s' not found", name)
}

func normalizeAlphaNum(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower == "" {
		return ""
	}
	return nonAlphaNumPattern.ReplaceAllString(lower, "")
}

// FetchFormDataOptions queries the database for available form options based on intent
func (r *EntityResolver) FetchFormDataOptions(ctx context.Context, intentCode string) *FormDataOptions {
	opts := &FormDataOptions{}

	switch intentCode {
	case "CREATE_SALES_QUOTATION", "CREATE_SALES_ORDER":
		// Fetch distinct customer names from existing sales documents
		var customers []struct{ CustomerName string }
		r.db.WithContext(ctx).Table("sales_quotations").
			Select("DISTINCT customer_name").
			Where("deleted_at IS NULL AND customer_name != ''").
			Order("customer_name").
			Limit(10).Scan(&customers)
		for _, c := range customers {
			opts.Customers = append(opts.Customers, c.CustomerName)
		}

		// Fetch active products
		var products []struct {
			Name string
			SKU  string
		}
		r.db.WithContext(ctx).Table("products").
			Select("name, sku").
			Where("deleted_at IS NULL AND is_active = true").
			Order("name").
			Limit(10).Scan(&products)
		for _, p := range products {
			opts.Products = append(opts.Products, FormDataProduct{Name: p.Name, SKU: p.SKU})
		}

		// Fetch payment terms
		var terms []struct {
			Name string
			Code string
		}
		r.db.WithContext(ctx).Table("payment_terms").
			Select("name, code").
			Where("deleted_at IS NULL AND is_active = true").
			Order("name").Scan(&terms)
		for _, t := range terms {
			opts.PaymentTerms = append(opts.PaymentTerms, FormDataOption{Name: t.Name, Code: t.Code})
		}

		// Fetch business units
		var units []struct{ Name string }
		r.db.WithContext(ctx).Table("business_units").
			Select("name").
			Where("deleted_at IS NULL AND is_active = true").
			Order("name").Scan(&units)
		for _, bu := range units {
			opts.BusinessUnits = append(opts.BusinessUnits, FormDataOption{Name: bu.Name})
		}
	}

	return opts
}
