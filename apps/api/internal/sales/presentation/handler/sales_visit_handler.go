package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SalesVisitHandler handles sales visit HTTP requests
type SalesVisitHandler struct {
	visitUC usecase.SalesVisitUsecase
}

// NewSalesVisitHandler creates a new SalesVisitHandler
func NewSalesVisitHandler(visitUC usecase.SalesVisitUsecase) *SalesVisitHandler {
	return &SalesVisitHandler{visitUC: visitUC}
}

// List handles list sales visits request
func (h *SalesVisitHandler) List(c *gin.Context) {
	var req dto.ListSalesVisitsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	visits, pagination, err := h.visitUC.List(c.Request.Context(), &req)
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
	if req.Status != "" {
		meta.Filters["status"] = req.Status
	}
	if req.EmployeeID != "" {
		meta.Filters["employee_id"] = req.EmployeeID
	}
	if req.CompanyID != "" {
		meta.Filters["company_id"] = req.CompanyID
	}

	response.SuccessResponse(c, visits, meta)
}

// GetByID handles get sales visit by ID request
func (h *SalesVisitHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	visit, err := h.visitUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesVisitNotFound {
			errors.ErrorResponse(c, "SALES_VISIT_NOT_FOUND", map[string]interface{}{
				"visit_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, visit, nil)
}

// ListDetails handles list visit details request
func (h *SalesVisitHandler) ListDetails(c *gin.Context) {
	visitID := c.Param("id")

	var req dto.ListSalesVisitDetailsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	details, pagination, err := h.visitUC.ListDetails(c.Request.Context(), visitID, &req)
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

	response.SuccessResponse(c, details, meta)
}

// ListProgressHistory handles list progress history request
func (h *SalesVisitHandler) ListProgressHistory(c *gin.Context) {
	visitID := c.Param("id")

	var req dto.ListSalesVisitProgressHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	history, pagination, err := h.visitUC.ListProgressHistory(c.Request.Context(), visitID, &req)
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

	response.SuccessResponse(c, history, meta)
}

// Create handles create sales visit request
func (h *SalesVisitHandler) Create(c *gin.Context) {
	var req dto.CreateSalesVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var createdBy *string
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = &id
		}
	}

	visit, err := h.visitUC.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if createdBy != nil {
		meta.CreatedBy = *createdBy
	}

	response.SuccessResponseCreated(c, visit, meta)
}

// Update handles update sales visit request
func (h *SalesVisitHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalesVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	visit, err := h.visitUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrSalesVisitNotFound {
			errors.ErrorResponse(c, "SALES_VISIT_NOT_FOUND", map[string]interface{}{
				"visit_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCannotModifyCompletedVisit {
			errors.ErrorResponse(c, "CANNOT_MODIFY_VISIT", map[string]interface{}{
				"message": "Cannot modify completed or cancelled visit",
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

	response.SuccessResponse(c, visit, meta)
}

// Delete handles delete sales visit request
func (h *SalesVisitHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.visitUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesVisitNotFound {
			errors.ErrorResponse(c, "SALES_VISIT_NOT_FOUND", map[string]interface{}{
				"visit_id": id,
			}, nil)
			return
		}
		errors.ErrorResponse(c, "CANNOT_DELETE_VISIT", map[string]interface{}{
			"message": err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "sales_visit", id, meta)
}

// UpdateStatus handles update sales visit status request
func (h *SalesVisitHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalesVisitStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	visit, err := h.visitUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrSalesVisitNotFound {
			errors.ErrorResponse(c, "SALES_VISIT_NOT_FOUND", map[string]interface{}{
				"visit_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidVisitTransition {
			errors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Invalid status transition",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID != nil {
		meta.UpdatedBy = *userID
	}

	response.SuccessResponse(c, visit, meta)
}

// CheckIn handles check-in request
func (h *SalesVisitHandler) CheckIn(c *gin.Context) {
	id := c.Param("id")
	var req dto.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	visit, err := h.visitUC.CheckIn(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrSalesVisitNotFound {
			errors.ErrorResponse(c, "SALES_VISIT_NOT_FOUND", map[string]interface{}{
				"visit_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrVisitAlreadyCheckedIn {
			errors.ErrorResponse(c, "ALREADY_CHECKED_IN", map[string]interface{}{
				"message": "Visit already checked in",
			}, nil)
			return
		}
		if err == usecase.ErrVisitAlreadyCompleted {
			errors.ErrorResponse(c, "VISIT_COMPLETED", map[string]interface{}{
				"message": "Visit already completed",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID != nil {
		meta.UpdatedBy = *userID
	}

	response.SuccessResponse(c, visit, meta)
}

// CheckOut handles check-out request
func (h *SalesVisitHandler) CheckOut(c *gin.Context) {
	id := c.Param("id")
	var req dto.CheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	visit, err := h.visitUC.CheckOut(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrSalesVisitNotFound {
			errors.ErrorResponse(c, "SALES_VISIT_NOT_FOUND", map[string]interface{}{
				"visit_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrVisitNotCheckedIn {
			errors.ErrorResponse(c, "NOT_CHECKED_IN", map[string]interface{}{
				"message": "Visit must be checked in first",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID != nil {
		meta.UpdatedBy = *userID
	}

	response.SuccessResponse(c, visit, meta)
}

// GetCalendarSummary handles get calendar summary request
func (h *SalesVisitHandler) GetCalendarSummary(c *gin.Context) {
	var req dto.GetCalendarSummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	summary, err := h.visitUC.GetCalendarSummary(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, summary, nil)
}
