package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// JobPositionHandler handles job position HTTP requests
type JobPositionHandler struct {
	jobPositionUC usecase.JobPositionUsecase
}

// NewJobPositionHandler creates a new JobPositionHandler
func NewJobPositionHandler(jobPositionUC usecase.JobPositionUsecase) *JobPositionHandler {
	return &JobPositionHandler{jobPositionUC: jobPositionUC}
}

func (h *JobPositionHandler) List(c *gin.Context) {
	var req dto.ListJobPositionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	jobPositions, pagination, err := h.jobPositionUC.List(c.Request.Context(), &req)
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

	response.SuccessResponse(c, jobPositions, meta)
}

func (h *JobPositionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	jobPosition, err := h.jobPositionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrJobPositionNotFound {
			errors.ErrorResponse(c, "JOB_POSITION_NOT_FOUND", map[string]interface{}{
				"job_position_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, jobPosition, nil)
}

func (h *JobPositionHandler) Create(c *gin.Context) {
	var req dto.CreateJobPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	jobPosition, err := h.jobPositionUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrJobPositionAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "job_position",
				"field":    "name",
				"value":    req.Name,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.CreatedBy = id
		}
	}

	response.SuccessResponseCreated(c, jobPosition, meta)
}

func (h *JobPositionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateJobPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	jobPosition, err := h.jobPositionUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrJobPositionNotFound {
			errors.ErrorResponse(c, "JOB_POSITION_NOT_FOUND", map[string]interface{}{
				"job_position_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrJobPositionAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "job_position",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}

	response.SuccessResponse(c, jobPosition, meta)
}

func (h *JobPositionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.jobPositionUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrJobPositionNotFound {
			errors.ErrorResponse(c, "JOB_POSITION_NOT_FOUND", map[string]interface{}{
				"job_position_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "job_position", id, meta)
}
