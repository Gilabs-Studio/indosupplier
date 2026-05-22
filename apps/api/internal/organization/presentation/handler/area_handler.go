package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AreaHandler handles area HTTP requests
type AreaHandler struct {
	areaUC usecase.AreaUsecase
}

// NewAreaHandler creates a new AreaHandler
func NewAreaHandler(areaUC usecase.AreaUsecase) *AreaHandler {
	return &AreaHandler{areaUC: areaUC}
}

func (h *AreaHandler) List(c *gin.Context) {
	var req dto.ListAreasRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	areas, pagination, err := h.areaUC.List(c.Request.Context(), &req)
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

	response.SuccessResponse(c, areas, meta)
}

func (h *AreaHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	area, err := h.areaUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrAreaNotFound {
			errors.ErrorResponse(c, "AREA_NOT_FOUND", map[string]interface{}{
				"area_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, area, nil)
}

func (h *AreaHandler) Create(c *gin.Context) {
	var req dto.CreateAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	area, err := h.areaUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrAreaAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "area",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, area, nil)
}

func (h *AreaHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	area, err := h.areaUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrAreaNotFound {
			errors.ErrorResponse(c, "AREA_NOT_FOUND", map[string]interface{}{
				"area_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrAreaAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "area",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, area, nil)
}

func (h *AreaHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.areaUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrAreaNotFound {
			errors.ErrorResponse(c, "AREA_NOT_FOUND", map[string]interface{}{
				"area_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrAreaHasAssignedEmployees {
			errors.ErrorResponse(c, "RESOURCE_IN_USE", map[string]interface{}{
				"resource": "area",
				"reason":   "has employees assigned",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "area", id, nil)
}

// GetDetail handles GET /areas/:id/detail and returns full area info with supervisor/member lists.
func (h *AreaHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")

	area, err := h.areaUC.GetByIDWithDetails(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrAreaNotFound {
			errors.ErrorResponse(c, "AREA_NOT_FOUND", map[string]interface{}{
				"area_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, area, nil)
}

// AssignSupervisors handles POST /areas/:id/supervisors
func (h *AreaHandler) AssignSupervisors(c *gin.Context) {
	id := c.Param("id")

	var req dto.AssignAreaSupervisorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	area, err := h.areaUC.AssignSupervisors(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrAreaNotFound {
			errors.ErrorResponse(c, "AREA_NOT_FOUND", map[string]interface{}{
				"area_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, area, nil)
}

// AssignMembers handles POST /areas/:id/members
func (h *AreaHandler) AssignMembers(c *gin.Context) {
	id := c.Param("id")

	var req dto.AssignAreaMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	area, err := h.areaUC.AssignMembers(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrAreaNotFound {
			errors.ErrorResponse(c, "AREA_NOT_FOUND", map[string]interface{}{
				"area_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, area, nil)
}

// RemoveEmployee handles DELETE /areas/:id/employees/:emp_id
func (h *AreaHandler) RemoveEmployee(c *gin.Context) {
	areaID := c.Param("id")
	employeeID := c.Param("emp_id")

	area, err := h.areaUC.RemoveEmployee(c.Request.Context(), areaID, employeeID)
	if err != nil {
		if err == usecase.ErrAreaNotFound {
			errors.ErrorResponse(c, "AREA_NOT_FOUND", map[string]interface{}{
				"area_id": areaID,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, area, nil)
}

// GetFormData handles GET /areas/form-data and returns employee options for area forms
func (h *AreaHandler) GetFormData(c *gin.Context) {
	formData, err := h.areaUC.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, formData, nil)
}