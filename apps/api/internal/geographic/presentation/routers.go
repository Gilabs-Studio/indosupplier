package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/geographic/data/repositories"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gilabs/gims/api/internal/geographic/presentation/handler"
	"github.com/gilabs/gims/api/internal/geographic/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all geographic domain routes
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	// Initialize repositories
	countryRepo := repositories.NewCountryRepository(db)
	provinceRepo := repositories.NewProvinceRepository(db)
	cityRepo := repositories.NewCityRepository(db)
	districtRepo := repositories.NewDistrictRepository(db)
	villageRepo := repositories.NewVillageRepository(db)
	mapDataRepo := repositories.NewMapDataRepository(db)

	// Initialize usecases
	countryUC := usecase.NewCountryUsecase(countryRepo)
	provinceUC := usecase.NewProvinceUsecase(provinceRepo, countryRepo)
	cityUC := usecase.NewCityUsecase(cityRepo, provinceRepo)
	districtUC := usecase.NewDistrictUsecase(districtRepo, cityRepo)
	villageUC := usecase.NewVillageUsecase(villageRepo, districtRepo)
	mapDataUC := usecase.NewMapDataUsecase(mapDataRepo)

	// Initialize handlers
	countryH := handler.NewCountryHandler(countryUC)
	provinceH := handler.NewProvinceHandler(provinceUC)
	cityH := handler.NewCityHandler(cityUC)
	districtH := handler.NewDistrictHandler(districtUC)
	villageH := handler.NewVillageHandler(villageUC)
	mapDataH := handler.NewMapDataHandler(mapDataUC)

	// Create geographic group under API
	group := api.Group("/geographic")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))

	// Register routes - map-data first for route specificity
	router.RegisterMapDataRoutes(group, mapDataH)
	router.RegisterCountryRoutes(group, countryH)
	router.RegisterProvinceRoutes(group, provinceH)
	router.RegisterCityRoutes(group, cityH)
	router.RegisterDistrictRoutes(group, districtH)
	router.RegisterVillageRoutes(group, villageH)
}
