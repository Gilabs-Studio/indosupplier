package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/supplier/presentation/handler"
)

func RegisterSupplierReviewRoutes(rg *gin.RouterGroup, h *handler.SupplierReviewHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/supplier/reviews")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.GetReviews)
		g.POST("/:id/reply", h.ReplyReview)
	}
}
