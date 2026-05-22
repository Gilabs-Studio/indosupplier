package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductListParams extends ListParams with product-specific filters
type ProductListParams struct {
	ListParams
	CategoryID        string
	BrandID           string
	SegmentID         string
	TypeID            string
	SupplierID        string
	Status            string
	IsApproved        *bool
	ProductKind       string
	POSScope          string
	IsIngredient      *bool
	OutletID          string
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	FindByID(ctx context.Context, id string) (*models.Product, error)
	List(ctx context.Context, params ProductListParams) ([]models.Product, int64, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id string) error
}

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new instance of ProductRepository
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	return database.GetDB(ctx, r.db).Create(product).Error
}

// stockSubquery is the SQL fragment that replaces the stale products.current_stock
// and products.reserved_stock columns with live aggregates from inventory_batches.
const stockSubquery = `products.*,
	COALESCE((SELECT SUM(ib.current_quantity)  FROM inventory_batches ib WHERE ib.product_id = products.id AND ib.tenant_id = products.tenant_id AND ib.deleted_at IS NULL), 0) AS current_stock,
	COALESCE((SELECT SUM(ib.reserved_quantity) FROM inventory_batches ib WHERE ib.product_id = products.id AND ib.tenant_id = products.tenant_id AND ib.deleted_at IS NULL), 0) AS reserved_stock`

func (r *productRepository) FindByID(ctx context.Context, id string) (*models.Product, error) {
	var product models.Product
	tenantDB := database.GetDB(ctx, r.db)
	hasProductOutlets := tenantDB.Migrator().HasTable(&models.ProductOutlet{})

	query := tenantDB.
		Select(stockSubquery).
		Preload("Category").
		Preload("Brand").
		Preload("Segment").
		Preload("Type").
		Preload("Uom").
		Preload("PurchaseUom").
		Preload("Packaging").
		Preload("ProcurementType").
		Preload("Supplier").
		Preload("BusinessUnit").
		Preload("RecipeItems", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("RecipeItems.IngredientProduct", func(db *gorm.DB) *gorm.DB {
			return db.Select(stockSubquery)
		}).
		Preload("RecipeItems.Uom")

	if hasProductOutlets {
		query = query.Preload("Outlets")
	}

	err := query.First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) List(ctx context.Context, params ProductListParams) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	tenantDB := database.GetDB(ctx, r.db)
	hasProductOutlets := tenantDB.Migrator().HasTable(&models.ProductOutlet{})

	query := tenantDB.Model(&models.Product{}).Select(stockSubquery)

	// Apply search filter
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR barcode ILIKE ? OR sku ILIKE ?", search, search, search, search)
	}

	// Apply filters
	if params.CategoryID != "" {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if params.BrandID != "" {
		query = query.Where("brand_id = ?", params.BrandID)
	}
	if params.SegmentID != "" {
		query = query.Where("segment_id = ?", params.SegmentID)
	}
	if params.TypeID != "" {
		query = query.Where("type_id = ?", params.TypeID)
	}
	if params.SupplierID != "" {
		query = query.Where("supplier_id = ?", params.SupplierID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.IsApproved != nil {
		query = query.Where("is_approved = ?", *params.IsApproved)
	}
	if params.ProductKind != "" {
		query = query.Where("product_kind = ?", params.ProductKind)
	}
	if params.POSScope != "" {
		query = query.Where("pos_scope = ?", params.POSScope)
	}
	if params.IsIngredient != nil {
		query = query.Where("is_ingredient = ?", *params.IsIngredient)
	}
	if params.OutletID != "" && hasProductOutlets {
		query = query.Where("id IN (SELECT product_id FROM product_outlets WHERE outlet_id = ? AND deleted_at IS NULL)", params.OutletID)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	// For computed stock columns we must use the full subquery expression in ORDER BY
	// to avoid PostgreSQL "ambiguous" error (products.* already contains current_stock column).
	const currentStockExpr = "(SELECT COALESCE(SUM(ib.current_quantity), 0) FROM inventory_batches ib WHERE ib.product_id = products.id AND ib.tenant_id = products.tenant_id AND ib.deleted_at IS NULL)"
	const reservedStockExpr = "(SELECT COALESCE(SUM(ib.reserved_quantity), 0) FROM inventory_batches ib WHERE ib.product_id = products.id AND ib.tenant_id = products.tenant_id AND ib.deleted_at IS NULL)"

	isDesc := params.SortDir == "desc"

	switch params.SortBy {
	case "current_stock":
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Raw: true, Name: currentStockExpr},
			Desc:   isDesc,
		})
	case "reserved_stock":
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Raw: true, Name: reservedStockExpr},
			Desc:   isDesc,
		})
	case "":
		query = query.Order("name ASC")
	default:
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: params.SortBy},
			Desc:   isDesc,
		})
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Preload relations
	query = query.
		Preload("Category").
		Preload("Brand").
		Preload("Segment").
		Preload("Type").
		Preload("Uom").
		Preload("PurchaseUom").
		Preload("Packaging").
		Preload("ProcurementType").
		Preload("Supplier").
		Preload("BusinessUnit").
		Preload("RecipeItems", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("RecipeItems.IngredientProduct", func(db *gorm.DB) *gorm.DB {
			return db.Select(stockSubquery)
		}).
		Preload("RecipeItems.Uom")

	if hasProductOutlets {
		query = query.Preload("Outlets")
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	return database.GetDB(ctx, r.db).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.Product{}, "id = ?", id).Error
}
