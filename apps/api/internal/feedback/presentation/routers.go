package presentation

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	feedbackRepos "github.com/gilabs/gims/api/internal/feedback/data/repositories"
	feedbackUsecase "github.com/gilabs/gims/api/internal/feedback/domain/usecase"
	feedbackHandler "github.com/gilabs/gims/api/internal/feedback/presentation/handler"
	feedbackRouter "github.com/gilabs/gims/api/internal/feedback/presentation/router"
	orgRepo "github.com/gilabs/gims/api/internal/organization/data/repositories"
)

// FeedbackDeps holds exported dependencies that other domains may reference.
type FeedbackDeps struct {
	Usecase feedbackUsecase.FeedbackUsecase
}

// RegisterRoutes wires the feedback domain and registers routes.
func RegisterRoutes(
	r *gin.Engine,
	api *gin.RouterGroup,
	db *gorm.DB,
	jwtManager *jwt.JWTManager,
	permService interface {
		GetPermissions(roleCode string) ([]string, error)
		GetPermissionsWithScope(roleCode string) (map[string]string, error)
	},
) FeedbackDeps {
	// ─── Repositories ────────────────────────────────────────────────────────
	formRepo := feedbackRepos.NewFeedbackFormRepository(db)
	tokenRepo := feedbackRepos.NewFeedbackTokenRepository(db)
	responseRepo := feedbackRepos.NewFeedbackResponseRepository(db)
	outletRepo := orgRepo.NewOutletRepository(db)

	// ─── Usecase ─────────────────────────────────────────────────────────────
	uc := feedbackUsecase.NewFeedbackUsecase(formRepo, tokenRepo, responseRepo, outletRepo)

	// ─── Handler ──────────────────────────────────────────────────────────────
	h := feedbackHandler.NewFeedbackHandler(uc)

	// ─── Routes ───────────────────────────────────────────────────────────────

	// Protected routes (require auth + permissions)
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(jwtManager, permService))
	protected.Use(middleware.ScopeMiddleware(db))
	feedbackRouter.RegisterFeedbackRoutes(protected, h)

	// Public routes (no auth — customer feedback submission)
	feedbackRouter.RegisterFeedbackPublicRoutes(api, h)

	return FeedbackDeps{Usecase: uc}
}
