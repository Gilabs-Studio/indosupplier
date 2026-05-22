package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BusinessUnitHandler handles business unit HTTP requests
type BusinessUnitHandler struct {
	businessUnitUC usecase.BusinessUnitUsecase
}

// NewBusinessUnitHandler creates a new BusinessUnitHandler
func NewBusinessUnitHandler(businessUnitUC usecase.BusinessUnitUsecase) *BusinessUnitHandler {
	return &BusinessUnitHandler{businessUnitUC: businessUnitUC}
}

func (h *BusinessUnitHandler) List(c *gin.Context) {
	var req dto.ListBusinessUnitsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	businessUnits, pagination, err := h.businessUnitUC.List(c.Request.Context(), &req)
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

	response.SuccessResponse(c, businessUnits, meta)
}

func (h *BusinessUnitHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	businessUnit, err := h.businessUnitUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrBusinessUnitNotFound {
			errors.ErrorResponse(c, "BUSINESS_UNIT_NOT_FOUND", map[string]interface{}{
				"business_unit_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, businessUnit, nil)
}

func (h *BusinessUnitHandler) Create(c *gin.Context) {
	var req dto.CreateBusinessUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	businessUnit, err := h.businessUnitUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrBusinessUnitAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "business_unit",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, businessUnit, nil)
}

func (h *BusinessUnitHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateBusinessUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	businessUnit, err := h.businessUnitUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrBusinessUnitNotFound {
			errors.ErrorResponse(c, "BUSINESS_UNIT_NOT_FOUND", map[string]interface{}{
				"business_unit_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrBusinessUnitAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "business_unit",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, businessUnit, nil)
}

func (h *BusinessUnitHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.businessUnitUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrBusinessUnitNotFound {
			errors.ErrorResponse(c, "BUSINESS_UNIT_NOT_FOUND", map[string]interface{}{
				"business_unit_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "business_unit", id, nil)
}
