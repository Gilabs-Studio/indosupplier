package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CityHandler handles city HTTP requests
type CityHandler struct {
	cityUC usecase.CityUsecase
}

// NewCityHandler creates a new CityHandler
func NewCityHandler(cityUC usecase.CityUsecase) *CityHandler {
	return &CityHandler{cityUC: cityUC}
}

// List handles list cities request
func (h *CityHandler) List(c *gin.Context) {
	var req dto.ListCitiesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	cities, pagination, err := h.cityUC.List(c.Request.Context(), &req)
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
	if req.ProvinceID != "" {
		meta.Filters["province_id"] = req.ProvinceID
	}
	if req.Type != "" {
		meta.Filters["type"] = req.Type
	}

	response.SuccessResponse(c, cities, meta)
}

// GetByID handles get city by ID request
func (h *CityHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	city, err := h.cityUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCityNotFound {
			errors.ErrorResponse(c, "CITY_NOT_FOUND", map[string]interface{}{
				"city_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, city, nil)
}

// Create handles create city request
func (h *CityHandler) Create(c *gin.Context) {
	var req dto.CreateCityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	city, err := h.cityUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrProvinceNotFound {
			errors.ErrorResponse(c, "PROVINCE_NOT_FOUND", map[string]interface{}{
				"province_id": req.ProvinceID,
			}, nil)
			return
		}
		if err == usecase.ErrCityAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "city",
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

	response.SuccessResponseCreated(c, city, meta)
}

// Update handles update city request
func (h *CityHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateCityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	city, err := h.cityUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrCityNotFound {
			errors.ErrorResponse(c, "CITY_NOT_FOUND", map[string]interface{}{
				"city_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrProvinceNotFound {
			errors.ErrorResponse(c, "PROVINCE_NOT_FOUND", map[string]interface{}{
				"province_id": req.ProvinceID,
			}, nil)
			return
		}
		if err == usecase.ErrCityAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "city",
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

	response.SuccessResponse(c, city, meta)
}

// Delete handles delete city request
func (h *CityHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.cityUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCityNotFound {
			errors.ErrorResponse(c, "CITY_NOT_FOUND", map[string]interface{}{
				"city_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCityHasDistricts {
			errors.ErrorResponse(c, "RESOURCE_IN_USE", map[string]interface{}{
				"resource": "city",
				"reason":   "has districts",
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

	response.SuccessResponseDeleted(c, "city", id, meta)
}
