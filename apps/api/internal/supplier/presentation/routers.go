package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/supplier/data/repositories"
	"github.com/gilabs/gims/api/internal/supplier/domain/usecase"
	"github.com/gilabs/gims/api/internal/supplier/presentation/handler"
	"github.com/gilabs/gims/api/internal/supplier/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all supplier domain routes
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize repositories
	supplierTypeRepo := repositories.NewSupplierTypeRepository(db)
	bankRepo := repositories.NewBankRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)

	// Initialize usecases
	supplierTypeUC := usecase.NewSupplierTypeUsecase(supplierTypeRepo)
	bankUC := usecase.NewBankUsecase(bankRepo)
	supplierUC := usecase.NewSupplierUsecase(supplierRepo)

	// Initialize handlers
	supplierTypeH := handler.NewSupplierTypeHandler(supplierTypeUC)
	bankH := handler.NewBankHandler(bankUC)
	supplierH := handler.NewSupplierHandler(supplierUC)

	// Create supplier group under API with auth middleware
	group := api.Group("/supplier")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))

	// Register routes
	router.RegisterSupplierTypeRoutes(group, supplierTypeH)
	router.RegisterBankRoutes(group, bankH)
	router.RegisterSupplierRoutes(group, supplierH)
}
