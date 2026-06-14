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

type BookmarkHandler struct {
	usecase usecase.BookmarkUsecase
}

func NewBookmarkHandler(uc usecase.BookmarkUsecase) *BookmarkHandler {
	return &BookmarkHandler{usecase: uc}
}

func (h *BookmarkHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *BookmarkHandler) Create(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.CreateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	bookmark, err := h.usecase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrSupplierProfileNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrBookmarkAlreadyExists) {
			errors.ErrorResponse(c, "BOOKMARK_ALREADY_EXISTS", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, bookmark, &response.Meta{CreatedBy: userID})
}

func (h *BookmarkHandler) Delete(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	err := h.usecase.Delete(c.Request.Context(), userID, id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrBookmarkNotFound) {
			errors.ErrorResponse(c, "BOOKMARK_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"deleted": true, "id": id}, nil)
}

func (h *BookmarkHandler) List(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	bookmarks, err := h.usecase.List(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, bookmarks, nil)
}
