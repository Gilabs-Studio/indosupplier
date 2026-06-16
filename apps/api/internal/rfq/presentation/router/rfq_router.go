package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/rfq/presentation/handler"
)

func RegisterRFQRoutes(rg *gin.RouterGroup, h *handler.RFQHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/rfqs")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.GetByID)
		g.GET("/:id/bids", h.GetBids)
		g.POST("/:id/bids/:bidId/accept", h.AcceptBid)
	}
}
