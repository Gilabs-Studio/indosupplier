package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// DistrictHandler handles district HTTP requests
type DistrictHandler struct {
	districtUC usecase.DistrictUsecase
}

// NewDistrictHandler creates a new DistrictHandler
func NewDistrictHandler(districtUC usecase.DistrictUsecase) *DistrictHandler {
	return &DistrictHandler{districtUC: districtUC}
}

// List handles list districts request
func (h *DistrictHandler) List(c *gin.Context) {
	var req dto.ListDistrictsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	districts, pagination, err := h.districtUC.List(c.Request.Context(), &req)
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
	if req.CityID != "" {
		meta.Filters["city_id"] = req.CityID
	}

	response.SuccessResponse(c, districts, meta)
}

// GetByID handles get district by ID request
func (h *DistrictHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	district, err := h.districtUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrDistrictNotFound {
			errors.ErrorResponse(c, "DISTRICT_NOT_FOUND", map[string]interface{}{
				"district_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, district, nil)
}

// Create handles create district request
func (h *DistrictHandler) Create(c *gin.Context) {
	var req dto.CreateDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	district, err := h.districtUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrCityNotFound {
			errors.ErrorResponse(c, "CITY_NOT_FOUND", map[string]interface{}{
				"city_id": req.CityID,
			}, nil)
			return
		}
		if err == usecase.ErrDistrictAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "district",
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

	response.SuccessResponseCreated(c, district, meta)
}

// Update handles update district request
func (h *DistrictHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	district, err := h.districtUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrDistrictNotFound {
			errors.ErrorResponse(c, "DISTRICT_NOT_FOUND", map[string]interface{}{
				"district_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCityNotFound {
			errors.ErrorResponse(c, "CITY_NOT_FOUND", map[string]interface{}{
				"city_id": req.CityID,
			}, nil)
			return
		}
		if err == usecase.ErrDistrictAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "district",
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

	response.SuccessResponse(c, district, meta)
}

// Delete handles delete district request
func (h *DistrictHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.districtUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrDistrictNotFound {
			errors.ErrorResponse(c, "DISTRICT_NOT_FOUND", map[string]interface{}{
				"district_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrDistrictHasVillages {
			errors.ErrorResponse(c, "RESOURCE_IN_USE", map[string]interface{}{
				"resource": "district",
				"reason":   "has villages",
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

	response.SuccessResponseDeleted(c, "district", id, meta)
}
