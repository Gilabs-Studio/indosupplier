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

// ProductAnalysisHandler handles HTTP requests for product analysis reports
type ProductAnalysisHandler struct {
	uc usecase.ProductAnalysisUsecase
}

// NewProductAnalysisHandler creates a new ProductAnalysisHandler instance
func NewProductAnalysisHandler(uc usecase.ProductAnalysisUsecase) *ProductAnalysisHandler {
	return &ProductAnalysisHandler{uc: uc}
}

// ListPerformance returns paginated product performance list
func (h *ProductAnalysisHandler) ListPerformance(c *gin.Context) {
	var req dto.ListProductPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListProductPerformance(c.Request.Context(), req)
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
	if req.CategoryID != "" {
		meta.Filters["category_id"] = req.CategoryID
	}
	if req.StartDate != "" {
		meta.Filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		meta.Filters["end_date"] = req.EndDate
	}

	response.SuccessResponse(c, results, meta)
}

// GetMonthlyOverview returns monthly product sales overview data
func (h *ProductAnalysisHandler) GetMonthlyOverview(c *gin.Context) {
	var req dto.MonthlyProductSalesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetMonthlyProductSales(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Filters: map[string]interface{}{},
	}
	if req.StartDate != "" {
		meta.Filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		meta.Filters["end_date"] = req.EndDate
	}

	response.SuccessResponse(c, result, meta)
}

// GetProductDetail returns detailed info for a specific product
func (h *ProductAnalysisHandler) GetProductDetail(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "product_id", "message": "Product ID is required"}, nil)
		return
	}

	var req dto.GetProductDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetProductDetail(c.Request.Context(), productID, req)
	if err != nil {
		if err.Error() == "product not found" {
			errors.NotFoundResponse(c, "product", productID)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetProductCustomers returns customers buying a specific product
func (h *ProductAnalysisHandler) GetProductCustomers(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "product_id", "message": "Product ID is required"}, nil)
		return
	}

	var req dto.ListProductCustomersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.GetProductCustomers(c.Request.Context(), productID, req)
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
	}

	response.SuccessResponse(c, results, meta)
}

// GetProductSalesReps returns sales reps selling a specific product
func (h *ProductAnalysisHandler) GetProductSalesReps(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "product_id", "message": "Product ID is required"}, nil)
		return
	}

	var req dto.ListProductSalesRepsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.GetProductSalesReps(c.Request.Context(), productID, req)
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
	}

	response.SuccessResponse(c, results, meta)
}

// GetProductMonthlyTrend returns monthly trend for a specific product
func (h *ProductAnalysisHandler) GetProductMonthlyTrend(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "product_id", "message": "Product ID is required"}, nil)
		return
	}

	var req dto.GetProductMonthlyTrendRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetProductMonthlyTrend(c.Request.Context(), productID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// ListCategoryPerformance returns paginated category performance aggregated from product sales
func (h *ProductAnalysisHandler) ListCategoryPerformance(c *gin.Context) {
	var req dto.ListCategoryPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListCategoryPerformance(c.Request.Context(), req)
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
	if req.StartDate != "" {
		meta.Filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		meta.Filters["end_date"] = req.EndDate
	}

	response.SuccessResponse(c, results, meta)
}

// ListSegmentPerformance returns paginated segment performance aggregated from product sales
func (h *ProductAnalysisHandler) ListSegmentPerformance(c *gin.Context) {
	var req dto.ListSegmentPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListSegmentPerformance(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildDimensionMeta(pagination, req.Search, req.StartDate, req.EndDate)
	response.SuccessResponse(c, results, meta)
}

// ListTypePerformance returns paginated product type performance aggregated from product sales
func (h *ProductAnalysisHandler) ListTypePerformance(c *gin.Context) {
	var req dto.ListTypePerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListTypePerformance(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildDimensionMeta(pagination, req.Search, req.StartDate, req.EndDate)
	response.SuccessResponse(c, results, meta)
}

// ListPackagingPerformance returns paginated packaging performance aggregated from product sales
func (h *ProductAnalysisHandler) ListPackagingPerformance(c *gin.Context) {
	var req dto.ListPackagingPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListPackagingPerformance(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildDimensionMeta(pagination, req.Search, req.StartDate, req.EndDate)
	response.SuccessResponse(c, results, meta)
}

// ListProcurementTypePerformance returns paginated procurement type performance aggregated from product sales
func (h *ProductAnalysisHandler) ListProcurementTypePerformance(c *gin.Context) {
	var req dto.ListProcurementTypePerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListProcurementTypePerformance(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildDimensionMeta(pagination, req.Search, req.StartDate, req.EndDate)
	response.SuccessResponse(c, results, meta)
}

// buildDimensionMeta constructs pagination meta for dimension performance endpoints
func buildDimensionMeta(pagination utils.PaginationResult, search, startDate, endDate string) *response.Meta {
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
	return meta
}
