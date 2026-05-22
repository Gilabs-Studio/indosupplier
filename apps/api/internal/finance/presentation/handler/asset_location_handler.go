package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type AssetLocationHandler struct {
	uc usecase.AssetLocationUsecase
}

func NewAssetLocationHandler(uc usecase.AssetLocationUsecase) *AssetLocationHandler {
	return &AssetLocationHandler{uc: uc}
}

func (h *AssetLocationHandler) Create(c *gin.Context) {
	var req dto.CreateAssetLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_LOCATION_CREATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

func (h *AssetLocationHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.UpdateAssetLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_LOCATION_UPDATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetLocationHandler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_LOCATION_DELETE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseDeleted(c, "asset_location", id, nil)
}

func (h *AssetLocationHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "ASSET_LOCATION_NOT_FOUND", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetLocationHandler) List(c *gin.Context) {
	var req dto.ListAssetLocationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	req.Page = page
	req.PerPage = perPage

	items, total, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_LOCATION_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}
