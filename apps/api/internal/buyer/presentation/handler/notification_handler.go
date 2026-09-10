package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/buyer/domain/usecase"
	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
)

type NotificationHandler struct {
	usecase usecase.NotificationUsecase
}

func NewNotificationHandler(uc usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{usecase: uc}
}

func (h *NotificationHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return "", false
	}
	userID, ok := userIDVal.(string)
	if !ok {
		errors.InternalServerErrorResponse(c, "invalid user ID structure in context")
		return "", false
	}
	return userID, true
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	notifs, err := h.usecase.ListNotifications(c.Request.Context(), userID)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, notifs, nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	if err := h.usecase.MarkAllRead(c.Request.Context(), userID); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"message": "All notifications marked as read"}, nil)
}
