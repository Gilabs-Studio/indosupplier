package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// VillageHandler handles village HTTP requests
type VillageHandler struct {
	villageUC usecase.VillageUsecase
}

// NewVillageHandler creates a new VillageHandler
func NewVillageHandler(villageUC usecase.VillageUsecase) *VillageHandler {
	return &VillageHandler{villageUC: villageUC}
}

// List handles list villages request
func (h *VillageHandler) List(c *gin.Context) {
	var req dto.ListVillagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	villages, pagination, err := h.villageUC.List(c.Request.Context(), &req)
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
	if req.DistrictID != "" {
		meta.Filters["district_id"] = req.DistrictID
	}
	if req.Type != "" {
		meta.Filters["type"] = req.Type
	}

	response.SuccessResponse(c, villages, meta)
}

// GetByID handles get village by ID request
func (h *VillageHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	village, err := h.villageUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrVillageNotFound {
			errors.ErrorResponse(c, "VILLAGE_NOT_FOUND", map[string]interface{}{
				"village_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, village, nil)
}

// Create handles create village request
func (h *VillageHandler) Create(c *gin.Context) {
	var req dto.CreateVillageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	village, err := h.villageUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrDistrictNotFound {
			errors.ErrorResponse(c, "DISTRICT_NOT_FOUND", map[string]interface{}{
				"district_id": req.DistrictID,
			}, nil)
			return
		}
		if err == usecase.ErrVillageAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "village",
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

	response.SuccessResponseCreated(c, village, meta)
}

// Update handles update village request
func (h *VillageHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateVillageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	village, err := h.villageUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrVillageNotFound {
			errors.ErrorResponse(c, "VILLAGE_NOT_FOUND", map[string]interface{}{
				"village_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrDistrictNotFound {
			errors.ErrorResponse(c, "DISTRICT_NOT_FOUND", map[string]interface{}{
				"district_id": req.DistrictID,
			}, nil)
			return
		}
		if err == usecase.ErrVillageAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "village",
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

	response.SuccessResponse(c, village, meta)
}

// Delete handles delete village request
func (h *VillageHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.villageUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrVillageNotFound {
			errors.ErrorResponse(c, "VILLAGE_NOT_FOUND", map[string]interface{}{
				"village_id": id,
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

	response.SuccessResponseDeleted(c, "village", id, meta)
}
