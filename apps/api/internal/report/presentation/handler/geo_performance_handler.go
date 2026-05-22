package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
	"github.com/gilabs/gims/api/internal/report/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GeoPerformanceHandler handles HTTP requests for geo performance reports
type GeoPerformanceHandler struct {
	uc usecase.GeoPerformanceUsecase
}

// NewGeoPerformanceHandler creates a new handler instance
func NewGeoPerformanceHandler(uc usecase.GeoPerformanceUsecase) *GeoPerformanceHandler {
	return &GeoPerformanceHandler{uc: uc}
}

// GetGeoPerformance returns geographic area performance data
func (h *GeoPerformanceHandler) GetGeoPerformance(c *gin.Context) {
	var req dto.GeoPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetGeoPerformance(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Filters: map[string]interface{}{
			"mode":  result.Mode,
			"level": result.Level,
		},
	}
	if req.StartDate != "" {
		meta.Filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		meta.Filters["end_date"] = req.EndDate
	}
	if req.SalesRepID != "" {
		meta.Filters["sales_rep_id"] = req.SalesRepID
	}

	response.SuccessResponse(c, result, meta)
}

// GetFormData returns filter options (sales reps) for the geo performance page
func (h *GeoPerformanceHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}
