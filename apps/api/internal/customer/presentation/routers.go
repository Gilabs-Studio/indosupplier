package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	crmRepos "github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/customer/data/repositories"
	"github.com/gilabs/gims/api/internal/customer/domain/usecase"
	"github.com/gilabs/gims/api/internal/customer/presentation/handler"
	"github.com/gilabs/gims/api/internal/customer/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all customer domain routes
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize repositories
	customerTypeRepo := repositories.NewCustomerTypeRepository(db)
	customerRepo := repositories.NewCustomerRepository(db)
	contactRepo := crmRepos.NewContactRepository(db)

	// Initialize usecases
	customerTypeUC := usecase.NewCustomerTypeUsecase(customerTypeRepo)
	customerUC := usecase.NewCustomerUsecase(customerRepo, customerTypeRepo, contactRepo, db)

	// Initialize handlers
	customerTypeH := handler.NewCustomerTypeHandler(customerTypeUC)
	customerH := handler.NewCustomerHandler(customerUC)

	// Create customer group under API with auth middleware
	group := api.Group("/customer")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))

	// Register routes
	router.RegisterCustomerTypeRoutes(group, customerTypeH)
	router.RegisterCustomerRoutes(group, customerH)
}
