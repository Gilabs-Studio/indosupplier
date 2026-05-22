package repository

import (
	"context"

	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
)

type InventoryRepository interface {
	GetStockList(ctx context.Context, req *dto.GetInventoryListRequest) ([]dto.InventoryStockItem, int64, error)
	GetInventoryMetrics(ctx context.Context) (*dto.InventoryMetrics, error)

	// Tree View
	GetTreeWarehouses(ctx context.Context) ([]dto.GetInventoryTreeWarehousesResponse, error)
	GetTreeProducts(ctx context.Context, req *dto.GetInventoryTreeProductsRequest) ([]dto.InventoryStockItem, int64, dto.TreeProductsSummary, error)
	GetTreeBatches(ctx context.Context, req *dto.GetInventoryTreeBatchesRequest) ([]dto.InventoryBatchItem, int64, error)
	GetProductLedgers(ctx context.Context, productID string, req *dto.GetProductStockLedgersRequest) ([]models.StockLedger, int64, error)

	// Stock Management
	UpdateProductReservedStock(ctx context.Context, productID string, quantity float64) error
	UpdateBatchQuantity(ctx context.Context, batchID string, quantity float64) error
	GetBatchesByProduct(ctx context.Context, productID string) ([]dto.InventoryBatchItem, error)
	CreateStockMovement(ctx context.Context, movement *dto.StockMovementRequest) (string, error)

	// Stock Receiving
	CreateBatch(ctx context.Context, batch *dto.CreateBatchParams) (string, error)
	UpdateProductAverageCost(ctx context.Context, productID string, newCost float64) error
	GetProductCostInfo(ctx context.Context, productID string) (float64, float64, error) // Returns CurrentHpp, TotalStock
	UpdateProductStock(ctx context.Context, productID string, delta float64) error

	// Batch-level Stock Reservation
	GetBatchByID(ctx context.Context, batchID string) (*dto.InventoryBatchDetail, error)
	UpdateBatchReservedQuantity(ctx context.Context, batchID string, quantity float64) error

	// Batch lookup by product + warehouse (for opname adjustments)
	GetBatchesByProductAndWarehouse(ctx context.Context, productID, warehouseID string) ([]dto.InventoryBatchItem, error)
}
