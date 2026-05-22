package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BusinessTypeHandler handles business type HTTP requests
type BusinessTypeHandler struct {
	businessTypeUC usecase.BusinessTypeUsecase
}

// NewBusinessTypeHandler creates a new BusinessTypeHandler
func NewBusinessTypeHandler(businessTypeUC usecase.BusinessTypeUsecase) *BusinessTypeHandler {
	return &BusinessTypeHandler{businessTypeUC: businessTypeUC}
}

func (h *BusinessTypeHandler) List(c *gin.Context) {
	var req dto.ListBusinessTypesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	businessTypes, pagination, err := h.businessTypeUC.List(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      pagination.Total,
			TotalPages: pagination.TotalPages,
			HasNext:    pagination.Page < pagination.TotalPages,
			HasPrev:    pagination.Page > 1,
		},
	}

	response.SuccessResponse(c, businessTypes, meta)
}

func (h *BusinessTypeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	businessType, err := h.businessTypeUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrBusinessTypeNotFound {
			errors.ErrorResponse(c, "BUSINESS_TYPE_NOT_FOUND", map[string]interface{}{
				"business_type_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, businessType, nil)
}

func (h *BusinessTypeHandler) Create(c *gin.Context) {
	var req dto.CreateBusinessTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	businessType, err := h.businessTypeUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrBusinessTypeAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "business_type",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, businessType, nil)
}

func (h *BusinessTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateBusinessTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	businessType, err := h.businessTypeUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrBusinessTypeNotFound {
			errors.ErrorResponse(c, "BUSINESS_TYPE_NOT_FOUND", map[string]interface{}{
				"business_type_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrBusinessTypeAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "business_type",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, businessType, nil)
}

func (h *BusinessTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.businessTypeUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrBusinessTypeNotFound {
			errors.ErrorResponse(c, "BUSINESS_TYPE_NOT_FOUND", map[string]interface{}{
				"business_type_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "business_type", id, nil)
}
