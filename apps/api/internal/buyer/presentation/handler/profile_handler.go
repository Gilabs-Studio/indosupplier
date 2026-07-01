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

type ProfileHandler struct {
	usecase usecase.ProfileUsecase
}

func NewProfileHandler(uc usecase.ProfileUsecase) *ProfileHandler {
	return &ProfileHandler{usecase: uc}
}

func (h *ProfileHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}
	profile, err := h.usecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, profile, nil)
}

func (h *ProfileHandler) UpdatePersonal(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}
	var req dto.UpdateBuyerPersonalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}
	profile, err := h.usecase.UpdatePersonal(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, profile, &response.Meta{UpdatedBy: userID})
}

func (h *ProfileHandler) UpdateCompany(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}
	var req dto.UpdateBuyerCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}
	profile, err := h.usecase.UpdateCompany(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, profile, &response.Meta{UpdatedBy: userID})
}

func (h *ProfileHandler) ListDocuments(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}
	documents, err := h.usecase.ListDocuments(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, documents, nil)
}

func (h *ProfileHandler) UploadDocument(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}
	var req dto.UploadBuyerDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}
	document, err := h.usecase.UploadDocument(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponseCreated(c, document, &response.Meta{CreatedBy: userID})
}
