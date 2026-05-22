package presentation

import (
	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gilabs/gims/api/internal/crm/presentation/router"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	productRepos "github.com/gilabs/gims/api/internal/product/data/repositories"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all CRM domain routes
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize repositories
	pipelineStageRepo := repositories.NewPipelineStageRepository(db)
	leadSourceRepo := repositories.NewLeadSourceRepository(db)
	leadStatusRepo := repositories.NewLeadStatusRepository(db)
	contactRoleRepo := repositories.NewContactRoleRepository(db)
	activityTypeRepo := repositories.NewActivityTypeRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	leadRepo := repositories.NewLeadRepository(db)
	dealRepo := repositories.NewDealRepository(db)
	visitReportRepo := repositories.NewVisitReportRepository(db)
	activityRepo := repositories.NewActivityRepository(db)
	taskRepo := repositories.NewTaskRepository(db)
	reminderRepo := repositories.NewReminderRepository(db)
	scheduleRepo := repositories.NewScheduleRepository(db)
	areaCaptureRepo := repositories.NewAreaCaptureRepository(db)
	areaMappingRepo := repositories.NewAreaMappingRepository(db)
	customerRepo := customerRepos.NewCustomerRepository(db)
	employeeRepo := orgRepos.NewEmployeeRepository(db)
	businessTypeRepo := orgRepos.NewBusinessTypeRepository(db)
	areaRepo := orgRepos.NewAreaRepository(db)
	paymentTermsRepo := coreRepos.NewPaymentTermsRepository(db)
	productRepo := productRepos.NewProductRepository(db)
	salesQuotationRepo := salesRepos.NewSalesQuotationRepository(db)

	// Initialize usecases
	pipelineStageUC := usecase.NewPipelineStageUsecase(pipelineStageRepo, dealRepo)
	leadSourceUC := usecase.NewLeadSourceUsecase(leadSourceRepo)
	leadStatusUC := usecase.NewLeadStatusUsecase(leadStatusRepo)
	contactRoleUC := usecase.NewContactRoleUsecase(contactRoleRepo)
	activityTypeUC := usecase.NewActivityTypeUsecase(activityTypeRepo)
	contactUC := usecase.NewContactUsecase(contactRepo, contactRoleRepo, customerRepo)
	leadUC := usecase.NewLeadUsecase(leadRepo, leadStatusRepo, leadSourceRepo, contactRoleRepo, dealRepo, pipelineStageRepo, activityRepo, taskRepo, employeeRepo, businessTypeRepo, areaRepo, paymentTermsRepo)
	leadAutomationUC := usecase.NewLeadAutomationUsecase()
	dealUC := usecase.NewDealUsecase(dealRepo, pipelineStageRepo, customerRepo, contactRepo, employeeRepo, productRepo, leadRepo, activityRepo, salesQuotationRepo, db)
	visitReportUC := usecase.NewVisitReportUsecase(visitReportRepo, activityRepo, customerRepo, contactRepo, employeeRepo, dealRepo, leadRepo, productRepo, db)
	activityUC := usecase.NewActivityUsecase(activityRepo, activityTypeRepo, leadRepo, dealRepo)
	taskUC := usecase.NewTaskUsecase(taskRepo, scheduleRepo, reminderRepo, contactRepo, dealRepo, leadRepo, customerRepo, employeeRepo)
	scheduleUC := usecase.NewScheduleUsecase(scheduleRepo, taskRepo, employeeRepo)
	areaCaptureUC := usecase.NewAreaCaptureUsecase(areaCaptureRepo)
	areaMappingUC := usecase.NewAreaMappingUsecase(areaMappingRepo)

	// Initialize handlers
	pipelineStageH := handler.NewPipelineStageHandler(pipelineStageUC)
	leadSourceH := handler.NewLeadSourceHandler(leadSourceUC)
	leadStatusH := handler.NewLeadStatusHandler(leadStatusUC)
	contactRoleH := handler.NewContactRoleHandler(contactRoleUC)
	activityTypeH := handler.NewActivityTypeHandler(activityTypeUC)
	contactH := handler.NewContactHandler(contactUC)
	leadH := handler.NewLeadHandler(leadUC)
	leadAutomationH := handler.NewLeadAutomationHandler(leadAutomationUC)
	dealH := handler.NewDealHandler(dealUC)
	visitReportH := handler.NewVisitReportHandler(visitReportUC)
	visitReportPrintH := handler.NewVisitReportPrintHandler(visitReportUC)
	activityH := handler.NewActivityHandler(activityUC, db)
	taskH := handler.NewTaskHandler(taskUC)
	scheduleH := handler.NewScheduleHandler(scheduleUC)
	areaCaptureH := handler.NewAreaCaptureHandler(areaCaptureUC)
	areaMappingH := handler.NewAreaMappingHandler(areaMappingUC)

	// Create CRM group under API with auth middleware
	group := api.Group("/crm")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	group.Use(middleware.ScopeMiddleware(db))

	// Register routes
	router.RegisterPipelineStageRoutes(group, pipelineStageH)
	router.RegisterLeadSourceRoutes(group, leadSourceH)
	router.RegisterLeadStatusRoutes(group, leadStatusH)
	router.RegisterContactRoleRoutes(group, contactRoleH)
	router.RegisterActivityTypeRoutes(group, activityTypeH)
	router.RegisterContactRoutes(group, contactH)
	router.RegisterLeadRoutes(group, leadH)
	router.RegisterLeadAutomationRoutes(group, leadAutomationH)
	router.RegisterDealRoutes(group, dealH)
	router.RegisterVisitReportRoutes(group, visitReportH, visitReportPrintH)
	router.RegisterActivityRoutes(group, activityH)
	router.RegisterTaskRoutes(group, taskH)
	router.RegisterScheduleRoutes(group, scheduleH)
	router.RegisterAreaCaptureRoutes(group, areaCaptureH)
	router.RegisterAreaMappingRoutes(group, areaMappingH)
}
