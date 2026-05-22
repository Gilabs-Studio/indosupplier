package handler

import (
	"context"
	"fmt"
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

type SupplierInvoiceDownPaymentHandler struct {
	uc usecase.SupplierInvoiceDownPaymentUsecase
}

func NewSupplierInvoiceDownPaymentHandler(uc usecase.SupplierInvoiceDownPaymentUsecase) *SupplierInvoiceDownPaymentHandler {
	return &SupplierInvoiceDownPaymentHandler{uc: uc}
}

// Add handles GET /purchase/supplier-invoice-down-payments/add
func (h *SupplierInvoiceDownPaymentHandler) Add(c *gin.Context) {
	data, err := h.uc.AddData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, data, nil)
}

// List handles GET /purchase/supplier-invoice-down-payments
func (h *SupplierInvoiceDownPaymentHandler) List(c *gin.Context) {
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
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	params := repositories.SupplierInvoiceListParams{
		Search:  search,
		Status:  status,
		SortBy:  sortBy,
		SortDir: sortDir,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
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
		Sort: &response.SortMeta{
			Field: params.SortBy,
			Order: params.SortDir,
		},
	}
	if strings.TrimSpace(params.Search) != "" {
		meta.Filters["search"] = params.Search
	}
	if strings.TrimSpace(params.Status) != "" {
		meta.Filters["status"] = params.Status
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// GetByID handles GET /purchase/supplier-invoice-down-payments/:id
func (h *SupplierInvoiceDownPaymentHandler) GetByID(c *gin.Context) {
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
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Create handles POST /purchase/supplier-invoice-down-payments
func (h *SupplierInvoiceDownPaymentHandler) Create(c *gin.Context) {
	var req dto.CreateSupplierInvoiceDownPaymentRequest
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
		if err == usecase.ErrPurchaseOrderNotFound {
			errors.NotFoundResponse(c, "purchase_order", fmt.Sprintf("%v", req.PurchaseOrderID))
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrSupplierInvoiceConflict || err == usecase.ErrSupplierInvoiceInvalid {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

// Update handles PUT /purchase/supplier-invoice-down-payments/:id
func (h *SupplierInvoiceDownPaymentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}

	var req dto.UpdateSupplierInvoiceDownPaymentRequest
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
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrPurchaseOrderNotFound {
			errors.NotFoundResponse(c, "purchase_order", fmt.Sprintf("%v", req.PurchaseOrderID))
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrSupplierInvoiceConflict || err == usecase.ErrSupplierInvoiceInvalid {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Delete handles DELETE /purchase/supplier-invoice-down-payments/:id
func (h *SupplierInvoiceDownPaymentHandler) Delete(c *gin.Context) {
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
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrSupplierInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, map[string]interface{}{"id": id}, nil)
}

// Pending handles POST /purchase/supplier-invoice-down-payments/:id/pending
func (h *SupplierInvoiceDownPaymentHandler) Pending(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Pending(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrSupplierInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Submit handles POST /purchase/supplier-invoice-down-payments/:id/submit
func (h *SupplierInvoiceDownPaymentHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Submit(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrSupplierInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Approve handles POST /purchase/supplier-invoice-down-payments/:id/approve
func (h *SupplierInvoiceDownPaymentHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Approve(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrSupplierInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Reject handles POST /purchase/supplier-invoice-down-payments/:id/reject
func (h *SupplierInvoiceDownPaymentHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Reject(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrSupplierInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Cancel handles POST /purchase/supplier-invoice-down-payments/:id/cancel
func (h *SupplierInvoiceDownPaymentHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Cancel(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			errors.NotFoundResponse(c, "supplier_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrSupplierInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// AuditTrail handles GET /purchase/supplier-invoice-down-payments/:id/audit-trail
func (h *SupplierInvoiceDownPaymentHandler) AuditTrail(c *gin.Context) {
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

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// Export handles GET /purchase/supplier-invoice-down-payments/export
func (h *SupplierInvoiceDownPaymentHandler) Export(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	if limit < 1 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		items, _, err := h.uc.List(ctx, repositories.SupplierInvoiceListParams{
			Search:  search,
			Status:  status,
			SortBy:  sortBy,
			SortDir: sortDir,
			Limit:   limit,
			Offset:  0,
		})
		if err != nil {
			return nil, err
		}

		var b strings.Builder
		b.WriteString("code,invoice_number,invoice_date,due_date,purchase_order_code,amount,status,created_at\n")
		for _, it := range items {
			poCode := ""
			if it.PurchaseOrder != nil {
				poCode = it.PurchaseOrder.Code
			}
			row := []string{
				csvEscape(it.Code),
				csvEscape(it.InvoiceNumber),
				csvEscape(it.InvoiceDate),
				csvEscape(it.DueDate),
				csvEscape(poCode),
				csvEscape(fmt.Sprintf("%v", it.Amount)),
				csvEscape(it.Status),
				csvEscape(it.CreatedAt.String()),
			}
			b.WriteString(strings.Join(row, ","))
			b.WriteString("\n")
		}

		return &exportjob.GeneratedFile{
			FileName:    "supplier_invoice_down_payments.csv",
			ContentType: "text/csv; charset=utf-8",
			Bytes:       []byte(b.String()),
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
