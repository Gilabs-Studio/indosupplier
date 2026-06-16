package handler

import (
	stderrors "errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/usecase"
	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
)

type ReviewHandler struct {
	usecase usecase.ReviewUsecase
}

func NewReviewHandler(uc usecase.ReviewUsecase) *ReviewHandler {
	return &ReviewHandler{usecase: uc}
}

func (h *ReviewHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *ReviewHandler) ListEligible(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	items, err := h.usecase.ListEligible(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, items, nil)
}

func (h *ReviewHandler) ListHistory(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	items, err := h.usecase.ListHistory(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, items, nil)
}

func (h *ReviewHandler) Create(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	review, err := h.usecase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrTransactionNotFound) {
			errors.ErrorResponse(c, "TRANSACTION_NOT_FOUND", map[string]interface{}{"purchaseOrderId": req.PurchaseOrderID}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrTransactionNotCompleted) {
			errors.ErrorResponse(c, "TRANSACTION_NOT_COMPLETED", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrReviewAlreadyExists) {
			errors.ErrorResponse(c, "REVIEW_ALREADY_EXISTS", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, review, &response.Meta{CreatedBy: userID})
}
