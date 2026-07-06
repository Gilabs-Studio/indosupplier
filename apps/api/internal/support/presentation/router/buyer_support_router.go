package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/support/presentation/handler"
)

func RegisterBuyerSupportRoutes(rg *gin.RouterGroup, h *handler.BuyerSupportHandler, jwtManager *jwt.JWTManager) {
	g := rg.Group("/buyer/support/tickets")
	g.Use(middleware.AuthMiddleware(jwtManager))
	{
		g.GET("", h.ListTickets)
		g.POST("", h.CreateTicket)
		g.GET("/:id", h.GetTicketDetail)
		g.POST("/:id/messages", h.ReplyTicket)
	}
}
