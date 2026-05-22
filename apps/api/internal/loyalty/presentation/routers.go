package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	"github.com/gilabs/gims/api/internal/loyalty/data/repositories"
	"github.com/gilabs/gims/api/internal/loyalty/domain/usecase"
	"github.com/gilabs/gims/api/internal/loyalty/presentation/handler"
	"github.com/gilabs/gims/api/internal/loyalty/presentation/router"
	organizationRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoyaltyDeps exposes the loyalty usecase for injection into other domains (e.g. POS payment).
type LoyaltyDeps struct {
	Usecase usecase.LoyaltyUsecase
}

func RegisterRoutes(
	r *gin.Engine,
	api *gin.RouterGroup,
	db *gorm.DB,
	jwtManager *jwt.JWTManager,
	permService interface {
		GetPermissions(roleCode string) ([]string, error)
		GetPermissionsWithScope(roleCode string) (map[string]string, error)
	},
	customerRepo customerRepos.CustomerRepository,
) LoyaltyDeps {
	programRepo := repositories.NewLoyaltyProgramRepository(db)
	memberRepo := repositories.NewLoyaltyMemberRepository(db)
	ledgerRepo := repositories.NewLoyaltyPointLedgerRepository(db)
	outletRepo := organizationRepos.NewOutletRepository(db)

	uc := usecase.NewLoyaltyUsecase(db, programRepo, memberRepo, ledgerRepo, customerRepo, outletRepo)
	h := handler.NewLoyaltyHandler(uc)

	group := api.Group("/")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	router.RegisterLoyaltyRoutes(group, h)

	// Public endpoint — no auth required.
	router.RegisterLoyaltyPublicRoutes(api, h)

	return LoyaltyDeps{Usecase: uc}
}
