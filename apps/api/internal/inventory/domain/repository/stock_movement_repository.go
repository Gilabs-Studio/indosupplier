package repository

import (
	"context"

	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
)

type StockMovementRepository interface {
	Create(ctx context.Context, movement *models.StockMovement) error
	FindAll(ctx context.Context, req *dto.GetStockMovementsRequest) ([]models.StockMovement, int64, error)
	GetLastBalance(ctx context.Context, warehouseID, productID string) (float64, error)
}
