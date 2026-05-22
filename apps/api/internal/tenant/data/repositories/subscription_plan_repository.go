package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SubscriptionPlanRepository defines data-access operations for SubscriptionPlanConfig.
type SubscriptionPlanRepository interface {
	// ListActive returns all active plan configs, ordered by sort_order.
	ListActive(ctx context.Context) ([]*models.SubscriptionPlanConfig, error)
	// ListAll returns all plan configs (including inactive) for system admins.
	ListAll(ctx context.Context, page, perPage int) ([]*models.SubscriptionPlanConfig, int64, error)
	// FindBySlug returns a single plan config by its slug.
	FindBySlug(ctx context.Context, slug string) (*models.SubscriptionPlanConfig, error)
	// Upsert inserts or updates a plan config (keyed on slug).
	Upsert(ctx context.Context, plan *models.SubscriptionPlanConfig) error
	// SetActive toggles a plan's active state.
	SetActive(ctx context.Context, slug string, active bool) error
	// ListEntitlements returns all module entitlements for a given plan slug.
	ListEntitlements(ctx context.Context, planSlug string) ([]*models.PlanModuleEntitlement, error)
	// SyncEntitlements replaces all entitlements for a plan with the provided list.
	SyncEntitlements(ctx context.Context, planSlug string, modules []string) error
	// ListMenuEntitlements returns all enabled menu URLs for a given plan slug.
	ListMenuEntitlements(ctx context.Context, planSlug string) ([]string, error)
	// SyncMenuEntitlements replaces all menu entitlement rows for a plan.
	SyncMenuEntitlements(ctx context.Context, planSlug string, menuURLs []string) error
	// GetEnabledModules returns the set of enabled module slugs for a plan slug.
	GetEnabledModules(ctx context.Context, planSlug string) ([]string, error)
	// GetEnabledModulesForTenant looks up the tenant's active subscription plan and
	// returns the enabled module slugs for that plan. Falls back to empty slice on error.
	GetEnabledModulesForTenant(ctx context.Context, tenantID string) ([]string, error)
}

type subscriptionPlanRepository struct {
	db *gorm.DB
}

var fallbackPlanModules = map[string][]string{
	"pos_growth":     {"pos", "purchase", "sales", "inventory", "finance", "core"},
	"erp_pro":        {"purchase", "inventory", "finance", "sales", "core"},
	"crm_growth":     {"crm", "core"},
	"hr_growth":      {"hr", "core"},
	"growth_suite":   {"pos", "purchase", "inventory", "finance", "sales", "crm", "core"},
	"ultimate_suite": {"pos", "purchase", "inventory", "finance", "sales", "crm", "hr", "core"},
	"enterprise":     {"pos", "purchase", "inventory", "finance", "sales", "crm", "hr", "core"},
}

func normalizePlanSlug(slug string) string {
	normalized := strings.ToLower(strings.TrimSpace(slug))
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
	case "full_access":
		return "ultimate_suite"
	default:
		return normalized
	}
}

func fallbackModulesForPlan(slug string) []string {
	mods, ok := fallbackPlanModules[normalizePlanSlug(slug)]
	if !ok {
		return nil
	}
	out := make([]string, len(mods))
	copy(out, mods)
	return out
}

func deriveModuleFromMenuURL(menuURL string) string {
	path := strings.ToLower(strings.TrimSpace(menuURL))
	switch {
	case strings.HasPrefix(path, "/dashboard"):
		return "core"
	case strings.HasPrefix(path, "/pos"):
		return "pos"
	case strings.HasPrefix(path, "/sales"):
		return "sales"
	case strings.HasPrefix(path, "/purchase"):
		return "purchase"
	case strings.HasPrefix(path, "/stock"):
		return "inventory"
	case strings.HasPrefix(path, "/finance"):
		return "finance"
	case strings.HasPrefix(path, "/crm"):
		return "crm"
	case strings.HasPrefix(path, "/hrd"):
		return "hr"
	case strings.HasPrefix(path, "/master-data"):
		return "core"
	default:
		return ""
	}
}

func DeriveModulesFromMenuURLs(menuURLs []string) []string {
	seen := make(map[string]struct{})
	modules := make([]string, 0)
	for _, menuURL := range menuURLs {
		module := deriveModuleFromMenuURL(menuURL)
		if module == "" {
			continue
		}
		if _, exists := seen[module]; exists {
			continue
		}
		seen[module] = struct{}{}
		modules = append(modules, module)
	}
	return modules
}

// NewSubscriptionPlanRepository creates a new SubscriptionPlanRepository.
func NewSubscriptionPlanRepository(db *gorm.DB) SubscriptionPlanRepository {
	return &subscriptionPlanRepository{db: db}
}

// getDB returns the db scoped to the context (respects transaction if present).
// Plan configs are platform-wide, so we bypass tenant scoping intentionally.
func (r *subscriptionPlanRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := database.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *subscriptionPlanRepository) ListActive(ctx context.Context) ([]*models.SubscriptionPlanConfig, error) {
	var plans []*models.SubscriptionPlanConfig
	err := r.getDB(ctx).
		Preload("Entitlements").
		Preload("PermissionEntitlements").
		Where("is_active = true AND deleted_at IS NULL").
		Order("sort_order ASC, name ASC").
		Find(&plans).Error
	return plans, err
}

func (r *subscriptionPlanRepository) ListAll(ctx context.Context, page, perPage int) ([]*models.SubscriptionPlanConfig, int64, error) {
	var plans []*models.SubscriptionPlanConfig
	var total int64

	q := r.getDB(ctx).Model(&models.SubscriptionPlanConfig{}).Where("deleted_at IS NULL")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := q.Preload("Entitlements").
		Preload("PermissionEntitlements").
		Order("sort_order ASC, name ASC").
		Offset(offset).Limit(perPage).
		Find(&plans).Error
	return plans, total, err
}

func (r *subscriptionPlanRepository) FindBySlug(ctx context.Context, slug string) (*models.SubscriptionPlanConfig, error) {
	var plan models.SubscriptionPlanConfig
	err := r.getDB(ctx).
		Preload("Entitlements").
		Preload("PermissionEntitlements").
		Where("slug = ? AND deleted_at IS NULL", slug).
		First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionPlanRepository) Upsert(ctx context.Context, plan *models.SubscriptionPlanConfig) error {
	values := map[string]any{
		"slug":                     plan.Slug,
		"name":                     plan.Name,
		"category":                 plan.Category,
		"description":              plan.Description,
		"billing_type":             plan.BillingType,
		"price_monthly_idr":        plan.PriceMonthlyIDR,
		"price_yearly_idr":         plan.PriceYearlyIDR,
		"outlet_addon_monthly_idr": plan.OutletAddonMonthlyIDR,
		"outlet_addon_yearly_idr":  plan.OutletAddonYearlyIDR,
		"min_users":                plan.MinUsers,
		"max_users":                plan.MaxUsers,
		"is_active":                plan.IsActive,
		"is_highlighted":           plan.IsHighlighted,
		"sort_order":               plan.SortOrder,
		"features":                 plan.Features,
		"role_templates":           plan.RoleTemplates,
	}

	return r.getDB(ctx).
		Table("subscription_plan_configs").
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.Assignments(map[string]any{
				"name":                     plan.Name,
				"category":                 plan.Category,
				"description":              plan.Description,
				"billing_type":             plan.BillingType,
				"price_monthly_idr":        plan.PriceMonthlyIDR,
				"price_yearly_idr":         plan.PriceYearlyIDR,
				"outlet_addon_monthly_idr": plan.OutletAddonMonthlyIDR,
				"outlet_addon_yearly_idr":  plan.OutletAddonYearlyIDR,
				"min_users":                plan.MinUsers,
				"max_users":                plan.MaxUsers,
				"is_active":                plan.IsActive,
				"is_highlighted":           plan.IsHighlighted,
				"sort_order":               plan.SortOrder,
				"features":                 plan.Features,
				"role_templates":           plan.RoleTemplates,
				"updated_at":               gorm.Expr("NOW()"),
			}),
		}).
		Create(values).Error
}

func (r *subscriptionPlanRepository) SetActive(ctx context.Context, slug string, active bool) error {
	return r.getDB(ctx).
		Model(&models.SubscriptionPlanConfig{}).
		Where("slug = ? AND deleted_at IS NULL", slug).
		Update("is_active", active).Error
}

func (r *subscriptionPlanRepository) ListEntitlements(ctx context.Context, planSlug string) ([]*models.PlanModuleEntitlement, error) {
	var entitlements []*models.PlanModuleEntitlement
	err := r.getDB(ctx).
		Where("plan_slug = ? AND is_enabled = true", planSlug).
		Find(&entitlements).Error
	return entitlements, err
}

// SyncEntitlements atomically replaces all enabled entitlements for a plan.
// Modules not in the list are soft-deleted (is_enabled = false).
func (r *subscriptionPlanRepository) SyncEntitlements(ctx context.Context, planSlug string, modules []string) error {
	db := r.getDB(ctx)

	// Disable all existing entitlements for this plan.
	if err := db.Model(&models.PlanModuleEntitlement{}).
		Where("plan_slug = ?", planSlug).
		Update("is_enabled", false).Error; err != nil {
		return err
	}

	// Re-enable or create each provided module.
	for _, mod := range modules {
		entitlement := models.PlanModuleEntitlement{
			PlanSlug:   planSlug,
			ModuleSlug: mod,
			IsEnabled:  true,
		}
		if err := db.Where(models.PlanModuleEntitlement{PlanSlug: planSlug, ModuleSlug: mod}).
			Assign(entitlement).
			FirstOrCreate(&entitlement).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *subscriptionPlanRepository) GetEnabledModules(ctx context.Context, planSlug string) ([]string, error) {
	normalizedPlanSlug := normalizePlanSlug(planSlug)
	entitlements, err := r.ListEntitlements(ctx, normalizedPlanSlug)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(entitlements))
	modules := make([]string, 0, len(entitlements))
	for _, e := range entitlements {
		if e.IsEnabled {
			module := strings.ToLower(strings.TrimSpace(e.ModuleSlug))
			if module == "" {
				continue
			}
			if _, exists := seen[module]; exists {
				continue
			}
			seen[module] = struct{}{}
			modules = append(modules, module)
		}
	}
	menuURLs, menuErr := r.ListMenuEntitlements(ctx, normalizedPlanSlug)
	if menuErr != nil {
		return nil, menuErr
	}
	for _, module := range DeriveModulesFromMenuURLs(menuURLs) {
		if _, exists := seen[module]; exists {
			continue
		}
		seen[module] = struct{}{}
		modules = append(modules, module)
	}
	if len(modules) == 0 {
		if fallback := fallbackModulesForPlan(normalizedPlanSlug); len(fallback) > 0 {
			return fallback, nil
		}
	}
	return modules, nil
}

// GetEnabledModulesForTenant resolves the tenant's active plan and returns enabled modules.
// Uses a direct DB query to avoid a separate repo dependency.
func (r *subscriptionPlanRepository) GetEnabledModulesForTenant(ctx context.Context, tenantID string) ([]string, error) {
	// Find the active plan slug for the tenant.
	var planSlug string
	err := r.db.WithContext(ctx).
		Table("tenant_subscriptions").
		Select("plan").
		Where("tenant_id = ? AND status IN ('active','trial') AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(1).
		Row().Scan(&planSlug)
	if err != nil || planSlug == "" {
		// Tenant has no active subscription — grant no module access.
		return []string{}, nil
	}
	return r.GetEnabledModules(ctx, normalizePlanSlug(planSlug))
}

func (r *subscriptionPlanRepository) ListMenuEntitlements(ctx context.Context, planSlug string) ([]string, error) {
	rows := make([]models.PlanPermissionEntitlement, 0)
	err := r.getDB(ctx).
		Where("plan_slug = ? AND is_enabled = true AND permission_code = '' AND menu_url <> ''", normalizePlanSlug(planSlug)).
		Order("menu_url ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	menuURLs := make([]string, 0, len(rows))
	for _, row := range rows {
		menuURLs = append(menuURLs, strings.ToLower(strings.TrimSpace(row.MenuURL)))
	}
	return menuURLs, nil
}

func (r *subscriptionPlanRepository) SyncMenuEntitlements(ctx context.Context, planSlug string, menuURLs []string) error {
	db := r.getDB(ctx)
	normalizedPlanSlug := normalizePlanSlug(planSlug)

	if err := db.Where("plan_slug = ? AND permission_code = ''", normalizedPlanSlug).Delete(&models.PlanPermissionEntitlement{}).Error; err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(menuURLs))
	for _, menuURL := range menuURLs {
		normalizedMenuURL := strings.ToLower(strings.TrimSpace(menuURL))
		if normalizedMenuURL == "" {
			continue
		}
		if _, exists := seen[normalizedMenuURL]; exists {
			continue
		}
		seen[normalizedMenuURL] = struct{}{}
		entitlement := models.PlanPermissionEntitlement{
			PlanSlug:       normalizedPlanSlug,
			PermissionCode: "",
			MenuURL:        normalizedMenuURL,
			IsEnabled:      true,
		}
		if err := db.Create(&entitlement).Error; err != nil {
			return err
		}
	}

	return nil
}
