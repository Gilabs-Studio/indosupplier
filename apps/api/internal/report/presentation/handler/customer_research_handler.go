package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
	"github.com/gilabs/gims/api/internal/report/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CustomerResearchHandler handles HTTP requests for customer research reports.
type CustomerResearchHandler struct {
	uc usecase.CustomerResearchUsecase
}

// NewCustomerResearchHandler creates a new CustomerResearchHandler instance.
func NewCustomerResearchHandler(uc usecase.CustomerResearchUsecase) *CustomerResearchHandler {
	return &CustomerResearchHandler{uc: uc}
}

// GetKPIs returns KPI metrics for customer research.
func (h *CustomerResearchHandler) GetKPIs(c *gin.Context) {
	var req dto.GetCustomerResearchKpisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetKPIs(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetRevenueTrend returns customer revenue trend data.
func (h *CustomerResearchHandler) GetRevenueTrend(c *gin.Context) {
	var req dto.GetRevenueTrendRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetRevenueTrend(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// ListCustomers returns paginated customer research table data.
func (h *CustomerResearchHandler) ListCustomers(c *gin.Context) {
	var req dto.ListCustomersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, pagination, err := h.uc.ListCustomers(c.Request.Context(), req)
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

	response.SuccessResponse(c, result, meta)
}

// ListRevenueByCustomer returns paginated customer revenue ranking.
func (h *CustomerResearchHandler) ListRevenueByCustomer(c *gin.Context) {
	var req dto.ListRevenueByCustomerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, pagination, err := h.uc.ListRevenueByCustomer(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: &response.PaginationMeta{Page: pagination.Page, PerPage: pagination.PerPage, Total: pagination.Total, TotalPages: pagination.TotalPages, HasNext: pagination.Page < pagination.TotalPages, HasPrev: pagination.Page > 1}}
	response.SuccessResponse(c, result, meta)
}

// ListPurchaseFrequency returns paginated customer purchase frequency ranking.
func (h *CustomerResearchHandler) ListPurchaseFrequency(c *gin.Context) {
	var req dto.ListPurchaseFrequencyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, pagination, err := h.uc.ListPurchaseFrequency(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: &response.PaginationMeta{Page: pagination.Page, PerPage: pagination.PerPage, Total: pagination.Total, TotalPages: pagination.TotalPages, HasNext: pagination.Page < pagination.TotalPages, HasPrev: pagination.Page > 1}}
	response.SuccessResponse(c, result, meta)
}

// GetCustomerDetail returns detailed metrics for a specific customer.
func (h *CustomerResearchHandler) GetCustomerDetail(c *gin.Context) {
	customerID := c.Param("customerId")
	if customerID == "" {
		errors.InvalidQueryParamResponse(c)
		return
	}

	var req dto.GetCustomerResearchKpisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetCustomerDetail(c.Request.Context(), customerID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	if result == nil {
		errors.NotFoundResponse(c, "customer", customerID)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetCustomerTopProducts returns the top products sold to a specific customer.
func (h *CustomerResearchHandler) GetCustomerTopProducts(c *gin.Context) {
	customerID := c.Param("customerId")
	if customerID == "" {
		errors.InvalidQueryParamResponse(c)
		return
	}

	var req dto.GetCustomerTopProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetCustomerTopProducts(c.Request.Context(), customerID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}
