package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/general/data/repositories"
	"github.com/gilabs/gims/api/internal/general/domain/usecase"
	generalWS "github.com/gilabs/gims/api/internal/general/infrastructure/ws"
	"github.com/gilabs/gims/api/internal/general/presentation/handler"
	"github.com/gilabs/gims/api/internal/general/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all general domain routes
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize repositories
	dashboardRepo := repositories.NewDashboardRepository(db)
	layoutRepo := repositories.NewDashboardLayoutRepository(db)
	onboardingRepo := repositories.NewOnboardingRepository(db)

	// Initialize usecases
	dashboardUC := usecase.NewDashboardUsecase(dashboardRepo, layoutRepo)
	onboardingUC := usecase.NewOnboardingUsecase(onboardingRepo)
	onboardingHub := generalWS.DefaultOnboardingHub()
	onboardingUC = usecase.WithOnboardingPublisher(onboardingUC, onboardingHub)

	// Initialize handlers
	dashboardHandler := handler.NewDashboardHandler(dashboardUC)
	onboardingHandler := handler.NewOnboardingHandler(onboardingUC)
	onboardingWSHandler := handler.NewOnboardingWSHandler(onboardingHub)

	// Create group with auth middleware
	group := api.Group("/general")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	group.Use(middleware.ScopeMiddleware(db))

	// Register routes
	router.RegisterDashboardRoutes(group, dashboardHandler, db)
	router.RegisterOnboardingRoutes(group, onboardingHandler, onboardingWSHandler)
}
