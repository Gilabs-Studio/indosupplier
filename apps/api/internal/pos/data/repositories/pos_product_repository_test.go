package repositories

import (
	"context"
	"testing"

	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPOSProductRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	schema := []string{
		`CREATE TABLE product_categories (
			id TEXT PRIMARY KEY,
			name TEXT,
			deleted_at DATETIME
		)`,
		`CREATE TABLE products (
			id TEXT PRIMARY KEY,
			code TEXT,
			name TEXT,
			category_id TEXT,
			selling_price REAL,
			status TEXT,
			is_approved BOOLEAN,
			is_active BOOLEAN,
			pos_scope TEXT NOT NULL DEFAULT 'none',
			product_kind TEXT,
			image_url TEXT,
			deleted_at DATETIME
		)`,
		`CREATE TABLE inventory_batches (
			id TEXT PRIMARY KEY,
			product_id TEXT,
			warehouse_id TEXT,
			expiry_date DATETIME,
			current_quantity REAL,
			reserved_quantity REAL,
			deleted_at DATETIME
		)`,
		`CREATE TABLE product_recipe_items (
			id TEXT PRIMARY KEY,
			product_id TEXT,
			ingredient_product_id TEXT,
			quantity REAL,
			uom_id TEXT,
			notes TEXT,
			sort_order INTEGER,
			deleted_at DATETIME
		)`,
	}

	for _, ddl := range schema {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("failed to create sqlite schema: %v", err)
		}
	}

	return db
}

func seedPOSCatalogProducts(t *testing.T, db *gorm.DB) {
	t.Helper()
	categoryID := "c0000001-0000-0000-0000-000000000001"
	err := db.Exec(`
		INSERT INTO product_categories (id, name, deleted_at)
		VALUES (?, ?, NULL)
	`, categoryID, "Food").Error
	if err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	products := []productModels.Product{
		{
			ID:             "10000001-0000-0000-0000-000000000001",
			Code:           "STK-1",
			Name:           "Stock Product",
			CategoryID:     &categoryID,
			SellingPrice:   10000,
			Status:         productModels.ProductStatusApproved,
			IsApproved:     true,
			IsActive:    true,
			POSScope:    productModels.POSScopeGlobal,
			ProductKind: productModels.ProductKindStock,
		},
		{
			ID:             "10000001-0000-0000-0000-000000000002",
			Code:           "RCP-1",
			Name:           "Recipe Product",
			CategoryID:     &categoryID,
			SellingPrice:   25000,
			Status:         productModels.ProductStatusApproved,
			IsApproved:     true,
			IsActive:    true,
			POSScope:    productModels.POSScopeGlobal,
			ProductKind: productModels.ProductKindRecipe,
		},
		{
			ID:             "10000001-0000-0000-0000-000000000003",
			Code:           "SRV-1",
			Name:           "Service Product",
			CategoryID:     &categoryID,
			SellingPrice:   5000,
			Status:         productModels.ProductStatusApproved,
			IsApproved:     true,
			IsActive:    true,
			POSScope:    productModels.POSScopeGlobal,
			ProductKind: productModels.ProductKindService,
		},
		{
			ID:             "10000001-0000-0000-0000-000000000004",
			Code:           "PND-1",
			Name:           "Pending Product",
			CategoryID:     &categoryID,
			SellingPrice:   9000,
			Status:         productModels.ProductStatusPending,
			IsApproved:  false,
			IsActive:    true,
			POSScope:    productModels.POSScopeGlobal,
			ProductKind: productModels.ProductKindStock,
		},
	}

	for i := range products {
		if err := db.Exec(`
			INSERT INTO products (
				id, code, name, category_id, selling_price, status,
				is_approved, is_active, pos_scope, product_kind,
				image_url, deleted_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL)
		`,
			products[i].ID,
			products[i].Code,
			products[i].Name,
			products[i].CategoryID,
			products[i].SellingPrice,
			products[i].Status,
			products[i].IsApproved,
			products[i].IsActive,
			products[i].POSScope,
			products[i].ProductKind,
		).Error; err != nil {
			t.Fatalf("failed to seed product %s: %v", products[i].Code, err)
		}
	}

	warehouseID := "20000001-0000-0000-0000-000000000001"
	err = db.Exec(`
		INSERT INTO inventory_batches (
			id, warehouse_id, product_id, current_quantity, reserved_quantity, deleted_at
		) VALUES (?, ?, ?, ?, ?, NULL)
	`, "30000001-0000-0000-0000-000000000001", warehouseID, "10000001-0000-0000-0000-000000000001", 10, 2).Error
	if err != nil {
		t.Fatalf("failed to seed inventory batch: %v", err)
	}

	err = db.Exec(`
		INSERT INTO product_recipe_items (
			id, product_id, ingredient_product_id, quantity, uom_id, notes, sort_order, deleted_at
		) VALUES (?, ?, ?, ?, NULL, NULL, 0, NULL)
	`, "40000001-0000-0000-0000-000000000001", "10000001-0000-0000-0000-000000000002", "10000001-0000-0000-0000-000000000001", 2).Error
	if err != nil {
		t.Fatalf("failed to seed recipe item: %v", err)
	}
}

func TestFindPOSAvailable_ShouldIncludeStockAndRecipe_WhenApprovedAndPosAvailable(t *testing.T) {
	db := setupPOSProductRepoTestDB(t)
	seedPOSCatalogProducts(t, db)
	repo := NewPOSProductRepository(db)

	items, err := repo.FindPOSAvailable(context.Background(), "20000001-0000-0000-0000-000000000001", "")
	assert.NoError(t, err)
	assert.Len(t, items, 3)

	var stockItem POSCatalogProduct
	var recipeItem POSCatalogProduct
	var serviceItem POSCatalogProduct
	for _, item := range items {
		switch item.ProductCode {
		case "STK-1":
			stockItem = item
		case "RCP-1":
			recipeItem = item
		case "SRV-1":
			serviceItem = item
		}
	}

	assert.Equal(t, 8.0, stockItem.Stock)
	assert.True(t, stockItem.IsAvailable)
	assert.Equal(t, 4.0, recipeItem.Stock)
	assert.True(t, recipeItem.IsAvailable)
	assert.True(t, serviceItem.IsAvailable)
}

func TestFindPOSAvailable_ShouldKeepStockProductVisible_WhenNoWarehouseStock(t *testing.T) {
	db := setupPOSProductRepoTestDB(t)
	seedPOSCatalogProducts(t, db)
	repo := NewPOSProductRepository(db)

	// Warehouse with no inventory — STOCK and RECIPE stay visible but unavailable.
	items, err := repo.FindPOSAvailable(context.Background(), "20000001-0000-0000-0000-000000000099", "")
	assert.NoError(t, err)
	assert.Len(t, items, 3)

	var foundStock bool
	for _, item := range items {
		if item.ProductCode == "STK-1" {
			foundStock = true
			assert.Equal(t, 0.0, item.Stock)
			assert.False(t, item.IsAvailable)
		}
		if item.ProductCode == "RCP-1" {
			assert.Equal(t, 0.0, item.Stock)
			assert.False(t, item.IsAvailable)
		}
	}
	assert.True(t, foundStock)
}
