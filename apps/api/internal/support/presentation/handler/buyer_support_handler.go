package handler

import (
	stderrors "errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/support/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/support/domain/usecase"
)

type BuyerSupportHandler struct {
	usecase usecase.SupportUsecase
}

func NewBuyerSupportHandler(uc usecase.SupportUsecase) *BuyerSupportHandler {
	return &BuyerSupportHandler{usecase: uc}
}

func (h *BuyerSupportHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *BuyerSupportHandler) ListTickets(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	tickets, err := h.usecase.ListBuyerTickets(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, tickets, nil)
}

func (h *BuyerSupportHandler) GetTicketDetail(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	ticketNumber := c.Param("id")
	ticket, err := h.usecase.GetBuyerTicketDetail(c.Request.Context(), userID, ticketNumber)
	if err != nil {
		switch {
		case stderrors.Is(err, usecase.ErrBuyerProfileNotFound):
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
		case stderrors.Is(err, usecase.ErrSupportTicketNotFound):
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"ticket_number": ticketNumber}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, ticket, nil)
}

func (h *BuyerSupportHandler) CreateTicket(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.CreateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	ticket, err := h.usecase.CreateBuyerTicket(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, ticket, &response.Meta{CreatedBy: userID})
}

func (h *BuyerSupportHandler) ReplyTicket(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	ticketNumber := c.Param("id")
	var req dto.CreateSupportReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	message, err := h.usecase.ReplyBuyerTicket(c.Request.Context(), userID, ticketNumber, &req)
	if err != nil {
		switch {
		case stderrors.Is(err, usecase.ErrBuyerProfileNotFound):
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
		case stderrors.Is(err, usecase.ErrSupportTicketNotFound):
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"ticket_number": ticketNumber}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponseCreated(c, message, &response.Meta{UpdatedBy: userID})
}
