package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SalesQuotationHandler handles sales quotation HTTP requests
type SalesQuotationHandler struct {
	quotationUC usecase.SalesQuotationUsecase
}

// NewSalesQuotationHandler creates a new SalesQuotationHandler
func NewSalesQuotationHandler(quotationUC usecase.SalesQuotationUsecase) *SalesQuotationHandler {
	return &SalesQuotationHandler{quotationUC: quotationUC}
}

// List handles list sales quotations request
func (h *SalesQuotationHandler) List(c *gin.Context) {
	var req dto.ListSalesQuotationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	quotations, pagination, err := h.quotationUC.List(c.Request.Context(), &req)
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
	if req.SalesRepID != "" {
		meta.Filters["sales_rep_id"] = req.SalesRepID
	}
	if req.BusinessUnitID != "" {
		meta.Filters["business_unit_id"] = req.BusinessUnitID
	}

	response.SuccessResponse(c, quotations, meta)
}

// GetByID handles get sales quotation by ID request
func (h *SalesQuotationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	quotation, err := h.quotationUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesQuotationNotFound {
			errors.ErrorResponse(c, "SALES_QUOTATION_NOT_FOUND", map[string]interface{}{
				"quotation_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, quotation, nil)
}

// Create handles create sales quotation request
func (h *SalesQuotationHandler) Create(c *gin.Context) {
	var req dto.CreateSalesQuotationRequest
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

	quotation, err := h.quotationUC.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		if err == usecase.ErrProductNotFound {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"message": "One or more products not found",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if createdBy != nil {
		meta.CreatedBy = *createdBy
	}

	response.SuccessResponseCreated(c, quotation, meta)
}

// Update handles update sales quotation request
func (h *SalesQuotationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalesQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	quotation, err := h.quotationUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrSalesQuotationNotFound {
			errors.ErrorResponse(c, "SALES_QUOTATION_NOT_FOUND", map[string]interface{}{
				"quotation_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidQuotationStatus {
			errors.ErrorResponse(c, "INVALID_QUOTATION_STATUS", map[string]interface{}{
				"message": "Cannot modify quotation in current status",
			}, nil)
			return
		}
		if err == usecase.ErrProductNotFound {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"message": "One or more products not found",
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

	response.SuccessResponse(c, quotation, meta)
}

// Delete handles delete sales quotation request
func (h *SalesQuotationHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.quotationUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesQuotationNotFound {
			errors.ErrorResponse(c, "SALES_QUOTATION_NOT_FOUND", map[string]interface{}{
				"quotation_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidQuotationStatus {
			errors.ErrorResponse(c, "INVALID_QUOTATION_STATUS", map[string]interface{}{
				"message": "Cannot delete quotation in current status",
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

	response.SuccessResponseDeleted(c, "sales_quotation", id, meta)
}

// UpdateStatus handles update sales quotation status request
func (h *SalesQuotationHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalesQuotationStatusRequest
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

	quotation, err := h.quotationUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrSalesQuotationNotFound {
			errors.ErrorResponse(c, "SALES_QUOTATION_NOT_FOUND", map[string]interface{}{
				"quotation_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidStatusTransition {
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

	response.SuccessResponse(c, quotation, meta)
}

// ListItems handles list sales quotation items request with pagination
func (h *SalesQuotationHandler) ListItems(c *gin.Context) {
	quotationID := c.Param("id")

	var req dto.ListSalesQuotationItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	items, pagination, err := h.quotationUC.ListItems(c.Request.Context(), quotationID, &req)
	if err != nil {
		if err == usecase.ErrSalesQuotationNotFound {
			errors.ErrorResponse(c, "SALES_QUOTATION_NOT_FOUND", map[string]interface{}{
				"quotation_id": quotationID,
			}, nil)
			return
		}
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

	response.SuccessResponse(c, items, meta)
}

// AuditTrail handles list sales quotation audit trail with pagination.
func (h *SalesQuotationHandler) AuditTrail(c *gin.Context) {
	id := c.Param("id")

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

	entries, total, err := h.quotationUC.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, entries, meta)
}
