package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/core/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SOSourceHandler struct {
	uc usecase.SOSourceUsecase
}

func NewSOSourceHandler(uc usecase.SOSourceUsecase) *SOSourceHandler {
	return &SOSourceHandler{uc: uc}
}

func (h *SOSourceHandler) Create(c *gin.Context) {
	var req dto.CreateSOSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

func (h *SOSourceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "SO_SOURCE_NOT_FOUND", map[string]interface{}{"id": id}, nil)
		return
	}
	response.SuccessResponse(c, result, nil)
}

func (h *SOSourceHandler) List(c *gin.Context) {
	params := repositories.ListParams{
		Search:  c.Query("search"),
		SortBy:  c.DefaultQuery("sort_by", "name"),
		SortDir: c.DefaultQuery("sort_dir", "asc"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page: page, PerPage: perPage, Total: int(total), TotalPages: totalPages,
			HasNext: page < totalPages, HasPrev: page > 1,
		},
	}

	response.SuccessResponse(c, results, meta)
}

func (h *SOSourceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSOSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Update(c.Request.Context(), id, req)
	if err != nil {
		errors.ErrorResponse(c, "SO_SOURCE_NOT_FOUND", map[string]interface{}{"id": id}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

func (h *SOSourceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "SO_SOURCE_NOT_FOUND", map[string]interface{}{"id": id}, nil)
		return
	}
	response.SuccessResponseDeleted(c, "so_source", id, nil)
}
