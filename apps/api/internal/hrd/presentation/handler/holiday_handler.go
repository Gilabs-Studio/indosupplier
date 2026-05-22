package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// HolidayHandler handles holiday HTTP requests
type HolidayHandler struct {
	holidayUC usecase.HolidayUsecase
}

// NewHolidayHandler creates a new HolidayHandler
func NewHolidayHandler(holidayUC usecase.HolidayUsecase) *HolidayHandler {
	return &HolidayHandler{holidayUC: holidayUC}
}

// List handles list holidays request
func (h *HolidayHandler) List(c *gin.Context) {
	var req dto.ListHolidaysRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	holidays, pagination, err := h.holidayUC.List(c.Request.Context(), &req)
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
	if req.Year > 0 {
		meta.Filters["year"] = req.Year
	}
	if req.Type != "" {
		meta.Filters["type"] = req.Type
	}

	response.SuccessResponse(c, holidays, meta)
}

// GetByID handles get holiday by ID request
func (h *HolidayHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	holiday, err := h.holidayUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrHolidayNotFound {
			errors.ErrorResponse(c, "HOLIDAY_NOT_FOUND", map[string]interface{}{
				"holiday_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, holiday, nil)
}

// GetByYear handles get holidays by year request
func (h *HolidayHandler) GetByYear(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		errors.ErrorResponse(c, "INVALID_YEAR", map[string]interface{}{
			"year": yearStr,
		}, nil)
		return
	}

	holidays, err := h.holidayUC.GetByYear(c.Request.Context(), year)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, holidays, nil)
}

// GetCalendar handles get holiday calendar request
func (h *HolidayHandler) GetCalendar(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		// Default to current year
		year = apptime.Now().Year()
	}

	calendar, err := h.holidayUC.GetCalendar(c.Request.Context(), year)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, calendar, nil)
}

// Create handles create holiday request
func (h *HolidayHandler) Create(c *gin.Context) {
	var req dto.CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	holiday, err := h.holidayUC.Create(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, holiday, nil)
}

// CreateBatch handles batch create holidays request
func (h *HolidayHandler) CreateBatch(c *gin.Context) {
	var req []dto.CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if len(req) == 0 {
		errors.ErrorResponse(c, "EMPTY_REQUEST", map[string]interface{}{
			"message": "No holidays provided",
		}, nil)
		return
	}

	holidays, err := h.holidayUC.CreateBatch(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, holidays, nil)
}

// Update handles update holiday request
func (h *HolidayHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	holiday, err := h.holidayUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrHolidayNotFound {
			errors.ErrorResponse(c, "HOLIDAY_NOT_FOUND", map[string]interface{}{
				"holiday_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, holiday, nil)
}

// Delete handles delete holiday request
func (h *HolidayHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.holidayUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrHolidayNotFound {
			errors.ErrorResponse(c, "HOLIDAY_NOT_FOUND", map[string]interface{}{
				"holiday_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"message": "Holiday deleted successfully",
	}, nil)
}

// CheckHoliday handles check if date is holiday request
func (h *HolidayHandler) CheckHoliday(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = apptime.Now().Format("2006-01-02")
	}

	isHoliday, holidayInfo, err := h.holidayUC.IsHoliday(c.Request.Context(), date)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"date":       date,
		"is_holiday": isHoliday,
		"holiday":    holidayInfo,
	}, nil)
}
