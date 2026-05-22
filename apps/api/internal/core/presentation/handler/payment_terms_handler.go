package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/core/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentTermsHandler struct {
	uc usecase.PaymentTermsUsecase
}

func NewPaymentTermsHandler(uc usecase.PaymentTermsUsecase) *PaymentTermsHandler {
	return &PaymentTermsHandler{uc: uc}
}

func (h *PaymentTermsHandler) Create(c *gin.Context) {
	var req dto.CreatePaymentTermsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

func (h *PaymentTermsHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "PAYMENT_TERMS_NOT_FOUND", map[string]interface{}{"id": id}, nil)
		return
	}
	response.SuccessResponse(c, result, nil)
}

func (h *PaymentTermsHandler) List(c *gin.Context) {
	params := repositories.ListParams{
		Search:  c.Query("search"),
		SortBy:  c.DefaultQuery("sort_by", "name"),
		SortDir: c.DefaultQuery("sort_dir", "asc"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page: page, PerPage: perPage, Total: int(total), TotalPages: totalPages,
			HasNext: page < totalPages, HasPrev: page > 1,
		},
	}

	response.SuccessResponse(c, results, meta)
}

func (h *PaymentTermsHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdatePaymentTermsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Update(c.Request.Context(), id, req)
	if err != nil {
		errors.ErrorResponse(c, "PAYMENT_TERMS_NOT_FOUND", map[string]interface{}{"id": id}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

func (h *PaymentTermsHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "PAYMENT_TERMS_NOT_FOUND", map[string]interface{}{"id": id}, nil)
		return
	}
	response.SuccessResponseDeleted(c, "payment_terms", id, nil)
}
