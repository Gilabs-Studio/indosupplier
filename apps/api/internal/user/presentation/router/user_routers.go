package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/user/presentation/handler"
)

func RegisterUserRoutes(rg *gin.RouterGroup, h *handler.UserHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/users")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("/available", h.GetAvailable)
		g.GET("/limit", h.GetLimit)

		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}

	p := rg.Group("/profile")
	p.Use(middleware.AuthMiddleware(jwtManager))
	{
		p.PUT("", h.UpdateProfile)
		p.PUT("/password", h.ChangePassword)
		p.POST("/avatar", h.UploadAvatar)
		p.POST("/delete-account", h.RequestAccountDeletion)
	}
}
