package router

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/handler"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterUploadRoutes(rg *gin.RouterGroup, jwtManager *jwt.JWTManager, permissionService security.PermissionService) {
	h := handler.NewUploadHandler()

	upload := rg.Group("/upload")
	upload.Use(middleware.AuthMiddleware(jwtManager, permissionService))
	upload.Use(middleware.RateLimitMiddleware("upload"))
	{
		upload.POST("/image", h.UploadImage)
		upload.POST("/document", h.UploadDocument)
	}
}
