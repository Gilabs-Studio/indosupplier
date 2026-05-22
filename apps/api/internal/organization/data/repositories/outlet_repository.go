package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OutletRepository handles database operations for outlets
type OutletRepository interface {
	Create(ctx context.Context, outlet *models.Outlet) error
	GetByID(ctx context.Context, id string) (*models.Outlet, error)
	GetByCode(ctx context.Context, code string) (*models.Outlet, error)
	GetNextCode(ctx context.Context) (string, error)
	List(ctx context.Context, params OutletListParams) ([]*models.Outlet, int64, error)
	Update(ctx context.Context, outlet *models.Outlet) error
	Delete(ctx context.Context, id string) error
	FindByCompanyID(ctx context.Context, companyID string) ([]*models.Outlet, error)
	UpdateIsActiveByCompanyID(ctx context.Context, companyID string, isActive bool) error
	// FindByWarehouseIDs returns all outlets whose warehouse_id is in the given list.
	// Used for OUTLET-scoped RBAC resolution.
	FindByWarehouseIDs(ctx context.Context, warehouseIDs []string) ([]*models.Outlet, error)
}

// OutletListParams defines parameters for listing outlets
type OutletListParams struct {
	Search      string
	SortBy      string
	SortDir     string
	Limit       int
	Offset      int
	IsActive    *bool
	CompanyID   string
	WarehouseID string // Optional: filter outlets belonging to a specific warehouse
	UserID      string // Optional: if set, filter by user-warehouse permissions (RBAC)
}

type outletRepository struct {
	db *gorm.DB
}

// NewOutletRepository creates a new outlet repository
func NewOutletRepository(db *gorm.DB) OutletRepository {
	return &outletRepository{db: db}
}

// Create creates a new outlet
func (r *outletRepository) Create(ctx context.Context, outlet *models.Outlet) error {
	return database.GetDB(ctx, r.db).Create(outlet).Error
}

// GetByID retrieves an outlet by ID with preloaded relations
func (r *outletRepository) GetByID(ctx context.Context, id string) (*models.Outlet, error) {
	var outlet models.Outlet
	err := database.GetDB(ctx, r.db).
		Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Village.District.City.Province").
		Preload("Manager").
		Preload("Company").
		First(&outlet, "outlets.id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &outlet, nil
}

// GetByCode retrieves an outlet by code
func (r *outletRepository) GetByCode(ctx context.Context, code string) (*models.Outlet, error) {
	var outlet models.Outlet
	err := database.GetDB(ctx, r.db).
		First(&outlet, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &outlet, nil
}

// GetNextCode generates the next sequential outlet code in OT-XXXXX format.
func (r *outletRepository) GetNextCode(ctx context.Context) (string, error) {
	globalCtx := context.WithValue(ctx, middleware.IsSystemAdminKey, true)

	var nextSeq int64
	if err := r.db.WithContext(globalCtx).
		Unscoped().
		Model(&models.Outlet{}).
		Select("COALESCE(MAX(CAST(SUBSTRING(code FROM 4) AS INTEGER)), 0) + 1").
		Where("code ~ '^OT-[0-9]{5,}$'").
		Scan(&nextSeq).Error; err != nil {
		return "", err
	}

	if nextSeq < 1 {
		nextSeq = 1
	}

	return fmt.Sprintf("OT-%05d", nextSeq), nil
}

// List retrieves outlets with pagination and filtering
func (r *outletRepository) List(ctx context.Context, params OutletListParams) ([]*models.Outlet, int64, error) {
	var outlets []*models.Outlet
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.Outlet{})

	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OutletIDColumn:    "id",
		WarehouseIDColumn: "warehouse_id",
	})

	// Apply search filter
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR address ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if params.CompanyID != "" {
		query = query.Where("company_id = ?", params.CompanyID)
	}

	if params.WarehouseID != "" {
		query = query.Where("warehouse_id = ?", params.WarehouseID)
	}

	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting with whitelisted columns to prevent SQL injection.
	allowedSortColumns := map[string]string{
		"name":       "name",
		"code":       "code",
		"address":    "address",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"is_active":  "is_active",
	}

	sortColumn := allowedSortColumns[strings.ToLower(strings.TrimSpace(params.SortBy))]
	if sortColumn == "" {
		sortColumn = "name"
	}

	sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
	isDesc := sortDir == "desc"

	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "is_active"}, Desc: true})
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortColumn}, Desc: isDesc})

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit).Offset(params.Offset)
	}

	// Preload relations
	query = query.
		Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Village.District.City.Province").
		Preload("Manager").
		Preload("Company")

	if err := query.Find(&outlets).Error; err != nil {
		return nil, 0, err
	}

	return outlets, total, nil
}

func contextUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value("user_id").(string)
	return value
}

// Update updates an existing outlet
func (r *outletRepository) Update(ctx context.Context, outlet *models.Outlet) error {
	return database.GetDB(ctx, r.db).Save(outlet).Error
}

// Delete soft deletes an outlet
func (r *outletRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.Outlet{}, "id = ?", id).Error
}

// FindByCompanyID retrieves all outlets for a specific company
func (r *outletRepository) FindByCompanyID(ctx context.Context, companyID string) ([]*models.Outlet, error) {
	var outlets []*models.Outlet
	err := database.GetDB(ctx, r.db).
		Where("company_id = ?", companyID).
		Preload("Company").
		Find(&outlets).Error
	if err != nil {
		return nil, err
	}
	return outlets, nil
}

// UpdateIsActiveByCompanyID updates is_active status for all outlets of a company
func (r *outletRepository) UpdateIsActiveByCompanyID(ctx context.Context, companyID string, isActive bool) error {
	return database.GetDB(ctx, r.db).
		Model(&models.Outlet{}).
		Where("company_id = ?", companyID).
		Update("is_active", isActive).Error
}

// FindByWarehouseIDs returns outlets whose warehouse_id is in the provided list.
func (r *outletRepository) FindByWarehouseIDs(ctx context.Context, warehouseIDs []string) ([]*models.Outlet, error) {
	if len(warehouseIDs) == 0 {
		return []*models.Outlet{}, nil
	}
	var outlets []*models.Outlet
	err := database.GetDB(ctx, r.db).
		Select("id, name, warehouse_id").
		Where("warehouse_id IN ? AND deleted_at IS NULL", warehouseIDs).
		Find(&outlets).Error
	return outlets, err
}
