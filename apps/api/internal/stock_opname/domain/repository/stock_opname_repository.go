package repository

import (
	"context"

	"github.com/gilabs/gims/api/internal/stock_opname/data/models"
	"github.com/gilabs/gims/api/internal/stock_opname/domain/dto"
)

type StockOpnameRepository interface {
	Create(ctx context.Context, opname *models.StockOpname) error
	Update(ctx context.Context, opname *models.StockOpname) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.StockOpname, error)
	List(ctx context.Context, req *dto.ListStockOpnamesRequest) ([]models.StockOpname, int64, error)
	
	// Items
	ReplaceItems(ctx context.Context, opnameID string, items []models.StockOpnameItem) error
	ListItems(ctx context.Context, opnameID string) ([]models.StockOpnameItem, error)
	ListItemsPaginated(ctx context.Context, opnameID string, page, perPage int) ([]models.StockOpnameItem, int64, error)
	ListWarehouseStockSnapshot(ctx context.Context, warehouseID, scopeType string, categoryIDs, brandIDs []string) ([]dto.WarehouseStockSnapshot, error)

	// Status
	UpdateStatus(ctx context.Context, id string, status models.StockOpnameStatus, userID *string) error
	
	// Helper
	GetNextOpnameNumber(ctx context.Context) (string, error)

	// GetMyWarehouses returns warehouses assigned to the given user via user_warehouses.
	GetMyWarehouses(ctx context.Context, userID string) ([]dto.UserWarehouseInfo, error)
}
