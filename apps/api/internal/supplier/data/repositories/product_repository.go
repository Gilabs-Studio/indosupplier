package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*models.SupplierProduct, error)
	List(ctx context.Context, supplierProfileID string, search string, categoryID string, page int, perPage int) ([]models.SupplierProduct, int64, error)
	Create(ctx context.Context, p *models.SupplierProduct) error
	Update(ctx context.Context, p *models.SupplierProduct) error
	Delete(ctx context.Context, id string) error

	// Categories
	ListCategories(ctx context.Context) ([]models.Category, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*models.SupplierProduct, error) {
	var p models.SupplierProduct
	if err := r.getDB(ctx).
		Preload("Photos").
		Preload("Category").
		Where("id = ?", id).
		First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) List(ctx context.Context, supplierProfileID string, search string, categoryID string, page int, perPage int) ([]models.SupplierProduct, int64, error) {
	var products []models.SupplierProduct
	var total int64

	query := r.getDB(ctx).Model(&models.SupplierProduct{}).
		Where("supplier_profile_id = ?", supplierProfileID)

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Preload Category and Photos (GORM will fetch them efficiently in a single bulk query per table to prevent N+1)
	err := query.
		Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, created_at ASC")
		}).
		Preload("Category").
		Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&products).Error

	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Create(ctx context.Context, p *models.SupplierProduct) error {
	// GORM will automatically insert associated photos in the same transaction
	return r.getDB(ctx).Create(p).Error
}

func (r *productRepository) Update(ctx context.Context, p *models.SupplierProduct) error {
	return database.RetryTx(r.db, func(tx *gorm.DB) error {
		// 1. Delete existing photos associated with this product
		if err := tx.Where("supplier_product_id = ?", p.ID).Delete(&models.SupplierProductPhoto{}).Error; err != nil {
			return err
		}

		// 2. Save product updates (this will also re-insert the photos in the GORM struct if they are populated)
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(p).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	// Let GORM delete cascade the photos through constraint, but let's delete them explicitly to be safe
	return database.RetryTx(r.db, func(tx *gorm.DB) error {
		if err := tx.Where("supplier_product_id = ?", id).Delete(&models.SupplierProductPhoto{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.SupplierProduct{}).Error
	})
}

func (r *productRepository) ListCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := r.getDB(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error
	return categories, err
}
