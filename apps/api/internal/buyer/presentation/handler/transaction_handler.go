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

type TransactionHandler struct {
	usecase usecase.TransactionUsecase
}

func NewTransactionHandler(uc usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{usecase: uc}
}

func (h *TransactionHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *TransactionHandler) Create(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	tx, err := h.usecase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrSupplierProfileNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, tx, &response.Meta{CreatedBy: userID})
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	tx, err := h.usecase.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrTransactionNotFound) {
			errors.ErrorResponse(c, "TRANSACTION_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, tx, nil)
}

func (h *TransactionHandler) List(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.ListTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	txs, pagination, err := h.usecase.List(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: &response.PaginationMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      int(pagination.Total),
		TotalPages: pagination.TotalPages,
		HasNext:    pagination.Page < pagination.TotalPages,
		HasPrev:    pagination.Page > 1,
	}, Filters: map[string]interface{}{}}

	if req.Status != "" {
		meta.Filters["status"] = req.Status
	}

	response.SuccessResponse(c, txs, meta)
}
