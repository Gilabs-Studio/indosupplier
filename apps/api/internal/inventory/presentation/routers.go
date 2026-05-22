package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/data/repositories"
	"github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	"github.com/gilabs/gims/api/internal/inventory/presentation/handler"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(
	r *gin.Engine,
	v1 *gin.RouterGroup,
	db *gorm.DB,
	jwtManager *jwt.JWTManager,
	permissionService security.PermissionService,
	inventoryUsecase usecase.InventoryUsecase,
) {
	// Auto Migrate
	db.AutoMigrate(&models.InventoryBatch{}, &models.StockMovement{}, &models.StockLedger{})

	// Repositories
	// inventoryRepo := repositories.NewInventoryRepository(db) // Injected via usecase
	stockMovementRepo := repositories.NewStockMovementRepository(db)

	// Usecases
	// inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo) // Injected from main
	stockMovementUsecase := usecase.NewStockMovementService(stockMovementRepo)

	// Handlers
	inventoryHandler := handler.NewInventoryHandler(inventoryUsecase)
	stockMovementHandler := NewStockMovementHandler(stockMovementUsecase, inventoryUsecase)

	// Routes
	stock := v1.Group("/stock")
	stock.Use(middleware.AuthMiddleware(jwtManager, permissionService))
	stock.Use(middleware.ScopeMiddleware(db))
	{
		stock.GET("/inventory", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetStockList)
		// Metrics endpoint for owner/admin dashboards — placed before /:id routes for specificity
		stock.GET("/inventory/metrics", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetInventoryMetrics)

		// Tree View Routes
		stock.GET("/tree/warehouses", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetTreeWarehouses)
		stock.GET("/tree/products", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetTreeProducts)
		stock.GET("/tree/batches", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetTreeBatches)

		// Movement Routes
		stock.GET("/movements", middleware.PermissionMiddleware("stock_movement.read"), stockMovementHandler.GetMovements)
		stock.POST("/movements", middleware.PermissionMiddleware("stock_movement.create"), stockMovementHandler.CreateMovement)
	}

	inventory := v1.Group("/inventory")
	inventory.Use(middleware.AuthMiddleware(jwtManager, permissionService))
	inventory.Use(middleware.ScopeMiddleware(db))
	{
		// Alias endpoints for new inventory namespace (kept compatible with existing /stock routes)
		inventory.GET("/stock", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetStockList)
		inventory.GET("/stock/metrics", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetInventoryMetrics)
		inventory.GET("/stock/:product_id", middleware.PermissionMiddleware("inventory.read"), inventoryHandler.GetStockByProduct)
		inventory.GET("/movements", middleware.PermissionMiddleware("stock_movement.read"), stockMovementHandler.GetMovements)

		inventory.GET("/products/:id/ledgers", middleware.PermissionMiddleware("ledger.read"), inventoryHandler.GetProductLedgers)
	}
}
