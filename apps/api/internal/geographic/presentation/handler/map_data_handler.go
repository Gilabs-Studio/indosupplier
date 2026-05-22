package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// MapDataHandler handles geographic map data HTTP requests
type MapDataHandler struct {
	mapDataUC usecase.MapDataUsecase
}

// NewMapDataHandler creates a new MapDataHandler
func NewMapDataHandler(mapDataUC usecase.MapDataUsecase) *MapDataHandler {
	return &MapDataHandler{mapDataUC: mapDataUC}
}

// GetMapData handles GET /geographic/map-data
func (h *MapDataHandler) GetMapData(c *gin.Context) {
	var req dto.MapDataRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	featureCollection, err := h.mapDataUC.GetMapData(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrMapDataProvinceIDRequired:
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
				"message": "province_id is required when level is 'city'",
			}, nil)
		case usecase.ErrMapDataCityIDRequired:
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
				"message": "city_id is required when level is 'district'",
			}, nil)
		case usecase.ErrMapDataInvalidLevel:
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
				"message": "level must be one of: province, city, district",
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, featureCollection, nil)
}

// ReverseGeocode handles GET /geographic/reverse-geocode
// Resolves lat/lng coordinates to Province, City, and District
func (h *MapDataHandler) ReverseGeocode(c *gin.Context) {
	var req dto.ReverseGeocodeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.mapDataUC.ReverseGeocode(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrReverseGeocodeNotFound:
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"message": "No administrative boundary found for the given coordinates",
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, result, nil)
}
