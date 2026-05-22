package repositories

import (
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"gorm.io/gorm"
)

// TenantWithCounts is a Tenant enriched with derived fields for system admin views.
type TenantWithCounts struct {
	models.Tenant
	CurrentUsers   int64
	OwnerName      string
	OwnerEmail     string
	CompanyCount   int64
	OutletCount    int64
	WarehouseCount int64
}

// TenantRepository provides access to the tenants table
type TenantRepository interface {
	FindByID(ctx context.Context, id string) (*models.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	FindAll(ctx context.Context) ([]models.Tenant, error)
	FindPaginated(ctx context.Context, search string, page, perPage int) ([]models.Tenant, int64, error)
	FindPaginatedWithCounts(ctx context.Context, search string, page, perPage int) ([]TenantWithCounts, int64, error)
	FindByIDWithCounts(ctx context.Context, id string) (*TenantWithCounts, error)
	CountActiveUsers(ctx context.Context, tenantID string) (int64, error)
	Create(ctx context.Context, tenant *models.Tenant) error
	Update(ctx context.Context, tenant *models.Tenant) error
}

type tenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository creates a new TenantRepository
func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) getDB(ctx context.Context) *gorm.DB {
	// The tenants table is platform-scoped and does not contain tenant_id.
	// Use the raw context-bound DB so GORM does not inject tenant scoping.
	return r.db.WithContext(ctx)
}

func (r *tenantRepository) FindByID(ctx context.Context, id string) (*models.Tenant, error) {
	var tenant models.Tenant
	// tenants is a platform-level table (no tenant_id column), so use global DB.
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	return &tenant, err
}

func (r *tenantRepository) FindBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	// tenants is a platform-level table (no tenant_id column), so use global DB.
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error
	return &tenant, err
}

func (r *tenantRepository) FindAll(ctx context.Context) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := r.getDB(ctx).Find(&tenants).Error
	return tenants, err
}

func (r *tenantRepository) FindPaginated(ctx context.Context, search string, page, perPage int) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64

	query := r.getDB(ctx)
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", like, like)
	}

	if err := query.Model(&models.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

// FindPaginatedWithCounts fetches tenants with current user count and owner info.
// Uses r.db directly (not getDB) because the tenants table has no tenant_id column —
// this is a platform-wide query executed by system admins only.
func (r *tenantRepository) FindPaginatedWithCounts(ctx context.Context, search string, page, perPage int) ([]TenantWithCounts, int64, error) {
	return r.queryWithCounts(ctx, "", search, page, perPage)
}

// FindByIDWithCounts fetches a single tenant with current user count and owner info.
func (r *tenantRepository) FindByIDWithCounts(ctx context.Context, id string) (*TenantWithCounts, error) {
	rows, _, err := r.queryWithCounts(ctx, id, "", 1, 1)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &rows[0], nil
}

func (r *tenantRepository) CountActiveUsers(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("users").
		Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", tenantID, "active").
		Count(&count).Error
	return count, err
}

func (r *tenantRepository) queryWithCounts(ctx context.Context, id, search string, page, perPage int) ([]TenantWithCounts, int64, error) {
	type row struct {
		ID                     string     `gorm:"column:id"`
		Name                   string     `gorm:"column:name"`
		Slug                   string     `gorm:"column:slug"`
		Status                 string     `gorm:"column:status"`
		Plan                   string     `gorm:"column:plan"`
		SeatLimit              int        `gorm:"column:seat_limit"`
		OwnerUserID            *string    `gorm:"column:owner_user_id"`
		DeletionRequestedAt    *time.Time `gorm:"column:deletion_requested_at"`
		DeletionScheduledAt    *time.Time `gorm:"column:deletion_scheduled_at"`
		DeletionRequestedBy    *string    `gorm:"column:deletion_requested_by"`
		DeletionPreviousStatus *string    `gorm:"column:deletion_previous_status"`
		CurrentUsers           int64      `gorm:"column:current_users"`
		OwnerName              string     `gorm:"column:owner_name"`
		OwnerEmail             string     `gorm:"column:owner_email"`
		CompanyCount           int64      `gorm:"column:company_count"`
		OutletCount            int64      `gorm:"column:outlet_count"`
		WarehouseCount         int64      `gorm:"column:warehouse_count"`
	}

	baseQ := r.db.WithContext(ctx).Table("tenants").Where("tenants.deleted_at IS NULL")
	if id != "" {
		baseQ = baseQ.Where("tenants.id = ?", id)
	}
	if search != "" {
		like := "%" + search + "%"
		baseQ = baseQ.Where("tenants.name ILIKE ? OR tenants.slug ILIKE ?", like, like)
	}

	var total int64
	if err := baseQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []row
	offset := (page - 1) * perPage
	err := baseQ.
		Select(`tenants.id, tenants.name, tenants.slug, tenants.status, tenants.plan,
			tenants.deletion_requested_at, tenants.deletion_scheduled_at, tenants.deletion_requested_by, tenants.deletion_previous_status,
			COALESCE((SELECT ts.seat_limit
				FROM tenant_subscriptions ts
				WHERE ts.tenant_id = tenants.id
				  AND ts.status IN ('active', 'trial')
				  AND ts.deleted_at IS NULL
				ORDER BY ts.created_at DESC
				LIMIT 1), tenants.max_users) AS seat_limit,
			tenants.owner_user_id,
			COALESCE((SELECT COUNT(*) FROM users WHERE users.tenant_id = tenants.id AND users.deleted_at IS NULL), 0) AS current_users,
			COALESCE((SELECT COUNT(*) FROM companies WHERE companies.tenant_id = tenants.id AND companies.deleted_at IS NULL), 0) AS company_count,
			COALESCE((SELECT COUNT(*) FROM outlets WHERE outlets.tenant_id = tenants.id AND outlets.deleted_at IS NULL), 0) AS outlet_count,
			COALESCE((SELECT COUNT(*) FROM warehouses WHERE warehouses.tenant_id = tenants.id AND warehouses.deleted_at IS NULL), 0) AS warehouse_count,
			COALESCE(u.name, '') AS owner_name,
			COALESCE(u.email, '') AS owner_email`).
		Joins("LEFT JOIN users u ON u.id = tenants.owner_user_id AND u.deleted_at IS NULL").
		Order("tenants.created_at DESC").
		Offset(offset).Limit(perPage).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]TenantWithCounts, 0, len(rows))
	for _, row := range rows {
		t := TenantWithCounts{
			CurrentUsers:   row.CurrentUsers,
			OwnerName:      row.OwnerName,
			OwnerEmail:     row.OwnerEmail,
			CompanyCount:   row.CompanyCount,
			OutletCount:    row.OutletCount,
			WarehouseCount: row.WarehouseCount,
		}
		t.ID = row.ID
		t.Name = row.Name
		t.Slug = row.Slug
		t.Status = row.Status
		t.Plan = row.Plan
		t.MaxUsers = row.SeatLimit
		t.OwnerUserID = row.OwnerUserID
		t.DeletionRequestedAt = row.DeletionRequestedAt
		t.DeletionScheduledAt = row.DeletionScheduledAt
		t.DeletionRequestedBy = row.DeletionRequestedBy
		t.DeletionPreviousStatus = row.DeletionPreviousStatus
		result = append(result, t)
	}

	return result, total, nil
}

func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	return r.getDB(ctx).Create(tenant).Error
}

func (r *tenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).
		Session(&gorm.Session{NewDB: true}).
		Table("tenants").
		Where("id = ?", tenant.ID).
		Updates(map[string]interface{}{
			"name":                     tenant.Name,
			"slug":                     tenant.Slug,
			"owner_user_id":            tenant.OwnerUserID,
			"status":                   tenant.Status,
			"plan":                     tenant.Plan,
			"max_users":                tenant.MaxUsers,
			"deletion_requested_at":    tenant.DeletionRequestedAt,
			"deletion_scheduled_at":    tenant.DeletionScheduledAt,
			"deletion_requested_by":    tenant.DeletionRequestedBy,
			"deletion_recovered_at":    tenant.DeletionRecoveredAt,
			"deletion_previous_status": tenant.DeletionPreviousStatus,
		}).Error
}
