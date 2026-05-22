package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
)

// AssignOutlets handles POST /employees/:id/outlets
func (h *EmployeeHandler) AssignOutlets(c *gin.Context) {
	id := c.Param("id")

	var req dto.AssignEmployeeOutletsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.AssignOutlets(c.Request.Context(), id, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// AssignWarehouses handles POST /employees/:id/warehouses
func (h *EmployeeHandler) AssignWarehouses(c *gin.Context) {
	id := c.Param("id")

	var req dto.AssignEmployeeWarehousesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.AssignWarehouses(c.Request.Context(), id, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// BulkUpdateOutlets handles PUT /employees/:id/outlets
func (h *EmployeeHandler) BulkUpdateOutlets(c *gin.Context) {
	id := c.Param("id")

	var req dto.BulkUpdateEmployeeOutletsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.BulkUpdateOutlets(c.Request.Context(), id, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// BulkUpdateWarehouses handles PUT /employees/:id/warehouses
func (h *EmployeeHandler) BulkUpdateWarehouses(c *gin.Context) {
	id := c.Param("id")

	var req dto.BulkUpdateEmployeeWarehousesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.BulkUpdateWarehouses(c.Request.Context(), id, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}
