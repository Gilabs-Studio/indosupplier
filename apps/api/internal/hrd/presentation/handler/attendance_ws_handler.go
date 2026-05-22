package handler

import (
	"net/http"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/middleware"
	hrdWS "github.com/gilabs/gims/api/internal/hrd/infrastructure/ws"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// AttendanceWSHandler upgrades HTTP connections to WebSocket for real-time attendance events.
type AttendanceWSHandler struct {
	hub          *hrdWS.AttendanceHub
	employeeRepo orgRepos.EmployeeRepository
	upgrader     websocket.Upgrader
}

func NewAttendanceWSHandler(hub *hrdWS.AttendanceHub, employeeRepo orgRepos.EmployeeRepository) *AttendanceWSHandler {
	return &AttendanceWSHandler{
		hub:          hub,
		employeeRepo: employeeRepo,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// SubscribeToday upgrades the connection and subscribes the client to tenant attendance updates.
// This endpoint is server-push only.
func (h *AttendanceWSHandler) SubscribeToday(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" || middleware.IsSystemAdmin(c.Request.Context()) {
		coreErrors.UnauthorizedResponse(c, "tenant session required")
		return
	}

	// Resolve employee context for consistency with attendance self-service flows.
	// The value is currently not used for filtering, but we validate availability.
	if _, ok := h.resolveEmployeeID(c); !ok {
		coreErrors.UnauthorizedResponse(c, "user context is required")
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

func (h *AttendanceWSHandler) resolveEmployeeID(c *gin.Context) (string, bool) {
	if employeeID, ok := c.Get("employee_id"); ok {
		if id, castOK := employeeID.(string); castOK && id != "" {
			return id, true
		}
	}

	userID, ok := c.Get("user_id")
	if !ok {
		return "", false
	}

	uid, castOK := userID.(string)
	if !castOK || uid == "" {
		return "", false
	}

	emp, err := h.employeeRepo.FindByUserID(c.Request.Context(), uid)
	if err == nil && emp != nil && emp.ID != "" {
		return emp.ID, true
	}

	return uid, true
}
