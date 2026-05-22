package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
	"github.com/gilabs/gims/api/internal/report/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SupplierResearchHandler handles HTTP requests for supplier research reports.
type SupplierResearchHandler struct {
	uc usecase.SupplierResearchUsecase
}

func NewSupplierResearchHandler(uc usecase.SupplierResearchUsecase) *SupplierResearchHandler {
	return &SupplierResearchHandler{uc: uc}
}

func (h *SupplierResearchHandler) GetKpis(c *gin.Context) {
	var req dto.SupplierResearchKpisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetKpis(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Filters: map[string]interface{}{}}
	if req.StartDate != "" {
		meta.Filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		meta.Filters["end_date"] = req.EndDate
	}
	if req.CategoryIDs != "" {
		meta.Filters["category_ids"] = req.CategoryIDs
	}

	response.SuccessResponse(c, result, meta)
}

func (h *SupplierResearchHandler) ListPurchaseVolume(c *gin.Context) {
	var req dto.ListSupplierPurchaseVolumeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListPurchaseVolume(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildSupplierResearchMeta(pagination, req.Search, req.StartDate, req.EndDate, req.CategoryIDs)
	response.SuccessResponse(c, results, meta)
}

func (h *SupplierResearchHandler) ListDeliveryTime(c *gin.Context) {
	var req dto.ListSupplierDeliveryTimeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListDeliveryTime(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildSupplierResearchMeta(pagination, req.Search, req.StartDate, req.EndDate, req.CategoryIDs)
	response.SuccessResponse(c, results, meta)
}

func (h *SupplierResearchHandler) GetSpendTrend(c *gin.Context) {
	var req dto.SupplierSpendTrendRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetSpendTrend(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Filters: map[string]interface{}{"interval": req.Interval}}
	if req.StartDate != "" {
		meta.Filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		meta.Filters["end_date"] = req.EndDate
	}
	if req.CategoryIDs != "" {
		meta.Filters["category_ids"] = req.CategoryIDs
	}

	response.SuccessResponse(c, result, meta)
}

func (h *SupplierResearchHandler) ListSuppliers(c *gin.Context) {
	var req dto.ListSuppliersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListSuppliers(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildSupplierResearchMeta(pagination, req.Search, req.StartDate, req.EndDate, req.CategoryIDs)
	if meta.Filters == nil {
		meta.Filters = map[string]interface{}{}
	}
	meta.Filters["tab"] = req.Tab

	response.SuccessResponse(c, results, meta)
}

func (h *SupplierResearchHandler) GetSupplierDetail(c *gin.Context) {
	supplierID := c.Param("supplierId")
	if supplierID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "supplier_id", "message": "Supplier ID is required"}, nil)
		return
	}

	var req dto.SupplierResearchKpisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetSupplierDetail(c.Request.Context(), supplierID, req)
	if err != nil {
		if err.Error() == "supplier not found" {
			errors.NotFoundResponse(c, "supplier", supplierID)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

func buildSupplierResearchMeta(pagination utils.PaginationResult, search, startDate, endDate, categoryIDs string) *response.Meta {
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

	if search != "" {
		meta.Filters["search"] = search
	}
	if startDate != "" {
		meta.Filters["start_date"] = startDate
	}
	if endDate != "" {
		meta.Filters["end_date"] = endDate
	}
	if categoryIDs != "" {
		meta.Filters["category_ids"] = categoryIDs
	}

	return meta
}
