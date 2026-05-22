package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	stderrors "errors"
	"fmt"
	"strconv"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CustomerInvoiceHandler handles customer invoice HTTP requests
type CustomerInvoiceHandler struct {
	invoiceUC usecase.CustomerInvoiceUsecase
}

// NewCustomerInvoiceHandler creates a new CustomerInvoiceHandler
func NewCustomerInvoiceHandler(invoiceUC usecase.CustomerInvoiceUsecase) *CustomerInvoiceHandler {
	return &CustomerInvoiceHandler{invoiceUC: invoiceUC}
}

// List handles list customer invoices request
func (h *CustomerInvoiceHandler) List(c *gin.Context) {
	var req dto.ListCustomerInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	invoices, pagination, err := h.invoiceUC.List(c.Request.Context(), &req)
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
	if req.Type != "" {
		meta.Filters["type"] = req.Type
	}
	if req.SalesOrderID != "" {
		meta.Filters["sales_order_id"] = req.SalesOrderID
	}

	response.SuccessResponse(c, invoices, meta)
}

// GetByID handles get customer invoice by ID request
func (h *CustomerInvoiceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.invoiceUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.ErrorResponse(c, "CUSTOMER_INVOICE_NOT_FOUND", map[string]interface{}{
				"invoice_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, invoice, nil)
}

// Create handles create customer invoice request
func (h *CustomerInvoiceHandler) Create(c *gin.Context) {
	var req dto.CreateCustomerInvoiceRequest
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

	invoice, err := h.invoiceUC.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		if err == usecase.ErrProductNotFound {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"message": "One or more products not found",
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrInvoiceExceedsRemaining) {
			errors.ErrorResponse(c, "INVOICE_EXCEEDS_REMAINING", map[string]interface{}{
				"message": err.Error(),
			}, nil)
			return
		}
		if err == usecase.ErrInvoiceDOMismatch {
			errors.ErrorResponse(c, "DELIVERY_ORDER_MISMATCH", map[string]interface{}{
				"message": "Delivery order does not belong to the same sales order",
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrSalesOrderNotFound) {
			errors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"message": err.Error(),
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

	response.SuccessResponseCreated(c, invoice, meta)
}

// PreviewJournal handles POST /customer-invoices/preview
func (h *CustomerInvoiceHandler) PreviewJournal(c *gin.Context) {
	var req dto.CreateCustomerInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	preview, err := h.invoiceUC.PreviewJournal(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, preview, nil)
}

// Update handles update customer invoice request
func (h *CustomerInvoiceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateCustomerInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	invoice, err := h.invoiceUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.ErrorResponse(c, "CUSTOMER_INVOICE_NOT_FOUND", map[string]interface{}{
				"invoice_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidInvoiceStatus {
			errors.ErrorResponse(c, "INVALID_INVOICE_STATUS", map[string]interface{}{
				"message": "Cannot modify invoice in current status",
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

	response.SuccessResponse(c, invoice, meta)
}

// Delete handles delete customer invoice request
func (h *CustomerInvoiceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.invoiceUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.ErrorResponse(c, "CUSTOMER_INVOICE_NOT_FOUND", map[string]interface{}{
				"invoice_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidInvoiceStatus {
			errors.ErrorResponse(c, "INVALID_INVOICE_STATUS", map[string]interface{}{
				"message": "Cannot delete invoice in current status",
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

	response.SuccessResponseDeleted(c, "customer_invoice", id, meta)
}

// UpdateStatus handles update customer invoice status request
func (h *CustomerInvoiceHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateCustomerInvoiceStatusRequest
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

	invoice, err := h.invoiceUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.ErrorResponse(c, "CUSTOMER_INVOICE_NOT_FOUND", map[string]interface{}{
				"invoice_id": id,
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

	response.SuccessResponse(c, invoice, meta)
}

// Approve handles approve customer invoice request (sent → approved)
func (h *CustomerInvoiceHandler) Approve(c *gin.Context) {
	id := c.Param("id")

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if u, ok := uid.(string); ok {
			userID = &u
		}
	}

	req := dto.UpdateCustomerInvoiceStatusRequest{Status: "APPROVED"}
	invoice, err := h.invoiceUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.ErrorResponse(c, "CUSTOMER_INVOICE_NOT_FOUND", map[string]interface{}{
				"invoice_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidStatusTransition {
			errors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Invoice must be in sent status to approve",
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

	response.SuccessResponse(c, invoice, meta)
}

// ListItems handles list customer invoice items request with pagination
func (h *CustomerInvoiceHandler) ListItems(c *gin.Context) {
	invoiceID := c.Param("id")

	var req dto.ListCustomerInvoiceItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	items, pagination, err := h.invoiceUC.ListItems(c.Request.Context(), invoiceID, &req)
	if err != nil {
		if err == usecase.ErrCustomerInvoiceNotFound {
			errors.ErrorResponse(c, "CUSTOMER_INVOICE_NOT_FOUND", map[string]interface{}{
				"invoice_id": invoiceID,
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

// Export handles CSV export for customer invoices.
func (h *CustomerInvoiceHandler) Export(c *gin.Context) {
	var req dto.ListCustomerInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	generator := func(ctx context.Context, setProgress func(int)) (*exportjob.GeneratedFile, error) {
		req.Page = 1
		req.PerPage = 100

		invoices, pagination, err := h.invoiceUC.List(ctx, &req)
		if err != nil {
			return nil, err
		}

		totalPages := pagination.TotalPages
		if totalPages < 1 {
			totalPages = 1
		}

		setProgress(10)

		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)
		if err := writer.Write([]string{"code", "invoice_date", "due_date", "type", "status", "amount", "paid_amount", "remaining_amount"}); err != nil {
			return nil, err
		}

		writeInvoices := func(rows []dto.CustomerInvoiceResponse) error {
			for _, row := range rows {
				dueDate := ""
				if row.DueDate != nil {
					dueDate = *row.DueDate
				}
				if err := writer.Write([]string{
					row.Code,
					row.InvoiceDate,
					dueDate,
					row.Type,
					row.Status,
					fmt.Sprintf("%.2f", row.Amount),
					fmt.Sprintf("%.2f", row.PaidAmount),
					fmt.Sprintf("%.2f", row.RemainingAmount),
				}); err != nil {
					return err
				}
			}
			return nil
		}

		if err := writeInvoices(invoices); err != nil {
			return nil, err
		}

		for page := 2; page <= totalPages; page++ {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			req.Page = page
			rows, _, err := h.invoiceUC.List(ctx, &req)
			if err != nil {
				return nil, err
			}
			if err := writeInvoices(rows); err != nil {
				return nil, err
			}

			setProgress(utils.LinearProgress(page, totalPages, 10, 90))
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			return nil, err
		}

		setProgress(95)
		fileName := fmt.Sprintf("customer_invoices_%s.csv", apptime.Now().Format("20060102150405"))
		return &exportjob.GeneratedFile{
			FileName:    fileName,
			ContentType: "text/csv; charset=utf-8",
			Bytes:       buffer.Bytes(),
		}, nil
	}

	if exportjob.QueueIfRequestedWithProgress(c, generator) {
		return
	}

	file, err := generator(c.Request.Context(), utils.NoopProgress)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	exportjob.WriteSyncFile(c, file)
}

// AuditTrail handles list customer invoice audit trail with pagination.
func (h *CustomerInvoiceHandler) AuditTrail(c *gin.Context) {
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

	entries, total, err := h.invoiceUC.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, entries, meta)
}
