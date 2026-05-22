package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"strings"

	"github.com/gilabs/gims/api/internal/supplier/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SupplierTypeRepository defines the interface for supplier type data access
type SupplierTypeRepository interface {
	Create(ctx context.Context, supplierType *models.SupplierType) error
	FindByID(ctx context.Context, id string) (*models.SupplierType, error)
	List(ctx context.Context, params ListParams) ([]models.SupplierType, int64, error)
	Update(ctx context.Context, supplierType *models.SupplierType) error
	Delete(ctx context.Context, id string) error
}

type supplierTypeRepository struct {
	db *gorm.DB
}

// NewSupplierTypeRepository creates a new instance of SupplierTypeRepository
func NewSupplierTypeRepository(db *gorm.DB) SupplierTypeRepository {
	return &supplierTypeRepository{db: db}
}

func (r *supplierTypeRepository) Create(ctx context.Context, supplierType *models.SupplierType) error {
	return database.GetDB(ctx, r.db).Create(supplierType).Error
}

func (r *supplierTypeRepository) FindByID(ctx context.Context, id string) (*models.SupplierType, error) {
	var supplierType models.SupplierType
	err := database.GetDB(ctx, r.db).First(&supplierType, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &supplierType, nil
}

func (r *supplierTypeRepository) List(ctx context.Context, params ListParams) ([]models.SupplierType, int64, error) {
	var supplierTypes []models.SupplierType
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.SupplierType{})

	// Apply search filter
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	if params.ActiveOnly {
		query = query.Where("is_active = ?", true)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting with whitelist and clause builder to prevent SQL injection
	allowedSortColumns := map[string]string{
		"name":        "name",
		"description": "description",
		"is_active":   "is_active",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
	}

	sortBy := allowedSortColumns[strings.ToLower(strings.TrimSpace(params.SortBy))]
	if sortBy == "" {
		sortBy = "name"
	}

	isDesc := strings.ToLower(strings.TrimSpace(params.SortDir)) == "desc"
	query = query.Order("is_active DESC").Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&supplierTypes).Error; err != nil {
		return nil, 0, err
	}

	return supplierTypes, total, nil
}

func (r *supplierTypeRepository) Update(ctx context.Context, supplierType *models.SupplierType) error {
	return database.GetDB(ctx, r.db).Save(supplierType).Error
}

func (r *supplierTypeRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.SupplierType{}, "id = ?", id).Error
}
