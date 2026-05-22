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

type PurchaseOrderHandler struct {
	uc usecase.PurchaseOrderUsecase
}

func NewPurchaseOrderHandler(uc usecase.PurchaseOrderUsecase) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{uc: uc}
}

// Add handles GET /purchase/purchase-orders/add
func (h *PurchaseOrderHandler) Add(c *gin.Context) {
	data, err := h.uc.AddData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, data, nil)
}

// List handles GET /purchase/purchase-orders
func (h *PurchaseOrderHandler) List(c *gin.Context) {
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
	supplierID := strings.TrimSpace(c.Query("supplier_id"))
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	params := repositories.PurchaseOrderListParams{
		Search:     search,
		Status:     status,
		SupplierID: supplierID,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Limit:      perPage,
		Offset:     (page - 1) * perPage,
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
	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}
	if strings.TrimSpace(params.Status) != "" {
		meta.Filters["status"] = params.Status
	}
	if strings.TrimSpace(params.SupplierID) != "" {
		meta.Filters["supplier_id"] = params.SupplierID
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// GetByID handles GET /purchase/purchase-orders/:id
func (h *PurchaseOrderHandler) GetByID(c *gin.Context) {
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
		if err == usecase.ErrPurchaseOrderNotFound {
			errors.NotFoundResponse(c, "purchase_order", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Create handles POST /purchase/purchase-orders
func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	var req dto.CreatePurchaseOrderRequest
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
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", fmt.Sprintf("%v", req.PurchaseRequisitionID))
			return
		}
		if err == usecase.ErrSalesOrderNotFound {
			errors.NotFoundResponse(c, "sales_order", fmt.Sprintf("%v", req.SalesOrderID))
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrPurchaseOrderConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

// Update handles PUT /purchase/purchase-orders/:id
func (h *PurchaseOrderHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}

	var req dto.UpdatePurchaseOrderRequest
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
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrPurchaseOrderNotFound {
			errors.NotFoundResponse(c, "purchase_order", id)
			return
		}
		if err == usecase.ErrPurchaseOrderConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Delete handles DELETE /purchase/purchase-orders/:id
func (h *PurchaseOrderHandler) Delete(c *gin.Context) {
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
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrPurchaseOrderNotFound {
			errors.NotFoundResponse(c, "purchase_order", id)
			return
		}
		if err == usecase.ErrPurchaseOrderConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, map[string]interface{}{"id": id}, nil)
}

// Submit handles POST /purchase/purchase-orders/:id/submit
func (h *PurchaseOrderHandler) Submit(c *gin.Context) {
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
		handlePurchaseOrderError(c, err)
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Approve handles POST /purchase/purchase-orders/:id/approve
func (h *PurchaseOrderHandler) Approve(c *gin.Context) {
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
		handlePurchaseOrderError(c, err)
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Reject handles POST /purchase/purchase-orders/:id/reject
func (h *PurchaseOrderHandler) Reject(c *gin.Context) {
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
		handlePurchaseOrderError(c, err)
		return
	}
	response.SuccessResponse(c, res, nil)
}

// AuditTrail handles GET /purchase/purchase-orders/:id/audit-trail
func (h *PurchaseOrderHandler) AuditTrail(c *gin.Context) {
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

// Export handles GET /purchase/purchase-orders/export
func (h *PurchaseOrderHandler) Export(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		items, _, err := h.uc.List(ctx, repositories.PurchaseOrderListParams{
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
		b.WriteString("code,order_date,due_date,status,total_amount\n")
		for _, it := range items {
			due := ""
			if it.DueDate != nil {
				due = *it.DueDate
			}
			row := []string{
				csvEscape(it.Code),
				csvEscape(it.OrderDate),
				csvEscape(due),
				csvEscape(it.Status),
				fmt.Sprintf("%v", it.TotalAmount),
			}
			b.WriteString(strings.Join(row, ","))
			b.WriteString("\n")
		}

		return &exportjob.GeneratedFile{
			FileName:    "purchase_orders.csv",
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

// handlePurchaseOrderError centralises error-to-HTTP mapping for purchase order actions.
func handlePurchaseOrderError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	switch {
	case err.Error() == "user not authenticated":
		errors.UnauthorizedResponse(c, "user not authenticated")
	case err == usecase.ErrPurchaseOrderNotFound:
		errors.NotFoundResponse(c, "purchase_order", c.Param("id"))
	case err == usecase.ErrPurchaseOrderConflict:
		errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
	case err == usecase.ErrInvalidStatus:
		errors.ErrorResponse(c, "INVALID_STATUS", map[string]interface{}{"message": "Purchase order cannot be approved or rejected in its current status. Only submitted purchase orders can be approved or rejected."}, nil)
	default:
		errors.InternalServerErrorResponse(c, err.Error())
	}
}
