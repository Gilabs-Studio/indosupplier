package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SalesPaymentHandler struct {
	uc usecase.SalesPaymentUsecase
}

func NewSalesPaymentHandler(uc usecase.SalesPaymentUsecase) *SalesPaymentHandler {
	return &SalesPaymentHandler{uc: uc}
}

// Add handles GET /sales/payments/add
func (h *SalesPaymentHandler) Add(c *gin.Context) {
	data, err := h.uc.AddData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, data, nil)
}

// List handles GET /sales/payments
func (h *SalesPaymentHandler) List(c *gin.Context) {
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

	search := c.Query("search")
	status := c.Query("status")
	method := c.Query("method")
	invoiceID := c.Query("invoice_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	params := repositories.SalesPaymentListParams{
		Search:    search,
		Status:    status,
		Method:    method,
		InvoiceID: invoiceID,
		StartDate: startDate,
		EndDate:   endDate,
		SortBy:    sortBy,
		SortDir:   sortDir,
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	}

	items, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
		Filters:    map[string]interface{}{},
		Sort:       &response.SortMeta{Field: params.SortBy, Order: params.SortDir},
	}
	if strings.TrimSpace(params.Search) != "" {
		meta.Filters["search"] = params.Search
	}
	if strings.TrimSpace(params.Status) != "" {
		meta.Filters["status"] = params.Status
	}
	if strings.TrimSpace(params.Method) != "" {
		meta.Filters["method"] = params.Method
	}
	if strings.TrimSpace(params.InvoiceID) != "" {
		meta.Filters["invoice_id"] = params.InvoiceID
	}
	if strings.TrimSpace(params.StartDate) != "" {
		meta.Filters["start_date"] = params.StartDate
	}
	if strings.TrimSpace(params.EndDate) != "" {
		meta.Filters["end_date"] = params.EndDate
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// GetByID handles GET /sales/payments/:id
func (h *SalesPaymentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesPaymentNotFound {
			errors.NotFoundResponse(c, "sales_payment", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Create handles POST /sales/payments
func (h *SalesPaymentHandler) Create(c *gin.Context) {
	var req dto.CreateSalesPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "customer invoice not found" {
			errors.NotFoundResponse(c, "customer_invoice", req.InvoiceID)
			return
		}
		if err == usecase.ErrSalesPaymentConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

// Delete handles DELETE /sales/payments/:id
func (h *SalesPaymentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if err == usecase.ErrSalesPaymentNotFound {
			errors.NotFoundResponse(c, "sales_payment", id)
			return
		}
		if err == usecase.ErrSalesPaymentConflict || err == usecase.ErrSalesPaymentDeletePendingOnly {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, map[string]interface{}{"id": id}, nil)
}

// Confirm handles POST /sales/payments/:id/confirm
func (h *SalesPaymentHandler) Confirm(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Confirm(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesPaymentNotFound {
			errors.NotFoundResponse(c, "sales_payment", id)
			return
		}
		if err == usecase.ErrSalesPaymentConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Reverse handles POST /sales/payments/:id/reverse
func (h *SalesPaymentHandler) Reverse(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Reverse(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesPaymentNotFound {
			errors.NotFoundResponse(c, "sales_payment", id)
			return
		}
		if err == usecase.ErrSalesPaymentConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// AuditTrail handles GET /sales/payments/:id/audit-trail
func (h *SalesPaymentHandler) AuditTrail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
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

	items, total, err := h.uc.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

// Export handles GET /sales/payments/export
func (h *SalesPaymentHandler) Export(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10000"))
	if limit < 1 {
		limit = 10000
	}
	if limit > 10000 {
		limit = 10000
	}

	params := repositories.SalesPaymentListParams{
		Search:    c.Query("search"),
		Status:    c.Query("status"),
		Method:    c.Query("method"),
		InvoiceID: c.Query("invoice_id"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortDir:   c.DefaultQuery("sort_dir", "desc"),
		Limit:     limit,
		Offset:    0,
	}

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		data, err := h.uc.ExportCSV(ctx, params)
		if err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    "sales_payments.csv",
			ContentType: "text/csv",
			Bytes:       data,
		}, nil
	}

	if exportjob.QueueIfRequested(c, generator) {
		return
	}

	file, err := generator(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	exportjob.WriteSyncFile(c, file)
}
