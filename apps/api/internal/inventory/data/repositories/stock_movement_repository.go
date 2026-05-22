package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
	"github.com/gilabs/gims/api/internal/inventory/domain/repository"
	"gorm.io/gorm"
)

type stockMovementRepository struct {
	db *gorm.DB
}

func NewStockMovementRepository(db *gorm.DB) repository.StockMovementRepository {
	return &stockMovementRepository{db}
}

func (r *stockMovementRepository) Create(ctx context.Context, movement *models.StockMovement) error {
	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// 1. Get current balance for this product + warehouse
		var lastBalance float64

		var lastMovement models.StockMovement
		err := tx.Where("warehouse_id = ? AND product_id = ?", movement.WarehouseID, movement.ProductID).
			Order("created_at desc").
			First(&lastMovement).Error

		if err == nil {
			lastBalance = lastMovement.Balance
		}

		// 2. Calculate new balance using directional quantities.
		movement.Balance = lastBalance + movement.QtyIn - movement.QtyOut

		// 3. Create record
		if err := tx.Create(movement).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *stockMovementRepository) FindAll(ctx context.Context, req *dto.GetStockMovementsRequest) ([]models.StockMovement, int64, error) {
	var movements []models.StockMovement
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.StockMovement{}).
		Preload("Product").
		Preload("Warehouse").
		Preload("Creator").
		Preload("InventoryBatch")

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.StockMovementScopeQueryOptions())

	if req.WarehouseID != "" {
		query = query.Where("warehouse_id = ?", req.WarehouseID)
	}
	if req.ProductID != "" {
		query = query.Where("product_id = ?", req.ProductID)
	}
	if req.Type != "" {
		query = query.Where("movement_type = ?", req.Type)
	}
	if req.StartDate != "" {
		sDate, _ := time.Parse("2006-01-02", req.StartDate)
		query = query.Where("date >= ?", sDate)
	}
	if req.EndDate != "" {
		eDate, _ := time.Parse("2006-01-02", req.EndDate)
		query = query.Where("date <= ?", eDate.Add(24*time.Hour).Add(-1*time.Second))
	}

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("ref_number ILIKE ? OR source ILIKE ?", search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PerPage
	err := query.Order("created_at desc").
		Limit(req.PerPage).
		Offset(offset).
		Find(&movements).Error

	if err != nil {
		return nil, 0, err
	}

	return movements, total, nil
}

func (r *stockMovementRepository) GetLastBalance(ctx context.Context, warehouseID, productID string) (float64, error) {
	var lastMovement models.StockMovement
	err := database.GetDB(ctx, r.db).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		Order("created_at desc").
		First(&lastMovement).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	return lastMovement.Balance, nil
}
