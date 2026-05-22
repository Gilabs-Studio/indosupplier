package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	inventoryUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	"github.com/gilabs/gims/api/internal/stock_opname/data/repositories"
	"github.com/gilabs/gims/api/internal/stock_opname/domain/usecase"
	"github.com/gilabs/gims/api/internal/stock_opname/presentation/handler"
	"github.com/gilabs/gims/api/internal/stock_opname/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(
	r *gin.Engine,
	v1 *gin.RouterGroup,
	db *gorm.DB,
	jwtManager *jwt.JWTManager,
	permissionService security.PermissionService,
	invUC inventoryUsecase.InventoryUsecase,
	journalUC finUsecase.JournalEntryUsecase,
	coaUC finUsecase.ChartOfAccountUsecase,
	settingsUC financesettings.SettingsService,
) {
	// Repositories
	opnameRepo := repositories.NewStockOpnameRepository(db)

	// Usecases — inject inventory usecase for stock movement creation on Post
	opnameUC := usecase.NewStockOpnameUsecase(opnameRepo, invUC, journalUC, coaUC, settingsUC)

	// Handlers
	opnameHandler := handler.NewStockOpnameHandler(opnameUC)

	// Routes
	router.RegisterStockOpnameRoutes(v1, opnameHandler, jwtManager, permissionService, db)
}
