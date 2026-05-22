package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductCategoryRepository defines the interface for product category data access
type ProductCategoryRepository interface {
	Create(ctx context.Context, category *models.ProductCategory) error
	FindByID(ctx context.Context, id string) (*models.ProductCategory, error)
	List(ctx context.Context, params ListParams) ([]models.ProductCategory, int64, error)
	Update(ctx context.Context, category *models.ProductCategory) error
	Delete(ctx context.Context, id string) error
	// Tree methods
	GetRootCategories(ctx context.Context, onlyActive bool) ([]models.ProductCategory, error)
	GetChildrenByParentID(ctx context.Context, parentID string, onlyActive bool) ([]models.ProductCategory, error)
	HasChildren(ctx context.Context, categoryID string) (bool, error)
	CountProductsByCategory(ctx context.Context, categoryID string) (int64, error)
}

type productCategoryRepository struct {
	db *gorm.DB
}

// NewProductCategoryRepository creates a new instance of ProductCategoryRepository
func NewProductCategoryRepository(db *gorm.DB) ProductCategoryRepository {
	return &productCategoryRepository{db: db}
}

func (r *productCategoryRepository) Create(ctx context.Context, category *models.ProductCategory) error {
	return database.GetDB(ctx, r.db).Create(category).Error
}

func (r *productCategoryRepository) FindByID(ctx context.Context, id string) (*models.ProductCategory, error) {
	var category models.ProductCategory
	err := database.GetDB(ctx, r.db).Preload("Parent").First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *productCategoryRepository) List(ctx context.Context, params ListParams) ([]models.ProductCategory, int64, error) {
	var categories []models.ProductCategory
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.ProductCategory{})

	// Apply search filter
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	query = query.Order("is_active DESC")
	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: params.SortBy},
			Desc:   params.SortDir == "desc",
		})
	} else {
		query = query.Order("name ASC")
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Preload parent for display
	query = query.Preload("Parent")

	if err := query.Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *productCategoryRepository) Update(ctx context.Context, category *models.ProductCategory) error {
	return database.GetDB(ctx, r.db).Save(category).Error
}

func (r *productCategoryRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.ProductCategory{}, "id = ?", id).Error
}

// GetRootCategories returns all categories without a parent (root level)
func (r *productCategoryRepository) GetRootCategories(ctx context.Context, onlyActive bool) ([]models.ProductCategory, error) {
	var categories []models.ProductCategory
	query := database.GetDB(ctx, r.db).Where("parent_id IS NULL")
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	query = query.Order("name ASC")

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// GetChildrenByParentID returns all direct children of a category
func (r *productCategoryRepository) GetChildrenByParentID(ctx context.Context, parentID string, onlyActive bool) ([]models.ProductCategory, error) {
	var categories []models.ProductCategory
	query := database.GetDB(ctx, r.db).Where("parent_id = ?", parentID)
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	query = query.Order("name ASC")

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// HasChildren checks if a category has any children
func (r *productCategoryRepository) HasChildren(ctx context.Context, categoryID string) (bool, error) {
	var count int64
	err := database.GetDB(ctx, r.db).Model(&models.ProductCategory{}).Where("parent_id = ?", categoryID).Count(&count).Error
	return count > 0, err
}

// CountProductsByCategory counts products in a specific category
func (r *productCategoryRepository) CountProductsByCategory(ctx context.Context, categoryID string) (int64, error) {
	var count int64
	err := database.GetDB(ctx, r.db).Model(&models.Product{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

