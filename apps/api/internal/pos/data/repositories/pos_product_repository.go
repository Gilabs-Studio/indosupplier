package repositories

import (
	"context"
	"math"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"gorm.io/gorm"

	inventoryModels "github.com/gilabs/gims/api/internal/inventory/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
)

// POSCatalogProduct holds product info plus computed available stock for POS display
type POSCatalogProduct struct {
	ProductID   string
	ProductCode string
	ProductName string
	ProductKind string
	Price       float64
	Stock       float64
	ImageURL    string
	Category    string
	IsAvailable bool
}

// POSProductRepository provides product data access needed by the POS domain
type POSProductRepository interface {
	FindByID(ctx context.Context, id string) (*productModels.Product, error)
	FindByIDWithRecipe(ctx context.Context, id string) (*productModels.Product, error)
	FindPOSAvailable(ctx context.Context, warehouseID string, outletID string) ([]POSCatalogProduct, error)
}

type posProductRepository struct {
	db *gorm.DB
}

// NewPOSProductRepository creates the concrete implementation
func NewPOSProductRepository(db *gorm.DB) POSProductRepository {
	return &posProductRepository{db: db}
}

func (r *posProductRepository) FindByID(ctx context.Context, id string) (*productModels.Product, error) {
	var product productModels.Product
	err := database.GetDB(ctx, r.db).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *posProductRepository) FindByIDWithRecipe(ctx context.Context, id string) (*productModels.Product, error) {
	var product productModels.Product
	err := database.GetDB(ctx, r.db).
		Preload("RecipeItems").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindPOSAvailable returns products visible in the POS catalog for the given outlet + warehouse.
// - pos_scope='global': visible in all outlets
// - pos_scope='specific': visible only if outlet is registered in product_outlets
// When warehouseID is empty, stock values are 0 (catalog only, no stock info).
func (r *posProductRepository) FindPOSAvailable(ctx context.Context, warehouseID string, outletID string) ([]POSCatalogProduct, error) {
	var products []productModels.Product

	db := database.GetDB(ctx, r.db).
		Preload("Category").
		Preload("RecipeItems").
		Where("deleted_at IS NULL AND status = ? AND is_approved = ? AND is_active = ? AND pos_scope IN ('global', 'specific')",
			productModels.ProductStatusApproved, true, true)

	if outletID != "" {
		// Include global-scope products OR specific-scope products assigned to this outlet
		db = db.Where(
			"pos_scope = 'global' OR EXISTS (SELECT 1 FROM product_outlets WHERE product_id = products.id AND outlet_id = ? AND deleted_at IS NULL)",
			outletID,
		)
	}

	err := db.Order("name ASC").Find(&products).Error
	if err != nil {
		return nil, err
	}

	// Build a warehouse stock map via a single inventory_batches query
	stockMap := make(map[string]float64)
	if warehouseID != "" {
		var batches []inventoryModels.InventoryBatch
		batchErr := database.GetDB(ctx, r.db).
			Select("product_id, current_quantity, reserved_quantity").
			Where("warehouse_id = ? AND deleted_at IS NULL AND (expiry_date IS NULL OR expiry_date > CURRENT_TIMESTAMP)", warehouseID).
			Find(&batches).Error
		if batchErr == nil {
			for _, b := range batches {
				avail := b.CurrentQuantity - b.ReservedQuantity
				if avail < 0 {
					avail = 0
				}
				stockMap[b.ProductID] += avail
			}
		}
	}

	result := make([]POSCatalogProduct, 0, len(products))
	for _, p := range products {
		avail := stockMap[p.ID]
		if p.ProductKind == productModels.ProductKindRecipe {
			avail = computeRecipeProducibleQty(p.RecipeItems, stockMap)
		}

		categoryName := ""
		if p.Category != nil {
			categoryName = p.Category.Name
		}
		imageURL := ""
		if p.ImageURL != nil {
			imageURL = *p.ImageURL
		}
		isAvailable := false
		switch p.ProductKind {
		case productModels.ProductKindStock:
			isAvailable = avail > 0
		case productModels.ProductKindRecipe:
			isAvailable = avail > 0
		case productModels.ProductKindService:
			isAvailable = true
		default:
			isAvailable = false
		}
		result = append(result, POSCatalogProduct{
			ProductID:   p.ID,
			ProductCode: p.Code,
			ProductName: p.Name,
			ProductKind: p.ProductKind,
			Price:       p.SellingPrice,
			Stock:       avail,
			ImageURL:    imageURL,
			Category:    categoryName,
			IsAvailable: isAvailable,
		})
	}
	return result, nil
}

func computeRecipeProducibleQty(items []productModels.ProductRecipeItem, ingredientStock map[string]float64) float64 {
	if len(items) == 0 {
		return 0
	}

	maxProducible := math.MaxFloat64
	hasValidIngredient := false
	for _, item := range items {
		if item.Quantity <= 0 {
			continue
		}
		hasValidIngredient = true
		available := ingredientStock[item.IngredientProductID]
		possible := available / item.Quantity
		if possible < maxProducible {
			maxProducible = possible
		}
	}

	if !hasValidIngredient || maxProducible == math.MaxFloat64 {
		return 0
	}
	if maxProducible < 0 {
		return 0
	}

	return math.Floor(maxProducible)
}
