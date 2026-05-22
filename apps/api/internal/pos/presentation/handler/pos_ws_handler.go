package handler

import (
	"errors"
	"net/http"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/middleware"
	orgRepo "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	posWS "github.com/gilabs/gims/api/internal/pos/infrastructure/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// POSWSHandler upgrades HTTP connections to WebSocket for real-time POS events.
// The staff client must supply ?outlet_id=<uuid> to subscribe to outlet-specific events.
type POSWSHandler struct {
	hub      *posWS.PosHub
	outletRepo orgRepo.OutletRepository
	upgrader websocket.Upgrader
}

// NewPOSWSHandler creates a POSWSHandler backed by the given hub.
func NewPOSWSHandler(hub *posWS.PosHub, outletRepo orgRepo.OutletRepository) *POSWSHandler {
	return &POSWSHandler{
		hub:        hub,
		outletRepo: outletRepo,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Subscribe upgrades the connection and subscribes it to the given outlet's POS events.
// Query param: outlet_id (required, uuid)
func (h *POSWSHandler) Subscribe(c *gin.Context) {
	outletID := c.Query("outlet_id")
	if outletID == "" {
		coreErrors.HandleValidationError(c, nil)
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" && !middleware.IsSystemAdmin(c.Request.Context()) {
		coreErrors.UnauthorizedResponse(c, "tenant session required")
		return
	}

	outlet, err := h.outletRepo.GetByID(c.Request.Context(), outletID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	if outlet == nil || outlet.ID == "" {
		coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
		return
	}
	if !middleware.IsSystemAdmin(c.Request.Context()) && outlet.TenantID != tenantID {
		coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
		return
	}
	if tenantID == "" {
		tenantID = outlet.TenantID
	}

	allowedOutletIDs, err := usecase.ResolveScopedPOSOutletIDs(c.Request.Context(), h.outletRepo)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	if allowedOutletIDs != nil && !usecase.IsOutletAllowed(allowedOutletIDs, outletID) {
		coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}

	clientID := h.hub.Register(conn, tenantID, outletID)
	defer h.hub.Unregister(clientID)

	// Block until the client disconnects (reads are ignored — this is server-push only).
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
