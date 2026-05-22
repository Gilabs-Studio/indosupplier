package usecase

import (
	"context"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	invModels "github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/repository"
	prodModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecipeConsumptionService handles ingredient stock deduction when a RECIPE product is sold.
// For each recipe item, it selects batches via FIFO and creates OUT stock movements.
type RecipeConsumptionService interface {
	// ConsumeRecipeIngredients deducts ingredient stock for a recipe product sale.
	// qtySold is the number of recipe products sold (e.g., 2 portions of "Nasi Goreng").
	// refType/refID link the movements to the originating POS order or sales document.
	ConsumeRecipeIngredients(ctx context.Context, req ConsumeRecipeRequest) error

	// ReverseRecipeConsumption reverses ingredient deductions (e.g., order cancellation).
	ReverseRecipeConsumption(ctx context.Context, req ReverseRecipeRequest) error
}

type ConsumeRecipeRequest struct {
	ProductID   string  // The RECIPE product ID
	WarehouseID string  // The outlet/warehouse to deduct from
	QtySold     float64 // Number of portions sold
	RefType     string  // e.g., "POS_ORDER"
	RefID       string  // The order document ID
}

type ReverseRecipeRequest struct {
	ProductID   string
	WarehouseID string
	QtySold     float64
	RefType     string
	RefID       string
}

type recipeConsumptionService struct {
	db       *gorm.DB
	invRepo  repository.InventoryRepository
}

// NewRecipeConsumptionService creates a new RecipeConsumptionService
func NewRecipeConsumptionService(db *gorm.DB, invRepo repository.InventoryRepository) RecipeConsumptionService {
	return &recipeConsumptionService{
		db:      db,
		invRepo: invRepo,
	}
}

func (s *recipeConsumptionService) ConsumeRecipeIngredients(ctx context.Context, req ConsumeRecipeRequest) error {
	// Load the recipe product with its recipe items
	var product prodModels.Product
	if err := s.db.WithContext(ctx).
		Preload("RecipeItems").
		First(&product, "id = ?", req.ProductID).Error; err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	if product.ProductKind != prodModels.ProductKindRecipe {
		return fmt.Errorf("product %s is not a RECIPE kind", product.Code)
	}

	if len(product.RecipeItems) == 0 {
		return fmt.Errorf("product %s has no recipe items", product.Code)
	}

	return database.RetryTx(s.db, func(tx *gorm.DB) error {
		for _, item := range product.RecipeItems {
			totalQty := item.Quantity * req.QtySold

			// Select batches via FIFO for this ingredient + warehouse
			batches, err := s.selectFIFOBatches(ctx, tx, item.IngredientProductID, req.WarehouseID, totalQty)
			if err != nil {
				return fmt.Errorf("insufficient stock for ingredient %s: %w", item.IngredientProductID, err)
			}

			for _, batch := range batches {
				// Deduct from batch
				if err := tx.Model(&invModels.InventoryBatch{}).
					Where("id = ?", batch.BatchID).
					Update("current_quantity", gorm.Expr("current_quantity - ?", batch.DeductQty)).Error; err != nil {
					return fmt.Errorf("failed to deduct batch %s: %w", batch.BatchID, err)
				}

				// Create OUT movement
				movement := invModels.StockMovement{
					ID:               uuid.New().String(),
					ProductID:        item.IngredientProductID,
					WarehouseID:      req.WarehouseID,
					InventoryBatchID: &batch.BatchID,
					MovementType:     "OUT",
					QtyOut:           batch.DeductQty,
					RefType:          req.RefType,
					RefID:            req.RefID,
					Cost:             batch.CostPrice * batch.DeductQty,
					Source:           fmt.Sprintf("Recipe consumption: %s x%.2f", product.Name, req.QtySold),
				}
				if err := tx.Create(&movement).Error; err != nil {
					return fmt.Errorf("failed to create movement: %w", err)
				}
			}
		}
		return nil
	})
}

func (s *recipeConsumptionService) ReverseRecipeConsumption(ctx context.Context, req ReverseRecipeRequest) error {
	// Load the recipe product with its recipe items
	var product prodModels.Product
	if err := s.db.WithContext(ctx).
		Preload("RecipeItems").
		First(&product, "id = ?", req.ProductID).Error; err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	return database.RetryTx(s.db, func(tx *gorm.DB) error {
		for _, item := range product.RecipeItems {
			totalQty := item.Quantity * req.QtySold

			// Find the original OUT movements for this ref
			var movements []invModels.StockMovement
			if err := tx.Where("product_id = ? AND warehouse_id = ? AND ref_type = ? AND ref_id = ? AND movement_type = 'OUT'",
				item.IngredientProductID, req.WarehouseID, req.RefType, req.RefID).
				Find(&movements).Error; err != nil {
				return err
			}

			if len(movements) == 0 {
				// No movements found to reverse, create a simple IN adjustment
				movement := invModels.StockMovement{
					ID:           uuid.New().String(),
					ProductID:    item.IngredientProductID,
					WarehouseID:  req.WarehouseID,
					MovementType: "IN",
					QtyIn:        totalQty,
					RefType:      req.RefType + "_REVERSAL",
					RefID:        req.RefID,
					Source:       fmt.Sprintf("Recipe reversal: %s x%.2f", product.Name, req.QtySold),
				}
				if err := tx.Create(&movement).Error; err != nil {
					return err
				}
				continue
			}

			// Reverse each original movement: restore batch quantity + create IN movement
			for _, mov := range movements {
				if mov.InventoryBatchID != nil {
					if err := tx.Model(&invModels.InventoryBatch{}).
						Where("id = ?", *mov.InventoryBatchID).
						Update("current_quantity", gorm.Expr("current_quantity + ?", mov.QtyOut)).Error; err != nil {
						return err
					}
				}

				reversal := invModels.StockMovement{
					ID:               uuid.New().String(),
					ProductID:        mov.ProductID,
					WarehouseID:      mov.WarehouseID,
					InventoryBatchID: mov.InventoryBatchID,
					MovementType:     "IN",
					QtyIn:            mov.QtyOut,
					RefType:          req.RefType + "_REVERSAL",
					RefID:            req.RefID,
					Cost:             mov.Cost,
					Source:           fmt.Sprintf("Recipe reversal: %s x%.2f", product.Name, req.QtySold),
				}
				if err := tx.Create(&reversal).Error; err != nil {
					return err
				}
			}
			_ = totalQty // used for context; actual reversal based on original movements
		}
		return nil
	})
}

type batchDeduction struct {
	BatchID   string
	DeductQty float64
	CostPrice float64
}

// selectFIFOBatches selects batches in FIFO order for a given product+warehouse
func (s *recipeConsumptionService) selectFIFOBatches(ctx context.Context, tx *gorm.DB, productID, warehouseID string, requiredQty float64) ([]batchDeduction, error) {
	var batches []invModels.InventoryBatch
	if err := tx.WithContext(ctx).
		Where("product_id = ? AND warehouse_id = ? AND current_quantity > 0", productID, warehouseID).
		Order("created_at ASC"). // FIFO: oldest first
		Find(&batches).Error; err != nil {
		return nil, err
	}

	var deductions []batchDeduction
	remaining := requiredQty

	for _, batch := range batches {
		if remaining <= 0 {
			break
		}

		available := batch.CurrentQuantity - batch.ReservedQuantity
		if available <= 0 {
			continue
		}

		deductQty := available
		if deductQty > remaining {
			deductQty = remaining
		}

		deductions = append(deductions, batchDeduction{
			BatchID:   batch.ID,
			DeductQty: deductQty,
			CostPrice: batch.CostPrice,
		})
		remaining -= deductQty
	}

	if remaining > 0 {
		return nil, fmt.Errorf("insufficient stock: need %.4f more units", remaining)
	}

	return deductions, nil
}
