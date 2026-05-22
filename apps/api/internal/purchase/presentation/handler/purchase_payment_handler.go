package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/purchase/data/repositories"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
	"github.com/gilabs/gims/api/internal/purchase/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PurchasePaymentHandler struct {
	uc usecase.PurchasePaymentUsecase
}

func NewPurchasePaymentHandler(uc usecase.PurchasePaymentUsecase) *PurchasePaymentHandler {
	return &PurchasePaymentHandler{uc: uc}
}

// Add handles GET /purchase/payments/add
func (h *PurchasePaymentHandler) Add(c *gin.Context) {
	data, err := h.uc.AddData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, data, nil)
}

// List handles GET /purchase/payments
func (h *PurchasePaymentHandler) List(c *gin.Context) {
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

	params := repositories.PurchasePaymentListParams{
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

// GetByID handles GET /purchase/payments/:id
func (h *PurchasePaymentHandler) GetByID(c *gin.Context) {
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
		if err == usecase.ErrPurchasePaymentNotFound {
			errors.NotFoundResponse(c, "purchase_payment", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Create handles POST /purchase/payments
func (h *PurchasePaymentHandler) Create(c *gin.Context) {
	var req dto.CreatePurchasePaymentRequest
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
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice", req.InvoiceID)
			return
		}
		if err == usecase.ErrPurchasePaymentConflict {
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

// CreateBatch handles POST /purchase/payments/batch
func (h *PurchasePaymentHandler) CreateBatch(c *gin.Context) {
	var req dto.CreatePurchasePaymentBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	res, err := h.uc.CreateBatch(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice", "batch_item")
			return
		}
		if err == usecase.ErrPurchasePaymentConflict {
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

// Update handles PUT /purchase/payments/:id
func (h *PurchasePaymentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}

	var req dto.UpdatePurchasePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrPurchasePaymentNotFound {
			errors.NotFoundResponse(c, "purchase_payment", id)
			return
		}
		if err == usecase.ErrPurchasePaymentConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		if err.Error() == "bank account not found" {
			errors.NotFoundResponse(c, "bank_account", req.BankAccountID)
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

// Delete handles DELETE /purchase/payments/:id
func (h *PurchasePaymentHandler) Delete(c *gin.Context) {
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
		if err == usecase.ErrPurchasePaymentNotFound {
			errors.NotFoundResponse(c, "purchase_payment", id)
			return
		}
		if err == usecase.ErrPurchasePaymentConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, map[string]interface{}{"id": id}, nil)
}

// Confirm handles POST /purchase/payments/:id/confirm
func (h *PurchasePaymentHandler) Confirm(c *gin.Context) {
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
		if err == usecase.ErrPurchasePaymentNotFound {
			errors.NotFoundResponse(c, "purchase_payment", id)
			return
		}
		if err == usecase.ErrPurchasePaymentConflict {
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

// ConfirmBatch handles PATCH /purchase/payments/batch/confirm and /batch/post
func (h *PurchasePaymentHandler) ConfirmBatch(c *gin.Context) {
	var req dto.ConfirmPurchasePaymentBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	res, err := h.uc.ConfirmBatch(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrPurchasePaymentNotFound {
			errors.NotFoundResponse(c, "purchase_payment", "batch_item")
			return
		}
		if err == usecase.ErrPurchasePaymentConflict {
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

// AuditTrail handles GET /purchase/payments/:id/audit-trail
func (h *PurchasePaymentHandler) AuditTrail(c *gin.Context) {
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

// Export handles GET /purchase/payments/export
func (h *PurchasePaymentHandler) Export(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10000"))
	if limit < 1 {
		limit = 10000
	}
	if limit > 10000 {
		limit = 10000
	}

	params := repositories.PurchasePaymentListParams{
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
			FileName:    "purchase_payments.csv",
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
