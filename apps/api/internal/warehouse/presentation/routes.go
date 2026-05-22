package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/warehouse/data/repositories"
	"github.com/gilabs/gims/api/internal/warehouse/domain/usecase"
	"github.com/gilabs/gims/api/internal/warehouse/presentation/handler"
	"github.com/gilabs/gims/api/internal/warehouse/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all warehouse domain routes
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize layers
	repo := repositories.NewWarehouseRepository(db)
	uc := usecase.NewWarehouseUsecase(repo)
	h := handler.NewWarehouseHandler(uc)

	// Create warehouse group under API with auth middleware
	group := api.Group("/warehouse")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))

	// Register routes
	router.RegisterWarehouseRoutes(group, h)
}
