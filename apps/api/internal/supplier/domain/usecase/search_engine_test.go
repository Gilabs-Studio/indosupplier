package usecase_test

import (
	"context"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	buyerUsecase "github.com/gilabs/indosupplier/api/internal/buyer/domain/usecase"
	supplierRepos "github.com/gilabs/indosupplier/api/internal/supplier/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/usecase"
)

func getTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5435 user=postgres password=postgres dbname=indosupplier_db sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping integration test; could not connect to Postgres: %v", err)
	}
	return db
}

func TestSearchEngineTypoTolerance(t *testing.T) {
	db := getTestDB(t)
	discoveryUC := usecase.NewDiscoveryUsecase(db)
	ctx := context.Background()

	// 1. Exact search: "laptop"
	t.Run("Exact search for laptop", func(t *testing.T) {
		products, err := discoveryUC.ListProducts(ctx, dto.ListPublicProductsParams{
			Query: "laptop",
			Limit: 10,
		})
		if err != nil {
			t.Fatalf("ListProducts failed: %v", err)
		}
		if len(products) == 0 {
			t.Fatalf("Expected laptop products to be found, got 0")
		}
		t.Logf("Found %d products for query 'laptop'", len(products))
		for _, p := range products {
			t.Logf(" - Product: %s (Price: %v)", p.Name, p.Price)
		}
	})

	// 2. Typo search: "Lapta" (matches user's specific use case)
	t.Run("Typo search for Lapta", func(t *testing.T) {
		products, err := discoveryUC.ListProducts(ctx, dto.ListPublicProductsParams{
			Query: "Lapta",
			Limit: 10,
		})
		if err != nil {
			t.Fatalf("ListProducts failed with typo query: %v", err)
		}
		if len(products) == 0 {
			t.Fatalf("Expected laptop products to be found for typo 'Lapta', got 0")
		}
		t.Logf("Found %d products for typo 'Lapta'", len(products))
		for _, p := range products {
			t.Logf(" - [Typo Match] Product: %s", p.Name)
		}
	})

	// 3. Typo search: "leptop"
	t.Run("Typo search for leptop", func(t *testing.T) {
		products, err := discoveryUC.ListProducts(ctx, dto.ListPublicProductsParams{
			Query: "leptop",
			Limit: 10,
		})
		if err != nil {
			t.Fatalf("ListProducts failed: %v", err)
		}
		if len(products) == 0 {
			t.Fatalf("Expected laptop products for typo 'leptop', got 0")
		}
		t.Logf("Found %d products for typo 'leptop'", len(products))
	})
}

func TestCategoriesWithCounts(t *testing.T) {
	db := getTestDB(t)
	productRepo := supplierRepos.NewProductRepository(db)
	ctx := context.Background()

	cats, err := productRepo.ListCategories(ctx)
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}

	if len(cats) < 8 {
		t.Fatalf("Expected at least 8 categories, got %d", len(cats))
	}

	t.Logf("Categories found: %d", len(cats))
	foundElectronics := false
	for _, c := range cats {
		t.Logf(" - Category [%s] '%s': IconURL=%s, Products=%d, Suppliers=%d",
			c.Slug, c.Name, c.IconURL, c.ProductCount, c.SupplierCount)
		if c.IconURL == "" {
			t.Errorf("Category %s has empty IconURL", c.Slug)
		}
		if c.Slug == "electronics-it" {
			foundElectronics = true
			if c.ProductCount == 0 {
				t.Errorf("Expected electronics-it to have products, got %d", c.ProductCount)
			}
		}
	}

	if !foundElectronics {
		t.Errorf("electronics-it category not found in categories list")
	}
}

func TestBuyerCompareClear(t *testing.T) {
	db := getTestDB(t)
	compareUC := buyerUsecase.NewCompareUsecase(db)
	ctx := context.Background()

	var buyerProfile struct {
		UserID string
	}
	if err := db.Table("buyer_profiles").Select("user_id").First(&buyerProfile).Error; err != nil {
		t.Skip("No buyer profile found")
	}
	testUserID := buyerProfile.UserID
	err := compareUC.ClearProducts(ctx, testUserID)
	if err != nil {
		t.Fatalf("ClearProducts failed: %v", err)
	}
	t.Log("Compare products clear succeeded")

	// Test Clear (suppliers)
	err = compareUC.Clear(ctx, testUserID)
	if err != nil {
		t.Fatalf("Clear suppliers failed: %v", err)
	}
	t.Log("Compare suppliers clear succeeded")
}
