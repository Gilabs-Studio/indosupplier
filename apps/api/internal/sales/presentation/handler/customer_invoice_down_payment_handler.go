package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CustomerInvoiceDownPaymentHandler struct {
	uc usecase.CustomerInvoiceDownPaymentUsecase
}

func NewCustomerInvoiceDownPaymentHandler(uc usecase.CustomerInvoiceDownPaymentUsecase) *CustomerInvoiceDownPaymentHandler {
	return &CustomerInvoiceDownPaymentHandler{uc: uc}
}

// Add handles GET /sales/customer-invoice-down-payments/add
func (h *CustomerInvoiceDownPaymentHandler) Add(c *gin.Context) {
	data, err := h.uc.AddData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, data, nil)
}

// List handles GET /sales/customer-invoice-down-payments
func (h *CustomerInvoiceDownPaymentHandler) List(c *gin.Context) {
	var req dto.ListCustomerInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errors.ErrorResponse(c, "INVALID_QUERY_PARAMS", map[string]interface{}{"message": "Invalid query parameters"}, nil)
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	items, total, err := h.uc.List(c.Request.Context(), &req)
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
			Field: req.SortBy,
			Order: req.SortDir,
		},
	}
	if strings.TrimSpace(req.Search) != "" {
		meta.Filters["search"] = req.Search
	}
	if strings.TrimSpace(req.Status) != "" {
		meta.Filters["status"] = req.Status
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// GetByID handles GET /sales/customer-invoice-down-payments/:id
func (h *CustomerInvoiceDownPaymentHandler) GetByID(c *gin.Context) {
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
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.NotFoundResponse(c, "customer_invoice_down_payment", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Create handles POST /sales/customer-invoice-down-payments
func (h *CustomerInvoiceDownPaymentHandler) Create(c *gin.Context) {
	var req dto.CreateCustomerInvoiceDownPaymentRequest
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
		if err == usecase.ErrSalesOrderNotFound {
			errors.NotFoundResponse(c, "sales_order", req.SalesOrderID)
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrCustomerInvoiceConflict || err == usecase.ErrCustomerInvoiceInvalid {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

// Update handles PUT /sales/customer-invoice-down-payments/:id
func (h *CustomerInvoiceDownPaymentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}

	var req dto.UpdateCustomerInvoiceDownPaymentRequest
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
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.NotFoundResponse(c, "customer_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrSalesOrderNotFound {
			errors.NotFoundResponse(c, "sales_order", req.SalesOrderID)
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrCustomerInvoiceConflict || err == usecase.ErrCustomerInvoiceInvalid {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Delete handles DELETE /sales/customer-invoice-down-payments/:id
func (h *CustomerInvoiceDownPaymentHandler) Delete(c *gin.Context) {
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
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.NotFoundResponse(c, "customer_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrCustomerInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, map[string]interface{}{"id": id}, nil)
}

// Pending handles POST /sales/customer-invoice-down-payments/:id/pending
func (h *CustomerInvoiceDownPaymentHandler) Pending(c *gin.Context) {
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
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.NotFoundResponse(c, "customer_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrCustomerInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Approve handles POST /sales/customer-invoice-down-payments/:id/approve
func (h *CustomerInvoiceDownPaymentHandler) Approve(c *gin.Context) {
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
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.NotFoundResponse(c, "customer_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrCustomerInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Cancel handles POST /sales/customer-invoice-down-payments/:id/cancel
func (h *CustomerInvoiceDownPaymentHandler) Cancel(c *gin.Context) {
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
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.NotFoundResponse(c, "customer_invoice_down_payment", id)
			return
		}
		if err == usecase.ErrCustomerInvoiceConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// AuditTrail handles GET /sales/customer-invoice-down-payments/:id/audit-trail
func (h *CustomerInvoiceDownPaymentHandler) AuditTrail(c *gin.Context) {
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

// Export handles GET /sales/customer-invoice-down-payments/export
func (h *CustomerInvoiceDownPaymentHandler) Export(c *gin.Context) {
	var req dto.ListCustomerInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errors.ErrorResponse(c, "INVALID_QUERY_PARAMS", map[string]interface{}{"message": "Invalid query parameters"}, nil)
		return
	}

	req.Page = 1
	req.PerPage = 1000

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		items, _, err := h.uc.List(ctx, &req)
		if err != nil {
			return nil, err
		}

		var b strings.Builder
		b.WriteString("code,invoice_number,invoice_date,due_date,sales_order_code,amount,status,created_at\n")
		for _, it := range items {
			soCode := ""
			if it.SalesOrder != nil {
				soCode = it.SalesOrder.Code
			}
			invNo := ""
			if it.InvoiceNumber != nil {
				invNo = *it.InvoiceNumber
			}
			dueDate := ""
			if it.DueDate != nil {
				dueDate = *it.DueDate
			}
			row := []string{
				it.Code,
				invNo,
				it.InvoiceDate,
				dueDate,
				soCode,
				fmt.Sprintf("%v", it.Amount),
				it.Status,
				it.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			b.WriteString(strings.Join(row, ","))
			b.WriteString("\n")
		}

		return &exportjob.GeneratedFile{
			FileName:    "customer_invoice_down_payments.csv",
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
