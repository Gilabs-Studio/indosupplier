package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ProvinceHandler handles province HTTP requests
type ProvinceHandler struct {
	provinceUC usecase.ProvinceUsecase
}

// NewProvinceHandler creates a new ProvinceHandler
func NewProvinceHandler(provinceUC usecase.ProvinceUsecase) *ProvinceHandler {
	return &ProvinceHandler{provinceUC: provinceUC}
}

// List handles list provinces request
func (h *ProvinceHandler) List(c *gin.Context) {
	var req dto.ListProvincesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	provinces, pagination, err := h.provinceUC.List(c.Request.Context(), &req)
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
	if req.CountryID != "" {
		meta.Filters["country_id"] = req.CountryID
	}

	response.SuccessResponse(c, provinces, meta)
}

// GetByID handles get province by ID request
func (h *ProvinceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	province, err := h.provinceUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrProvinceNotFound {
			errors.ErrorResponse(c, "PROVINCE_NOT_FOUND", map[string]interface{}{
				"province_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, province, nil)
}

// Create handles create province request
func (h *ProvinceHandler) Create(c *gin.Context) {
	var req dto.CreateProvinceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	province, err := h.provinceUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrCountryNotFound {
			errors.ErrorResponse(c, "COUNTRY_NOT_FOUND", map[string]interface{}{
				"country_id": req.CountryID,
			}, nil)
			return
		}
		if err == usecase.ErrProvinceAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "province",
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

	response.SuccessResponseCreated(c, province, meta)
}

// Update handles update province request
func (h *ProvinceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateProvinceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	province, err := h.provinceUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrProvinceNotFound {
			errors.ErrorResponse(c, "PROVINCE_NOT_FOUND", map[string]interface{}{
				"province_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCountryNotFound {
			errors.ErrorResponse(c, "COUNTRY_NOT_FOUND", map[string]interface{}{
				"country_id": req.CountryID,
			}, nil)
			return
		}
		if err == usecase.ErrProvinceAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "province",
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

	response.SuccessResponse(c, province, meta)
}

// Delete handles delete province request
func (h *ProvinceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.provinceUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrProvinceNotFound {
			errors.ErrorResponse(c, "PROVINCE_NOT_FOUND", map[string]interface{}{
				"province_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrProvinceHasCities {
			errors.ErrorResponse(c, "RESOURCE_IN_USE", map[string]interface{}{
				"resource": "province",
				"reason":   "has cities",
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

	response.SuccessResponseDeleted(c, "province", id, meta)
}
