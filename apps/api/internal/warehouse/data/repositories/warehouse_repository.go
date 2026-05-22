package repositories

import (
	"context"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/warehouse/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

// WarehouseRepository handles database operations for warehouses
type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *models.Warehouse) error
	GetByID(ctx context.Context, id string) (*models.Warehouse, error)
	GetByCode(ctx context.Context, code string) (*models.Warehouse, error)
	GetNextCode(ctx context.Context) (string, error)
	List(ctx context.Context, params WarehouseListParams) ([]*models.Warehouse, int64, error)
	Update(ctx context.Context, warehouse *models.Warehouse) error
	Delete(ctx context.Context, id string) error
	// HasActiveStock returns true when the warehouse has any inventory_batches
	// with current_quantity > 0. Used to block deletes that would orphan stock.
	HasActiveStock(ctx context.Context, warehouseID string) (bool, error)
	FindByOutletIDs(ctx context.Context, outletIDs []string) ([]*models.Warehouse, error)
	UpdateIsActiveByOutletIDs(ctx context.Context, outletIDs []string, isActive bool) error
}

// WarehouseListParams defines parameters for listing warehouses
type WarehouseListParams struct {
	ListParams
	IsActive *bool
}

// ListParams defines common list parameters
type ListParams struct {
	Search  string
	SortBy  string
	SortDir string
	Limit   int
	Offset  int
}

type warehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository creates a new warehouse repository
func NewWarehouseRepository(db *gorm.DB) WarehouseRepository {
	return &warehouseRepository{db: db}
}

// Create creates a new warehouse
func (r *warehouseRepository) Create(ctx context.Context, warehouse *models.Warehouse) error {
	return database.GetDB(ctx, r.db).Create(warehouse).Error
}

// GetByID retrieves a warehouse by ID with preloaded relations
func (r *warehouseRepository) GetByID(ctx context.Context, id string) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := database.GetDB(ctx, r.db).
		Select("warehouses.*, EXISTS(SELECT 1 FROM inventory_batches WHERE warehouse_id = warehouses.id AND current_quantity > 0 AND deleted_at IS NULL) AS has_stock").
		Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Village.District.City.Province").
		First(&warehouse, "warehouses.id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// GetNextCode generates the next sequential warehouse code in WH-XXXXX format.
// It counts all warehouses (including soft-deleted) to guarantee uniqueness.
func (r *warehouseRepository) GetNextCode(ctx context.Context) (string, error) {
	globalCtx := context.WithValue(ctx, middleware.IsSystemAdminKey, true)

	var count int64
	if err := r.db.WithContext(globalCtx).Unscoped().Model(&models.Warehouse{}).Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("WH-%05d", count+1), nil
}

// GetByCode retrieves a warehouse by code
func (r *warehouseRepository) GetByCode(ctx context.Context, code string) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := database.GetDB(ctx, r.db).
		First(&warehouse, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// List retrieves warehouses with pagination and filtering
func (r *warehouseRepository) List(ctx context.Context, params WarehouseListParams) ([]*models.Warehouse, int64, error) {
	var warehouses []*models.Warehouse
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.Warehouse{}).
		Select("warehouses.*, EXISTS(SELECT 1 FROM inventory_batches WHERE warehouse_id = warehouses.id AND current_quantity > 0 AND deleted_at IS NULL) AS has_stock")

	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		WarehouseIDColumn: "id",
		OutletIDColumn:    "outlet_id",
	})

	// Apply filters
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR address ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting with whitelist and clause builder to prevent SQL injection
	allowedSortColumns := map[string]string{
		"name":       "name",
		"code":       "code",
		"address":    "address",
		"is_active":  "is_active",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}

	sortBy := allowedSortColumns[strings.ToLower(strings.TrimSpace(params.SortBy))]
	if sortBy == "" {
		sortBy = "name"
	}

	isDesc := strings.ToLower(strings.TrimSpace(params.SortDir)) == "desc"

	query = query.Order("is_active DESC").Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit).Offset(params.Offset)
	}

	// Preload relations
	query = query.Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Village.District.City.Province")

	if err := query.Find(&warehouses).Error; err != nil {
		return nil, 0, err
	}

	return warehouses, total, nil
}

// Update updates an existing warehouse
func (r *warehouseRepository) Update(ctx context.Context, warehouse *models.Warehouse) error {
	return database.GetDB(ctx, r.db).Save(warehouse).Error
}

// Delete soft deletes a warehouse
func (r *warehouseRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.Warehouse{}, "id = ?", id).Error
}

// HasActiveStock checks whether any active inventory batches exist for the given warehouse.
func (r *warehouseRepository) HasActiveStock(ctx context.Context, warehouseID string) (bool, error) {
	var count int64
	err := database.GetDB(ctx, r.db).
		Table("inventory_batches").
		Where("warehouse_id = ? AND current_quantity > 0 AND deleted_at IS NULL", warehouseID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindByOutletIDs retrieves all warehouses linked to the specified outlet IDs
func (r *warehouseRepository) FindByOutletIDs(ctx context.Context, outletIDs []string) ([]*models.Warehouse, error) {
	var warehouses []*models.Warehouse
	err := database.GetDB(ctx, r.db).
		Where("outlet_id IN ?", outletIDs).
		Find(&warehouses).Error
	if err != nil {
		return nil, err
	}
	return warehouses, nil
}

// UpdateIsActiveByOutletIDs updates is_active status for all warehouses linked to specified outlet IDs
func (r *warehouseRepository) UpdateIsActiveByOutletIDs(ctx context.Context, outletIDs []string, isActive bool) error {
	return database.GetDB(ctx, r.db).
		Model(&models.Warehouse{}).
		Where("outlet_id IN ?", outletIDs).
		Update("is_active", isActive).Error
}
