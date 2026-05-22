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

type PurchaseRequisitionHandler struct {
	uc   usecase.PurchaseRequisitionUsecase
	poUc usecase.PurchaseOrderUsecase
}

func NewPurchaseRequisitionHandler(uc usecase.PurchaseRequisitionUsecase, poUc usecase.PurchaseOrderUsecase) *PurchaseRequisitionHandler {
	return &PurchaseRequisitionHandler{uc: uc, poUc: poUc}
}

// List handles GET /purchase/purchase-requisitions
func (h *PurchaseRequisitionHandler) List(c *gin.Context) {
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

	params := repositories.PurchaseRequisitionListParams{
		Search:  c.Query("search"),
		Status:  c.Query("status"),
		SortBy:  c.DefaultQuery("sort_by", "created_at"),
		SortDir: c.DefaultQuery("sort_dir", "desc"),
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
	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// Create handles POST /purchase/purchase-requisitions
func (h *PurchaseRequisitionHandler) Create(c *gin.Context) {
	var req dto.CreatePurchaseRequisitionRequest
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
		if err == usecase.ErrRequestDateInPast {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error(), "field": "request_date"}, nil)
			return
		}
		// Duplicate code (unique constraint) should be a conflict, not a 500.
		errStr := err.Error()
		if strings.Contains(errStr, "idx_purchase_requisitions_code") || strings.Contains(errStr, "SQLSTATE 23505") {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": "Purchase requisition code already exists"}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

// GetByID handles GET /purchase/purchase-requisitions/:id
func (h *PurchaseRequisitionHandler) GetByID(c *gin.Context) {
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
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

// Update handles PUT /purchase/purchase-requisitions/:id
func (h *PurchaseRequisitionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}

	var req dto.UpdatePurchaseRequisitionRequest
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
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", id)
			return
		}
		if err == usecase.ErrRequestDateInPast {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error(), "field": "request_date"}, nil)
			return
		}
		if err == usecase.ErrInvalidStatus {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

// Delete handles DELETE /purchase/purchase-requisitions/:id
func (h *PurchaseRequisitionHandler) Delete(c *gin.Context) {
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
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", id)
			return
		}
		if err == usecase.ErrInvalidStatus {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "purchase_requisition", id, nil)
}

// AddData handles GET /purchase/purchase-requisitions/add
func (h *PurchaseRequisitionHandler) AddData(c *gin.Context) {
	res, err := h.uc.AddData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Submit handles POST /purchase/purchase-requisitions/:id/submit
func (h *PurchaseRequisitionHandler) Submit(c *gin.Context) {
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
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", id)
			return
		}
		if err == usecase.ErrInvalidStatus {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

// Approve handles POST /purchase/purchase-requisitions/:id/approve
func (h *PurchaseRequisitionHandler) Approve(c *gin.Context) {
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
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", id)
			return
		}
		if err == usecase.ErrInvalidStatus {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

// Reject handles POST /purchase/purchase-requisitions/:id/reject
func (h *PurchaseRequisitionHandler) Reject(c *gin.Context) {
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
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", id)
			return
		}
		if err == usecase.ErrInvalidStatus {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

// Convert handles POST /purchase/purchase-requisitions/:id/convert
func (h *PurchaseRequisitionHandler) Convert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}

	if h.poUc == nil {
		errors.InternalServerErrorResponse(c, "purchase order usecase is nil")
		return
	}

	po, err := h.poUc.CreateFromPurchaseRequisition(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrPurchaseRequisitionNotFound {
			errors.NotFoundResponse(c, "purchase_requisition", id)
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrPurchaseOrderConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"purchase_requisition_id": id,
		"purchase_order_id":       po.ID,
		"purchase_order_code":     po.Code,
	}, nil)
}

// AuditTrail handles GET /purchase/purchase-requisitions/:id/audit-trail
func (h *PurchaseRequisitionHandler) AuditTrail(c *gin.Context) {
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

// Export handles GET /purchase/purchase-requisitions/export
func (h *PurchaseRequisitionHandler) Export(c *gin.Context) {
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	if limit < 1 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	params := repositories.PurchaseRequisitionListParams{
		Search:  search,
		SortBy:  sortBy,
		SortDir: sortDir,
		Limit:   limit,
		Offset:  0,
	}

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		items, _, err := h.uc.List(ctx, params)
		if err != nil {
			return nil, err
		}

		var b strings.Builder
		b.WriteString("code,request_date,supplier,status,subtotal,tax_amount,delivery_cost,other_cost,total_amount,notes,created_at\n")
		for _, it := range items {
			supplier := ""
			if it.Supplier != nil {
				supplier = it.Supplier.Name
			}
			row := []string{
				csvEscape(it.Code),
				csvEscape(it.RequestDate),
				csvEscape(supplier),
				csvEscape(string(it.Status)),
				fmt.Sprintf("%v", it.Subtotal),
				fmt.Sprintf("%v", it.TaxAmount),
				fmt.Sprintf("%v", it.DeliveryCost),
				fmt.Sprintf("%v", it.OtherCost),
				fmt.Sprintf("%v", it.TotalAmount),
				csvEscape(it.Notes),
				csvEscape(it.CreatedAt),
			}
			b.WriteString(strings.Join(row, ","))
			b.WriteString("\n")
		}

		return &exportjob.GeneratedFile{
			FileName:    "purchase_requisitions.csv",
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

func csvEscape(s string) string {
	if s == "" {
		return ""
	}
	needsQuote := strings.ContainsAny(s, ",\n\r\"")
	if !needsQuote {
		return s
	}
	escaped := strings.ReplaceAll(s, "\"", "\"\"")
	return "\"" + escaped + "\""
}
