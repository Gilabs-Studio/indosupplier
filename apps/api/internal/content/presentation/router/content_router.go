package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/content/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	sysadminRepo "github.com/gilabs/indosupplier/api/internal/sysadmin/data/repositories"
	sysadminMiddleware "github.com/gilabs/indosupplier/api/internal/sysadmin/presentation/middleware"
)

func RegisterContentRoutes(rg *gin.RouterGroup, h *handler.ContentHandler, jwtManager *jwt.JWTManager, adminRepo sysadminRepo.SystemAdminRepository) {
	public := rg.Group("/content/articles")
	{
		public.GET("", h.ListPublic)
		public.GET("/:slug", h.GetPublicBySlug)
	}

	admin := rg.Group("/sysadmin/content/articles")
	admin.Use(sysadminMiddleware.SysadminAuthMiddleware(jwtManager, adminRepo))
	{
		admin.GET("", h.ListAdmin)
		admin.GET("/:id", h.GetAdminByID)
		admin.POST("", h.Create)
		admin.PUT("/:id", h.Update)
		admin.DELETE("/:id", h.Delete)
	}
}
