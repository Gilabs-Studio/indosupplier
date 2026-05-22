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

type GoodsReceiptHandler struct {
	uc usecase.GoodsReceiptUsecase
}

func NewGoodsReceiptHandler(uc usecase.GoodsReceiptUsecase) *GoodsReceiptHandler {
	return &GoodsReceiptHandler{uc: uc}
}

// Add handles GET /purchase/goods-receipt/add
func (h *GoodsReceiptHandler) Add(c *gin.Context) {
	data, err := h.uc.AddData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, data, nil)
}

// List handles GET /purchase/goods-receipt
func (h *GoodsReceiptHandler) List(c *gin.Context) {
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
	warehouseID := c.Query("warehouse_id")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	params := repositories.GoodsReceiptListParams{
		Search:      search,
		Status:      status,
		WarehouseID: warehouseID,
		SortBy:      sortBy,
		SortDir:     sortDir,
		Limit:       perPage,
		Offset:      (page - 1) * perPage,
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
	if strings.TrimSpace(params.WarehouseID) != "" {
		meta.Filters["warehouse_id"] = params.WarehouseID
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// GetByID handles GET /purchase/goods-receipt/:id
func (h *GoodsReceiptHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	// Accept both UUID and code/reference number — usecase resolves accordingly.
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Create handles POST /purchase/goods-receipt
func (h *GoodsReceiptHandler) Create(c *gin.Context) {
	var req dto.CreateGoodsReceiptRequest
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
		if err == usecase.ErrPurchaseOrderNotFound {
			errors.NotFoundResponse(c, "purchase_order", fmt.Sprintf("%v", req.PurchaseOrderID))
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

// Update handles PUT /purchase/goods-receipt/:id
func (h *GoodsReceiptHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}

	var req dto.UpdateGoodsReceiptRequest
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
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Delete handles DELETE /purchase/goods-receipt/:id
func (h *GoodsReceiptHandler) Delete(c *gin.Context) {
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
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, map[string]interface{}{"id": id}, nil)
}

// Confirm handles POST /purchase/goods-receipt/:id/confirm
func (h *GoodsReceiptHandler) Confirm(c *gin.Context) {
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
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Submit handles POST /purchase/goods-receipt/:id/submit
func (h *GoodsReceiptHandler) Submit(c *gin.Context) {
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
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Approve handles POST /purchase/goods-receipt/:id/approve
func (h *GoodsReceiptHandler) Approve(c *gin.Context) {
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
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Reject handles POST /purchase/goods-receipt/:id/reject
func (h *GoodsReceiptHandler) Reject(c *gin.Context) {
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
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// CloseGR handles POST /purchase/goods-receipt/:id/close
func (h *GoodsReceiptHandler) CloseGR(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.Close(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrInvalidStatus || err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// ConvertToSupplierInvoice handles POST /purchase/goods-receipt/:id/convert
func (h *GoodsReceiptHandler) ConvertToSupplierInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
		return
	}
	res, err := h.uc.ConvertToSupplierInvoice(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not authenticated" {
			errors.UnauthorizedResponse(c, "user not authenticated")
			return
		}
		if err == usecase.ErrGoodsReceiptNotFound {
			errors.NotFoundResponse(c, "goods_receipt", id)
			return
		}
		if err == usecase.ErrGoodsReceiptConflict {
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// AuditTrail handles GET /purchase/goods-receipt/:id/audit-trail
func (h *GoodsReceiptHandler) AuditTrail(c *gin.Context) {
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

// Export handles GET /purchase/goods-receipt/export
func (h *GoodsReceiptHandler) Export(c *gin.Context) {
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
		items, _, err := h.uc.List(ctx, repositories.GoodsReceiptListParams{
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
		b.WriteString("code,purchase_order_code,receipt_date,status\n")
		for _, it := range items {
			poCode := ""
			if it.PurchaseOrder != nil {
				poCode = it.PurchaseOrder.Code
			}
			rd := ""
			if it.ReceiptDate != nil {
				rd = *it.ReceiptDate
			}
			row := []string{
				csvEscape(it.Code),
				csvEscape(poCode),
				csvEscape(rd),
				csvEscape(it.Status),
			}
			b.WriteString(strings.Join(row, ","))
			b.WriteString("\n")
		}

		return &exportjob.GeneratedFile{
			FileName:    "goods_receipts.csv",
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
