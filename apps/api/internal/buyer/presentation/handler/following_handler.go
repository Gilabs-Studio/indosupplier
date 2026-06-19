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

type FollowingHandler struct {
	usecase usecase.FollowingUsecase
}

func NewFollowingHandler(uc usecase.FollowingUsecase) *FollowingHandler {
	return &FollowingHandler{usecase: uc}
}

func (h *FollowingHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *FollowingHandler) Create(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.CreateFollowingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	following, err := h.usecase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrSupplierProfileNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrFollowingAlreadyExists) {
			errors.ErrorResponse(c, "SUPPLIER_ALREADY_FOLLOWED", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, following, &response.Meta{CreatedBy: userID})
}

func (h *FollowingHandler) Delete(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	supplierProfileID := c.Param("supplierProfileId")
	err := h.usecase.Delete(c.Request.Context(), userID, supplierProfileID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrFollowingNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_FOLLOWING_NOT_FOUND", map[string]interface{}{"supplierProfileId": supplierProfileID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"deleted": true, "supplierProfileId": supplierProfileID}, nil)
}

func (h *FollowingHandler) List(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	followings, err := h.usecase.List(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, followings, nil)
}
