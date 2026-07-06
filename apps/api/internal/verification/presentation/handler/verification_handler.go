package handler

import (
	stderrors "errors"

	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/verification/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/verification/domain/usecase"
)

type VerificationHandler struct {
	usecase usecase.VerificationUsecase
}

func NewVerificationHandler(uc usecase.VerificationUsecase) *VerificationHandler {
	return &VerificationHandler{usecase: uc}
}

func (h *VerificationHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *VerificationHandler) GetVerificationData(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	data, err := h.usecase.GetVerificationData(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrSupplierProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, data, nil)
}

func (h *VerificationHandler) UpdateVerificationData(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	data, err := h.usecase.UpdateVerificationData(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrSupplierProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, data, &response.Meta{UpdatedBy: userID})
}

func (h *VerificationHandler) SubmitVerification(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	data, err := h.usecase.SubmitVerification(c.Request.Context(), userID)
	if err != nil {
		switch {
		case stderrors.Is(err, usecase.ErrSupplierProfileNotFound):
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
		case stderrors.Is(err, usecase.ErrVerificationNotFound):
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID, "resource": "verification_request"}, nil)
		case stderrors.Is(err, usecase.ErrVerificationIncomplete):
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"reason": "verification_incomplete"}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, data, &response.Meta{UpdatedBy: userID})
}
