package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AreaCaptureHandler handles area capture / mapping HTTP requests
type AreaCaptureHandler struct {
	captureUC usecase.AreaCaptureUsecase
}

// NewAreaCaptureHandler creates a new AreaCaptureHandler
func NewAreaCaptureHandler(captureUC usecase.AreaCaptureUsecase) *AreaCaptureHandler {
	return &AreaCaptureHandler{captureUC: captureUC}
}

// Capture handles POST /crm/area-mapping/capture — capture GPS location
func (h *AreaCaptureHandler) Capture(c *gin.Context) {
	var req dto.CreateAreaCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	capture, err := h.captureUC.Capture(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, capture, nil)
}

// ListCaptures handles GET /crm/area-mapping/captures — list captured GPS points
func (h *AreaCaptureHandler) ListCaptures(c *gin.Context) {
	var req dto.ListAreaCapturesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	captures, pagination, err := h.captureUC.List(c.Request.Context(), &req)
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

	response.SuccessResponse(c, captures, meta)
}

// GetHeatmap handles GET /crm/area-mapping/heatmap — visit heatmap data
func (h *AreaCaptureHandler) GetHeatmap(c *gin.Context) {
	areaID := c.Query("area_id")

	heatmap, err := h.captureUC.GetHeatmap(c.Request.Context(), areaID)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, heatmap, nil)
}

// GetCoverage handles GET /crm/area-mapping/coverage — coverage analysis per area
func (h *AreaCaptureHandler) GetCoverage(c *gin.Context) {
	coverage, err := h.captureUC.GetCoverage(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, coverage, nil)
}
