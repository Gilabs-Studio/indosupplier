package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
	"github.com/gilabs/gims/api/internal/report/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SalesOverviewHandler handles HTTP requests for sales overview reports
type SalesOverviewHandler struct {
	uc usecase.SalesOverviewUsecase
}

func NewSalesOverviewHandler(uc usecase.SalesOverviewUsecase) *SalesOverviewHandler {
	return &SalesOverviewHandler{uc: uc}
}

// ListPerformance returns paginated sales rep performance list
func (h *SalesOverviewHandler) ListPerformance(c *gin.Context) {
	var req dto.ListSalesRepPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.ListSalesRepPerformance(c.Request.Context(), req)
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

// GetMonthlyOverview returns monthly sales overview data
func (h *SalesOverviewHandler) GetMonthlyOverview(c *gin.Context) {
	var req dto.MonthlySalesOverviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetMonthlySalesOverview(c.Request.Context(), req)
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

// GetSalesRepDetail returns detailed info for a specific sales rep
func (h *SalesOverviewHandler) GetSalesRepDetail(c *gin.Context) {
	employeeID := c.Param("employeeId")
	if employeeID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "employee_id", "message": "Employee ID is required"}, nil)
		return
	}

	var req dto.GetSalesRepDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetSalesRepDetail(c.Request.Context(), employeeID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	if result == nil {
		errors.NotFoundResponse(c, "sales_representative", employeeID)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetCheckInLocations returns check-in locations for a specific sales rep
func (h *SalesOverviewHandler) GetCheckInLocations(c *gin.Context) {
	employeeID := c.Param("employeeId")
	if employeeID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "employee_id", "message": "Employee ID is required"}, nil)
		return
	}

	var req dto.GetSalesRepCheckInLocationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetSalesRepCheckInLocations(c.Request.Context(), employeeID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetProducts returns product sales for a specific sales rep
func (h *SalesOverviewHandler) GetProducts(c *gin.Context) {
	employeeID := c.Param("employeeId")
	if employeeID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "employee_id", "message": "Employee ID is required"}, nil)
		return
	}

	var req dto.ListSalesRepProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.GetSalesRepProducts(c.Request.Context(), employeeID, req)
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

// GetCustomers returns customer data for a specific sales rep
func (h *SalesOverviewHandler) GetCustomers(c *gin.Context) {
	employeeID := c.Param("employeeId")
	if employeeID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"field": "employee_id", "message": "Employee ID is required"}, nil)
		return
	}

	var req dto.ListSalesRepCustomersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	results, pagination, err := h.uc.GetSalesRepCustomers(c.Request.Context(), employeeID, req)
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

// GetEmployeeDashboardMetrics returns aggregated metrics for current employee's dashboard
func (h *SalesOverviewHandler) GetEmployeeDashboardMetrics(c *gin.Context) {
	employeeID := c.GetString("employee_id")
	if employeeID == "" {
		if scopedEmployeeID, ok := c.Request.Context().Value("scope_employee_id").(string); ok {
			employeeID = scopedEmployeeID
		}
	}
	if employeeID == "" {
		if scopeCtx := middleware.GetScopeContext(c); scopeCtx != nil {
			employeeID = scopeCtx.EmployeeID
		}
	}

	// Some authenticated users may not be linked to an employee record yet.
	// Return empty metrics payload instead of a validation error to keep profile page usable.
	if employeeID == "" {
		response.SuccessResponse(c, &dto.EmployeeDashboardMetricsResponse{}, nil)
		return
	}

	var req dto.EmployeeDashboardMetricsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	result, err := h.uc.GetEmployeeDashboardMetrics(c.Request.Context(), employeeID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}
