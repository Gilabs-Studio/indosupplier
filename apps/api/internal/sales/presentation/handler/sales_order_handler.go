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

// SalesOrderHandler handles sales order HTTP requests
type SalesOrderHandler struct {
	orderUC usecase.SalesOrderUsecase
}

// NewSalesOrderHandler creates a new SalesOrderHandler
func NewSalesOrderHandler(orderUC usecase.SalesOrderUsecase) *SalesOrderHandler {
	return &SalesOrderHandler{orderUC: orderUC}
}

// List handles list sales orders request
func (h *SalesOrderHandler) List(c *gin.Context) {
	var req dto.ListSalesOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	orders, pagination, err := h.orderUC.List(c.Request.Context(), &req)
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
	if req.SalesQuotationID != "" {
		meta.Filters["sales_quotation_id"] = req.SalesQuotationID
	}

	response.SuccessResponse(c, orders, meta)
}

// GetByID handles get sales order by ID request
func (h *SalesOrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	order, err := h.orderUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrSalesOrderNotFound) {
			errors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"order_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, order, nil)
}

// Create handles create sales order request
func (h *SalesOrderHandler) Create(c *gin.Context) {
	var req dto.CreateSalesOrderRequest
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

	order, err := h.orderUC.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		if err == usecase.ErrOrderProductNotFound {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"message": "One or more products not found",
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrCreditLimitExceeded) {
			errors.ErrorResponse(c, "CREDIT_LIMIT_EXCEEDED", map[string]interface{}{
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

	response.SuccessResponseCreated(c, order, meta)
}

// Update handles update sales order request
func (h *SalesOrderHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	order, err := h.orderUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrSalesOrderNotFound {
			errors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidOrderStatus {
			errors.ErrorResponse(c, "INVALID_ORDER_STATUS", map[string]interface{}{
				"message": "Cannot modify order in current status",
			}, nil)
			return
		}
		if err == usecase.ErrOrderProductNotFound {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"message": "One or more products not found",
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrCreditLimitExceeded) {
			errors.ErrorResponse(c, "CREDIT_LIMIT_EXCEEDED", map[string]interface{}{
				"message": err.Error(),
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

	response.SuccessResponse(c, order, meta)
}

// Delete handles delete sales order request
func (h *SalesOrderHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.orderUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesOrderNotFound {
			errors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidOrderStatus {
			errors.ErrorResponse(c, "INVALID_ORDER_STATUS", map[string]interface{}{
				"message": "Cannot delete order in current status",
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

	response.SuccessResponseDeleted(c, "sales_order", id, meta)
}

// UpdateStatus handles update sales order status request
func (h *SalesOrderHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalesOrderStatusRequest
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

	order, err := h.orderUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrSalesOrderNotFound {
			errors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidOrderStatusTransition {
			errors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Invalid status transition",
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrCreditLimitExceeded) {
			errors.ErrorResponse(c, "CREDIT_LIMIT_EXCEEDED", map[string]interface{}{
				"message": err.Error(),
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

	response.SuccessResponse(c, order, meta)
}

// Approve handles approve sales order request (sent → approved)
func (h *SalesOrderHandler) Approve(c *gin.Context) {
	id := c.Param("id")

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if u, ok := uid.(string); ok {
			userID = &u
		}
	}

	req := dto.UpdateSalesOrderStatusRequest{Status: "approved"}
	order, err := h.orderUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrSalesOrderNotFound {
			errors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidOrderStatusTransition {
			errors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Order must be in sent status to approve",
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrCreditLimitExceeded) {
			errors.ErrorResponse(c, "CREDIT_LIMIT_EXCEEDED", map[string]interface{}{
				"message": err.Error(),
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

	response.SuccessResponse(c, order, meta)
}

// ConvertFromQuotation handles convert quotation to order request
func (h *SalesOrderHandler) ConvertFromQuotation(c *gin.Context) {
	var req dto.ConvertFromQuotationRequest
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

	order, err := h.orderUC.ConvertFromQuotation(c.Request.Context(), &req, createdBy)
	if err != nil {
		if err == usecase.ErrQuotationNotFound {
			errors.ErrorResponse(c, "SALES_QUOTATION_NOT_FOUND", map[string]interface{}{
				"quotation_id": req.QuotationID,
			}, nil)
			return
		}
		if err == usecase.ErrQuotationNotApproved {
			errors.ErrorResponse(c, "QUOTATION_NOT_APPROVED", map[string]interface{}{
				"message": "Quotation must be approved before converting to order",
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

	response.SuccessResponseCreated(c, order, meta)
}

// ListItems handles list sales order items request with pagination
func (h *SalesOrderHandler) ListItems(c *gin.Context) {
	orderID := c.Param("id")

	var req dto.ListSalesOrderItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	items, pagination, err := h.orderUC.ListItems(c.Request.Context(), orderID, &req)
	if err != nil {
		if err == usecase.ErrSalesOrderNotFound {
			errors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"order_id": orderID,
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

// Export handles CSV export for sales orders.
func (h *SalesOrderHandler) Export(c *gin.Context) {
	var req dto.ListSalesOrdersRequest
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

		orders, pagination, err := h.orderUC.List(ctx, &req)
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
		if err := writer.Write([]string{"code", "order_date", "customer_name", "sales_rep", "status", "total_amount"}); err != nil {
			return nil, err
		}

		writeOrders := func(rows []dto.SalesOrderResponse) error {
			for _, row := range rows {
				salesRep := ""
				if row.SalesRep != nil {
					salesRep = row.SalesRep.Name
				}
				if err := writer.Write([]string{
					row.Code,
					row.OrderDate,
					row.CustomerName,
					salesRep,
					row.Status,
					fmt.Sprintf("%.2f", row.TotalAmount),
				}); err != nil {
					return err
				}
			}
			return nil
		}

		if err := writeOrders(orders); err != nil {
			return nil, err
		}

		for page := 2; page <= totalPages; page++ {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			req.Page = page
			rows, _, err := h.orderUC.List(ctx, &req)
			if err != nil {
				return nil, err
			}
			if err := writeOrders(rows); err != nil {
				return nil, err
			}

			setProgress(utils.LinearProgress(page, totalPages, 10, 90))
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			return nil, err
		}

		setProgress(95)
		fileName := fmt.Sprintf("sales_orders_%s.csv", apptime.Now().Format("20060102150405"))
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

// AuditTrail handles list sales order audit trail with pagination.
func (h *SalesOrderHandler) AuditTrail(c *gin.Context) {
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

	entries, total, err := h.orderUC.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, entries, meta)
}
