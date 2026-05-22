package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// WorkScheduleHandler handles work schedule HTTP requests
type WorkScheduleHandler struct {
	workScheduleUC usecase.WorkScheduleUsecase
}

// NewWorkScheduleHandler creates a new WorkScheduleHandler
func NewWorkScheduleHandler(workScheduleUC usecase.WorkScheduleUsecase) *WorkScheduleHandler {
	return &WorkScheduleHandler{workScheduleUC: workScheduleUC}
}

// List handles list work schedules request
func (h *WorkScheduleHandler) List(c *gin.Context) {
	var req dto.ListWorkSchedulesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	schedules, pagination, err := h.workScheduleUC.List(c.Request.Context(), &req)
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
		Filters: map[string]interface{}{},
	}

	if req.Search != "" {
		meta.Filters["search"] = req.Search
	}
	if req.DivisionID != "" {
		meta.Filters["division_id"] = req.DivisionID
	}
	if req.IsActive != nil {
		meta.Filters["is_active"] = *req.IsActive
	}

	response.SuccessResponse(c, schedules, meta)
}

// GetByID handles get work schedule by ID request
func (h *WorkScheduleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	schedule, err := h.workScheduleUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrWorkScheduleNotFound {
			errors.ErrorResponse(c, "WORK_SCHEDULE_NOT_FOUND", map[string]interface{}{
				"schedule_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, schedule, nil)
}

// GetDefault handles get default work schedule request
func (h *WorkScheduleHandler) GetDefault(c *gin.Context) {
	schedule, err := h.workScheduleUC.GetDefault(c.Request.Context())
	if err != nil {
		if err == usecase.ErrWorkScheduleNotFound {
			errors.ErrorResponse(c, "WORK_SCHEDULE_NOT_FOUND", map[string]interface{}{
				"message": "No default work schedule configured",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, schedule, nil)
}

// Create handles create work schedule request
func (h *WorkScheduleHandler) Create(c *gin.Context) {
	var req dto.CreateWorkScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	schedule, err := h.workScheduleUC.Create(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, schedule, nil)
}

// Update handles update work schedule request
func (h *WorkScheduleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateWorkScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	schedule, err := h.workScheduleUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrWorkScheduleNotFound {
			errors.ErrorResponse(c, "WORK_SCHEDULE_NOT_FOUND", map[string]interface{}{
				"schedule_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, schedule, nil)
}

// Delete handles delete work schedule request
func (h *WorkScheduleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.workScheduleUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrWorkScheduleNotFound {
			errors.ErrorResponse(c, "WORK_SCHEDULE_NOT_FOUND", map[string]interface{}{
				"schedule_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCannotDeleteDefaultSchedule {
			errors.ErrorResponse(c, "CANNOT_DELETE_DEFAULT_SCHEDULE", map[string]interface{}{
				"schedule_id": id,
				"message":     "Cannot delete the default work schedule",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"message": "Work schedule deleted successfully",
	}, nil)
}

// SetDefault handles set default work schedule request
func (h *WorkScheduleHandler) SetDefault(c *gin.Context) {
	id := c.Param("id")

	err := h.workScheduleUC.SetDefault(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrWorkScheduleNotFound {
			errors.ErrorResponse(c, "WORK_SCHEDULE_NOT_FOUND", map[string]interface{}{
				"schedule_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCannotSetDivisionScheduleAsDefault {
			errors.ErrorResponse(c, "CANNOT_SET_DIVISION_SCHEDULE_AS_DEFAULT", map[string]interface{}{
				"schedule_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"message": "Default work schedule set successfully",
	}, nil)
}

// GetFormData handles GET /work-schedules/form-data
func (h *WorkScheduleHandler) GetFormData(c *gin.Context) {
	formData, err := h.workScheduleUC.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, formData, nil)
}
