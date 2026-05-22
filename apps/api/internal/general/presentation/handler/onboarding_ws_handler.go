package handler

import (
	"net/http"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/middleware"
	generalWS "github.com/gilabs/gims/api/internal/general/infrastructure/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// OnboardingWSHandler upgrades HTTP connections to WebSocket for onboarding state updates.
type OnboardingWSHandler struct {
	hub      *generalWS.OnboardingHub
	upgrader websocket.Upgrader
}

func NewOnboardingWSHandler(hub *generalWS.OnboardingHub) *OnboardingWSHandler {
	return &OnboardingWSHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Subscribe upgrades the connection and subscribes it to onboarding updates for the active tenant.
func (h *OnboardingWSHandler) Subscribe(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" || middleware.IsSystemAdmin(c.Request.Context()) {
		coreErrors.UnauthorizedResponse(c, "tenant session required")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}

	clientID := h.hub.Register(conn, tenantID)
	defer h.hub.Unregister(clientID)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
