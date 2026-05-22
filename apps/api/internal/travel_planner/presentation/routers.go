package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/travel_planner/data/repositories"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/mapper"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/usecase"
	"github.com/gilabs/gims/api/internal/travel_planner/presentation/handler"
	"github.com/gilabs/gims/api/internal/travel_planner/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	planRepo := repositories.NewTravelPlanRepository(db)
	planMapper := mapper.NewTravelPlanMapper()
	planUC := usecase.NewTravelPlanUsecase(db, planRepo, planMapper)
	planHandler := handler.NewTravelPlanHandler(planUC)

	group := api.Group("/travel")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	group.Use(middleware.ScopeMiddleware(db))

	router.RegisterTravelPlanRoutes(group, planHandler)

	wsGroup := r.Group("/ws")
	wsGroup.Use(middleware.AuthMiddleware(jwtManager, permService))
	wsGroup.Use(middleware.ScopeMiddleware(db))
	router.RegisterTravelPlannerWebSocketRoutes(wsGroup, planHandler)
}
