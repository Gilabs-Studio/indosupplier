package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/notification/data/repositories"
	"github.com/gilabs/gims/api/internal/notification/domain/usecase"
	notificationWS "github.com/gilabs/gims/api/internal/notification/infrastructure/ws"
	"github.com/gilabs/gims/api/internal/notification/presentation/handler"
	"github.com/gilabs/gims/api/internal/notification/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(_ *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}) {
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationUC := usecase.NewNotificationUsecase(notificationRepo)
	notificationH := handler.NewNotificationHandler(notificationUC, notificationWS.DefaultNotificationHub())

	group := api.Group("")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	router.RegisterNotificationRoutes(group, notificationH)
}
