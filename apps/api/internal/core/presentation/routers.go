package presentation

import (
	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/presentation/handler"
	"github.com/gilabs/gims/api/internal/core/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterMasterDataRoutes registers all core master data routes
func RegisterMasterDataRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize repositories
	paymentTermsRepo := repositories.NewPaymentTermsRepository(db)
	courierAgencyRepo := repositories.NewCourierAgencyRepository(db)
	soSourceRepo := repositories.NewSOSourceRepository(db)
	leaveTypeRepo := repositories.NewLeaveTypeRepository(db)
	currencyRepo := repositories.NewCurrencyRepository(db)
	bankAccountRepo := repositories.NewBankAccountRepository(db)

	// Initialize usecases
	paymentTermsUC := usecase.NewPaymentTermsUsecase(paymentTermsRepo)
	courierAgencyUC := usecase.NewCourierAgencyUsecase(courierAgencyRepo)
	soSourceUC := usecase.NewSOSourceUsecase(soSourceRepo)
	leaveTypeUC := usecase.NewLeaveTypeUsecase(leaveTypeRepo)
	currencyUC := usecase.NewCurrencyUsecase(currencyRepo)
	bankAccountUC := usecase.NewBankAccountUsecaseWithCurrency(db, bankAccountRepo, currencyRepo)

	// Initialize handlers
	paymentTermsH := handler.NewPaymentTermsHandler(paymentTermsUC)
	courierAgencyH := handler.NewCourierAgencyHandler(courierAgencyUC)
	soSourceH := handler.NewSOSourceHandler(soSourceUC)
	leaveTypeH := handler.NewLeaveTypeHandler(leaveTypeUC)
	currencyH := handler.NewCurrencyHandler(currencyUC)
	bankAccountH := handler.NewBankAccountHandler(bankAccountUC)
	exportJobH := handler.NewExportJobHandler(exportjob.DefaultManager)

	// Create master-data group under API with auth middleware
	group := api.Group("/master-data")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))

	// Register routes
	router.RegisterPaymentTermsRoutes(group, paymentTermsH)
	router.RegisterCourierAgencyRoutes(group, courierAgencyH)
	router.RegisterSOSourceRoutes(group, soSourceH)
	router.RegisterLeaveTypeRoutes(group, leaveTypeH)
	router.RegisterCurrencyRoutes(group, currencyH)

	// Finance master data (kept under /finance to match seeded menus)
	financeGroup := api.Group("/finance")
	financeGroup.Use(middleware.AuthMiddleware(jwtManager, permService))
	router.RegisterBankAccountRoutes(financeGroup, bankAccountH)

	exportGroup := api.Group("")
	exportGroup.Use(middleware.AuthMiddleware(jwtManager, permService))
	router.RegisterExportJobRoutes(exportGroup, exportJobH)
}
