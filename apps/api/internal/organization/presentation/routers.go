package presentation

import (
	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	financeRepositories "github.com/gilabs/gims/api/internal/finance/data/repositories"
	orgRepositories "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/service"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gilabs/gims/api/internal/organization/presentation/router"
	tenantRepositories "github.com/gilabs/gims/api/internal/tenant/data/repositories"
	warehouseRepositories "github.com/gilabs/gims/api/internal/warehouse/data/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all organization domain routes
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize repositories
	divisionRepo := orgRepositories.NewDivisionRepository(db)
	jobPositionRepo := orgRepositories.NewJobPositionRepository(db)
	businessUnitRepo := orgRepositories.NewBusinessUnitRepository(db)
	businessTypeRepo := orgRepositories.NewBusinessTypeRepository(db)
	areaRepo := orgRepositories.NewAreaRepository(db)
	companyRepo := orgRepositories.NewCompanyRepository(db)
	employeeRepo := orgRepositories.NewEmployeeRepository(db)
	employeeAreaRepo := orgRepositories.NewEmployeeAreaRepository(db)
	employeeOutletRepo := orgRepositories.NewEmployeeOutletRepository(db)
	employeeWarehouseRepo := orgRepositories.NewEmployeeWarehouseRepository(db)
	employeeContractRepo := orgRepositories.NewEmployeeContractRepository(db)
	educationHistoryRepo := orgRepositories.NewEmployeeEducationHistoryRepository(db)
	certificationRepo := orgRepositories.NewEmployeeCertificationRepository(db)
	assetRepo := orgRepositories.NewEmployeeAssetRepository(db)
	signatureRepo := orgRepositories.NewEmployeeSignatureRepository(db)

	// Finance repositories for asset borrowing integration
	financeAssetRepo := financeRepositories.NewAssetRepository(db)
	auditLogRepo := financeRepositories.NewAssetAuditLogRepository(db)

	// Core repositories
	timezoneRepo := repositories.NewTimeZoneRepository(db)

	// Initialize services
	timezoneService := service.NewTimezoneService(timezoneRepo)

	// Initialize usecases
	divisionUC := usecase.NewDivisionUsecase(divisionRepo)
	jobPositionUC := usecase.NewJobPositionUsecase(jobPositionRepo)
	businessUnitUC := usecase.NewBusinessUnitUsecase(businessUnitRepo)
	businessTypeUC := usecase.NewBusinessTypeUsecase(businessTypeRepo)
	// Pass employeeAreaRepo so the usecase can manage supervisor/member assignments.
	// Pass employeeRepo for GetFormData endpoint.
	areaUC := usecase.NewAreaUsecase(areaRepo, employeeAreaRepo, employeeRepo)
	// Initialize outlets and warehouse repositories early for company cascade deactivation
	outletRepo := orgRepositories.NewOutletRepository(db)
	warehouseRepo := warehouseRepositories.NewWarehouseRepository(db)
	companyUC := usecase.NewCompanyUsecase(companyRepo, outletRepo, warehouseRepo, timezoneService)
	employeeUC := usecase.NewEmployeeUsecase(employeeRepo, employeeAreaRepo, employeeOutletRepo, employeeWarehouseRepo, divisionRepo, jobPositionRepo, companyRepo, areaRepo, outletRepo, warehouseRepo, employeeContractRepo, educationHistoryRepo, certificationRepo, assetRepo, signatureRepo, financeAssetRepo, auditLogRepo)

	// Initialize handlers
	divisionH := handler.NewDivisionHandler(divisionUC)
	jobPositionH := handler.NewJobPositionHandler(jobPositionUC)
	businessUnitH := handler.NewBusinessUnitHandler(businessUnitUC)
	businessTypeH := handler.NewBusinessTypeHandler(businessTypeUC)
	areaH := handler.NewAreaHandler(areaUC)
	companyH := handler.NewCompanyHandler(companyUC)
	employeeH := handler.NewEmployeeHandler(employeeUC)

	// Outlet dependencies (cross-domain: warehouse repo)
	subscriptionRepo := tenantRepositories.NewSubscriptionRepository(db)
	outletUC := usecase.NewOutletUsecase(db, outletRepo, warehouseRepo, employeeRepo, companyRepo, subscriptionRepo)
	outletH := handler.NewOutletHandler(outletUC)

	// Create organization group under API with auth middleware
	group := api.Group("/organization")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	group.Use(middleware.ScopeMiddleware(db))

	// Register routes
	router.RegisterDivisionRoutes(group, divisionH)
	router.RegisterJobPositionRoutes(group, jobPositionH)
	router.RegisterBusinessUnitRoutes(group, businessUnitH)
	router.RegisterBusinessTypeRoutes(group, businessTypeH)
	router.RegisterAreaRoutes(group, areaH)
	router.RegisterCompanyRoutes(group, companyH)
	router.RegisterEmployeeRoutes(group, employeeH)
	router.RegisterOutletRoutes(group, outletH)
}
