package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CountryHandler handles country HTTP requests
type CountryHandler struct {
	countryUC usecase.CountryUsecase
}

// NewCountryHandler creates a new CountryHandler
func NewCountryHandler(countryUC usecase.CountryUsecase) *CountryHandler {
	return &CountryHandler{countryUC: countryUC}
}

// List handles list countries request
func (h *CountryHandler) List(c *gin.Context) {
	var req dto.ListCountriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	countries, pagination, err := h.countryUC.List(c.Request.Context(), &req)
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

	response.SuccessResponse(c, countries, meta)
}

// GetByID handles get country by ID request
func (h *CountryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	country, err := h.countryUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCountryNotFound {
			errors.ErrorResponse(c, "COUNTRY_NOT_FOUND", map[string]interface{}{
				"country_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, country, nil)
}

// Create handles create country request
func (h *CountryHandler) Create(c *gin.Context) {
	var req dto.CreateCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	country, err := h.countryUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrCountryAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "country",
				"field":    "code",
				"value":    req.Code,
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

	response.SuccessResponseCreated(c, country, meta)
}

// Update handles update country request
func (h *CountryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	country, err := h.countryUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrCountryNotFound {
			errors.ErrorResponse(c, "COUNTRY_NOT_FOUND", map[string]interface{}{
				"country_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCountryAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "country",
				"field":    "code",
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

	response.SuccessResponse(c, country, meta)
}

// Delete handles delete country request
func (h *CountryHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.countryUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCountryNotFound {
			errors.ErrorResponse(c, "COUNTRY_NOT_FOUND", map[string]interface{}{
				"country_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCountryHasProvinces {
			errors.ErrorResponse(c, "RESOURCE_IN_USE", map[string]interface{}{
				"resource": "country",
				"reason":   "has provinces",
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

	response.SuccessResponseDeleted(c, "country", id, meta)
}
