package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	sysadminRepo "github.com/gilabs/indosupplier/api/internal/sysadmin/data/repositories"
	sysadminMiddleware "github.com/gilabs/indosupplier/api/internal/sysadmin/presentation/middleware"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/presentation/handler"
)

func RegisterWaitingListRoutes(
	rg *gin.RouterGroup,
	h *handler.WaitingListHandler,
	jwtManager *jwt.JWTManager,
	adminRepo sysadminRepo.SystemAdminRepository,
) {
	// Public endpoint for joining the waiting list
	rg.POST("/waiting-list/join", h.Join)

	// Admin endpoints for managing waiting list
	adminGroup := rg.Group("/sysadmin/waiting-list")
	adminGroup.Use(sysadminMiddleware.SysadminAuthMiddleware(jwtManager, adminRepo))
	{
		adminGroup.GET("", h.List)
		adminGroup.PUT("/:id", h.UpdateStatus)
		adminGroup.DELETE("/:id", h.Delete)
	}
}
