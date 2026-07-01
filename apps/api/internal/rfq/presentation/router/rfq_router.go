package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/rfq/presentation/handler"
)

func RegisterRFQRoutes(rg *gin.RouterGroup, h *handler.RFQHandler, jwtManager *jwt.JWTManager) {
	buyerGroup := rg.Group("/buyer/rfqs")
	buyerGroup.Use(middleware.AuthMiddleware(jwtManager))
	{
		buyerGroup.GET("", h.List)
		buyerGroup.POST("", h.Create)
		buyerGroup.GET("/:id", h.GetByID)
		buyerGroup.GET("/:id/bids", h.GetBids)
		buyerGroup.POST("/:id/bids/:bidId/accept", h.AcceptBid)
	}

	supplierGroup := rg.Group("/supplier/rfqs")
	supplierGroup.Use(middleware.AuthMiddleware(jwtManager))
	{
		supplierGroup.GET("", h.ListForSupplier)
		supplierGroup.GET("/:id", h.GetSupplierRFQByID)
		supplierGroup.POST("/:id/proposals", h.SubmitProposal)
	}
}
