package handler

import (
	"net/http"
	"strconv"
	"strings"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/notification/domain/usecase"
	notificationWS "github.com/gilabs/gims/api/internal/notification/infrastructure/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const authRequiredMessage = "Authentication required"

type NotificationHandler struct {
	usecase  usecase.NotificationUsecase
	hub      *notificationWS.NotificationHub
	upgrader websocket.Upgrader
}

func NewNotificationHandler(usecase usecase.NotificationUsecase, hub *notificationWS.NotificationHub) *NotificationHandler {
	return &NotificationHandler{
		usecase: usecase,
		hub:     hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *NotificationHandler) SubscribeWS(c *gin.Context) {
	userID, ok := getValidatedUserID(c)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", authRequiredMessage, nil, nil)
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" && !middleware.IsSystemAdmin(c.Request.Context()) {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "tenant session required", nil, nil)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}

	clientID := h.hub.Register(conn, tenantID, userID)
	defer h.hub.Unregister(clientID)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userIDStr, ok := getValidatedUserID(c)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", authRequiredMessage, nil, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	notifType := strings.TrimSpace(c.Query("type"))
	entityType := strings.TrimSpace(c.Query("entity"))
	isReadParam := strings.TrimSpace(c.Query("is_read"))
	var isRead *bool
	if isReadParam != "" {
		parsed, err := strconv.ParseBool(isReadParam)
		if err != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid is_read value", map[string]interface{}{"is_read": isReadParam}, nil)
			return
		}
		isRead = &parsed
	}

	items, total, err := h.usecase.List(c.Request.Context(), userIDStr, page, perPage, notifType, entityType, isRead)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch notifications", nil, nil)
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	meta.Filters = map[string]interface{}{"type": notifType, "entity": entityType, "is_read": isRead}
	response.SuccessResponse(c, items, meta)
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userIDStr, ok := getValidatedUserID(c)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", authRequiredMessage, nil, nil)
		return
	}

	res, err := h.usecase.GetUnreadCount(c.Request.Context(), userIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch unread count", nil, nil)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userIDStr, ok := getValidatedUserID(c)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", authRequiredMessage, nil, nil)
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "notification id is required", nil, nil)
		return
	}

	item, err := h.usecase.MarkAsRead(c.Request.Context(), userIDStr, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "notification not found", nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to mark notification as read", nil, nil)
		return
	}

	response.SuccessResponse(c, item, nil)
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userIDStr, ok := getValidatedUserID(c)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", authRequiredMessage, nil, nil)
		return
	}

	res, err := h.usecase.MarkAllAsRead(c.Request.Context(), userIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to mark all notifications as read", nil, nil)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func getValidatedUserID(c *gin.Context) (string, bool) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userID, ok := userIDRaw.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", false
	}

	if _, err := uuid.Parse(userID); err != nil {
		return "", false
	}

	return userID, true
}
