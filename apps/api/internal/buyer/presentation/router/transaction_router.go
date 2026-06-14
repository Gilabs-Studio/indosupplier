package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/presentation/handler"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
)

func RegisterTransactionRoutes(rg *gin.RouterGroup, h *handler.TransactionHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/transactions")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.POST("", h.Create)
	}
}
