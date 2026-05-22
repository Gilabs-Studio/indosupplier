package router

import (
	"github.com/gilabs/gims/api/internal/general/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterOnboardingRoutes registers tenant onboarding state routes.
func RegisterOnboardingRoutes(rg *gin.RouterGroup, h *handler.OnboardingHandler, wsH *handler.OnboardingWSHandler) {
	g := rg.Group("/onboarding")
	{
		g.GET("/ws", wsH.Subscribe)
		g.GET("", h.GetState)
		g.PUT("/business-type", h.SetBusinessType)
		g.PUT("/complete", h.MarkComplete)
	}
}
