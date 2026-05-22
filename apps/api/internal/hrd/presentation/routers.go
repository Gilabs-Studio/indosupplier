package presentation

import (
	"context"

	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	financeRepos "github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	financeUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	hrdMapper "github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	hrdWS "github.com/gilabs/gims/api/internal/hrd/infrastructure/ws"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gilabs/gims/api/internal/hrd/presentation/router"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	orgUsecase "github.com/gilabs/gims/api/internal/organization/domain/usecase"
	warehouseRepositories "github.com/gilabs/gims/api/internal/warehouse/data/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HRDDeps holds exported HRD usecases for cross-module consumption
type HRDDeps struct {
	HolidayUC      usecase.HolidayUsecase
	LeaveRequestUC usecase.LeaveRequestUsecase
	AttendanceUC   usecase.AttendanceRecordUsecase
	SalaryUC       usecase.SalaryStructureUsecase
}

type HRDFinanceDeps struct {
	JournalUC  financeUsecase.JournalEntryUsecase
	CoaUC      financeUsecase.ChartOfAccountUsecase
	SettingsUC financesettings.SettingsService
}

// RegisterRoutes registers all HRD routes and returns shared dependencies
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}, financeDeps *HRDFinanceDeps) *HRDDeps {
	// Initialize repositories
	workScheduleRepo := repositories.NewWorkScheduleRepository(db)
	holidayRepo := repositories.NewHolidayRepository(db)
	attendanceRepo := repositories.NewAttendanceRecordRepository(db)
	salaryRepo := repositories.NewSalaryStructureRepository(db)
	overtimeRepo := repositories.NewOvertimeRequestRepository(db)
	leaveRequestRepo := repositories.NewLeaveRequestRepository(db)
	evaluationGroupRepo := repositories.NewEvaluationGroupRepository(db)
	evaluationCriteriaRepo := repositories.NewEvaluationCriteriaRepository(db)
	employeeEvaluationRepo := repositories.NewEmployeeEvaluationRepository(db)
	recruitmentRepo := repositories.NewRecruitmentRequestRepository(db)
	applicantRepo := repositories.NewRecruitmentApplicantRepository(db)
	applicantStageRepo := repositories.NewApplicantStageRepository(db)
	applicantActivityRepo := repositories.NewApplicantActivityRepository(db)

	// Seed default applicant stages
	_ = applicantStageRepo.SeedDefaultStages(context.Background())

	// Core repositories
	leaveTypeRepo := coreRepos.NewLeaveTypeRepository(db)

	// Organization repositories
	employeeRepo := orgRepos.NewEmployeeRepository(db)
	employeeAreaRepo := orgRepos.NewEmployeeAreaRepository(db)
	divisionRepo := orgRepos.NewDivisionRepository(db)
	positionRepo := orgRepos.NewJobPositionRepository(db)
	companyRepo := orgRepos.NewCompanyRepository(db)
	areaRepo := orgRepos.NewAreaRepository(db)
	employeeContractRepo := orgRepos.NewEmployeeContractRepository(db)
	educationHistoryRepo := orgRepos.NewEmployeeEducationHistoryRepository(db)
	certificationRepo := orgRepos.NewEmployeeCertificationRepository(db)
	assetRepo := orgRepos.NewEmployeeAssetRepository(db)
	employeeOutletRepo := orgRepos.NewEmployeeOutletRepository(db)
	employeeWarehouseRepo := orgRepos.NewEmployeeWarehouseRepository(db)
	outletRepo := orgRepos.NewOutletRepository(db)
	warehouseRepo := warehouseRepositories.NewWarehouseRepository(db)
	financeAssetRepo := financeRepos.NewAssetRepository(db)
	auditLogRepo := financeRepos.NewAssetAuditLogRepository(db)
	auditService := audit.NewAuditService(db)

	// Initialize usecases
	workScheduleUC := usecase.NewWorkScheduleUsecase(workScheduleRepo, divisionRepo, companyRepo)
	holidayUC := usecase.NewHolidayUsecase(holidayRepo)
	overtimeUC := usecase.NewOvertimeRequestUsecase(overtimeRepo, employeeRepo)
	attendanceUC := usecase.NewAttendanceRecordUsecase(attendanceRepo, workScheduleRepo, holidayRepo, leaveRequestRepo, employeeRepo, divisionRepo, overtimeUC)
	attendanceHub := hrdWS.DefaultAttendanceHub()
	attendanceUC = usecase.WithAttendanceTodayPublisher(attendanceUC, attendanceHub)
	leaveRequestUC := usecase.NewLeaveRequestUsecase(db, leaveRequestRepo, employeeRepo, leaveTypeRepo, holidayRepo, attendanceRepo)
	evaluationGroupUC := usecase.NewEvaluationGroupUsecase(db, evaluationGroupRepo, evaluationCriteriaRepo, auditService)
	evaluationCriteriaUC := usecase.NewEvaluationCriteriaUsecase(evaluationCriteriaRepo, evaluationGroupRepo, auditService)
	employeeEvaluationUC := usecase.NewEmployeeEvaluationUsecase(db, employeeEvaluationRepo, evaluationGroupRepo, evaluationCriteriaRepo, employeeRepo, auditService)
	recruitmentUC := usecase.NewRecruitmentRequestUsecase(recruitmentRepo, employeeRepo, divisionRepo, positionRepo)
	salaryMapper := hrdMapper.NewSalaryStructureMapper()
	var financeJournalUC financeUsecase.JournalEntryUsecase
	var financeSettingsUC financesettings.SettingsService
	var financeCoaUC financeUsecase.ChartOfAccountUsecase
	if financeDeps != nil {
		financeJournalUC = financeDeps.JournalUC
		financeSettingsUC = financeDeps.SettingsUC
		financeCoaUC = financeDeps.CoaUC
	}
	salaryUC := usecase.NewSalaryStructureUsecase(db, salaryRepo, salaryMapper, financeJournalUC, financeSettingsUC, financeCoaUC)
	signatureRepo := orgRepos.NewEmployeeSignatureRepository(db)
	employeeUC := orgUsecase.NewEmployeeUsecase(employeeRepo, employeeAreaRepo, employeeOutletRepo, employeeWarehouseRepo, divisionRepo, positionRepo, companyRepo, areaRepo, outletRepo, warehouseRepo, employeeContractRepo, educationHistoryRepo, certificationRepo, assetRepo, signatureRepo, financeAssetRepo, auditLogRepo)
	applicantUC := usecase.NewRecruitmentApplicantUsecase(applicantRepo, applicantStageRepo, applicantActivityRepo, recruitmentRepo, employeeUC)

	// Initialize handlers
	workScheduleHandler := handler.NewWorkScheduleHandler(workScheduleUC)
	holidayHandler := handler.NewHolidayHandler(holidayUC)
	attendanceHandler := handler.NewAttendanceRecordHandler(attendanceUC)
	attendanceWSHandler := handler.NewAttendanceWSHandler(attendanceHub, employeeRepo)
	overtimeHandler := handler.NewOvertimeRequestHandler(overtimeUC, employeeRepo)
	leaveRequestHandler := handler.NewLeaveRequestHandler(leaveRequestUC)
	evaluationGroupHandler := handler.NewEvaluationGroupHandler(evaluationGroupUC)
	evaluationCriteriaHandler := handler.NewEvaluationCriteriaHandler(evaluationCriteriaUC)
	employeeEvaluationHandler := handler.NewEmployeeEvaluationHandler(employeeEvaluationUC)
	recruitmentHandler := handler.NewRecruitmentRequestHandler(recruitmentUC)
	applicantHandler := handler.NewRecruitmentApplicantHandler(applicantUC)
	salaryHandler := handler.NewSalaryStructureHandler(salaryUC)

	// Create HRD group under API with auth middleware
	hrdGroup := api.Group("/hrd")
	hrdGroup.Use(middleware.AuthMiddleware(jwtManager, permService))
	hrdGroup.Use(middleware.ScopeMiddleware(db))

	// Register routes
	router.RegisterWorkScheduleRoutes(hrdGroup, workScheduleHandler)
	router.RegisterHolidayRoutes(hrdGroup, holidayHandler)
	router.RegisterAttendanceRecordRoutes(hrdGroup, attendanceHandler, attendanceWSHandler)
	router.RegisterOvertimeRequestRoutes(hrdGroup, overtimeHandler)
	router.RegisterLeaveRequestRoutes(hrdGroup, leaveRequestHandler)
	router.SetupEvaluationGroupRoutes(hrdGroup, evaluationGroupHandler)
	router.SetupEvaluationCriteriaRoutes(hrdGroup, evaluationCriteriaHandler)
	router.SetupEmployeeEvaluationRoutes(hrdGroup, employeeEvaluationHandler)
	router.SetupRecruitmentRequestRoutes(hrdGroup, recruitmentHandler, applicantHandler)
	router.SetupRecruitmentApplicantRoutes(hrdGroup, applicantHandler)
	router.RegisterSalaryStructureRoutes(hrdGroup, salaryHandler)

	return &HRDDeps{
		HolidayUC:      holidayUC,
		LeaveRequestUC: leaveRequestUC,
		AttendanceUC:   attendanceUC,
		SalaryUC:       salaryUC,
	}
}
