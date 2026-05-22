package handler

import (
	stdErrors "errors"
	"strconv"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	inventoryUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// DeliveryOrderHandler handles delivery order HTTP requests
type DeliveryOrderHandler struct {
	deliveryOrderUC usecase.DeliveryOrderUsecase
}

// NewDeliveryOrderHandler creates a new DeliveryOrderHandler
func NewDeliveryOrderHandler(deliveryOrderUC usecase.DeliveryOrderUsecase) *DeliveryOrderHandler {
	return &DeliveryOrderHandler{deliveryOrderUC: deliveryOrderUC}
}

// List handles list delivery orders request
func (h *DeliveryOrderHandler) List(c *gin.Context) {
	var req dto.ListDeliveryOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidQueryParamResponse(c)
		return
	}

	deliveryOrders, pagination, err := h.deliveryOrderUC.List(c.Request.Context(), &req)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, err.Error())
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
	if req.SalesOrderID != "" {
		meta.Filters["sales_order_id"] = req.SalesOrderID
	}

	response.SuccessResponse(c, deliveryOrders, meta)
}

// GetByID handles get delivery order by ID request
func (h *DeliveryOrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	deliveryOrder, err := h.deliveryOrderUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": id,
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, deliveryOrder, nil)
}

// Create handles create delivery order request
func (h *DeliveryOrderHandler) Create(c *gin.Context) {
	var req dto.CreateDeliveryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	var createdBy *string
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = &id
		}
	}

	deliveryOrder, err := h.deliveryOrderUC.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		if err == usecase.ErrDeliverySalesOrderNotFound {
			coreErrors.ErrorResponse(c, "SALES_ORDER_NOT_FOUND", map[string]interface{}{
				"sales_order_id": req.SalesOrderID,
			}, nil)
			return
		}
		if err == usecase.ErrDeliveryProductNotFound {
			coreErrors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"message": "One or more products not found",
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if createdBy != nil {
		meta.CreatedBy = *createdBy
	}

	response.SuccessResponseCreated(c, deliveryOrder, meta)
}

// Update handles update delivery order request
func (h *DeliveryOrderHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateDeliveryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	deliveryOrder, err := h.deliveryOrderUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidDeliveryOrderStatus {
			coreErrors.ErrorResponse(c, "INVALID_DELIVERY_ORDER_STATUS", map[string]interface{}{
				"message": "Cannot modify delivery order in current status",
			}, nil)
			return
		}
		if err == usecase.ErrDeliveryProductNotFound {
			coreErrors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"message": "One or more products not found",
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}

	response.SuccessResponse(c, deliveryOrder, meta)
}

// Delete handles delete delivery order request
func (h *DeliveryOrderHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.deliveryOrderUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidDeliveryOrderStatus {
			coreErrors.ErrorResponse(c, "INVALID_DELIVERY_ORDER_STATUS", map[string]interface{}{
				"message": "Cannot delete delivery order in current status",
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "delivery_order", id, meta)
}

// UpdateStatus handles update delivery order status request
func (h *DeliveryOrderHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateDeliveryOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	deliveryOrder, err := h.deliveryOrderUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidDeliveryStatusTransition {
			coreErrors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Invalid status transition",
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID != nil {
		meta.UpdatedBy = *userID
	}

	response.SuccessResponse(c, deliveryOrder, meta)
}

// Approve handles approve delivery order request (sent → approved)
func (h *DeliveryOrderHandler) Approve(c *gin.Context) {
	id := c.Param("id")

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if u, ok := uid.(string); ok {
			userID = &u
		}
	}

	req := dto.UpdateDeliveryOrderStatusRequest{Status: "approved"}
	deliveryOrder, err := h.deliveryOrderUC.UpdateStatus(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidDeliveryStatusTransition {
			coreErrors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Delivery order must be in sent status to approve",
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID != nil {
		meta.UpdatedBy = *userID
	}

	response.SuccessResponse(c, deliveryOrder, meta)
}

// Ship handles ship delivery order request
func (h *DeliveryOrderHandler) Ship(c *gin.Context) {
	id := c.Param("id")
	var req dto.ShipDeliveryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	deliveryOrder, err := h.deliveryOrderUC.Ship(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidDeliveryStatusTransition {
			coreErrors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Delivery order must be in prepared status to ship",
			}, nil)
			return
		}
		if stdErrors.Is(err, usecase.ErrInsufficientBatchStock) || stdErrors.Is(err, inventoryUsecase.ErrInsufficientBatchStock) {
			coreErrors.ErrorResponse(c, "INSUFFICIENT_STOCK", map[string]interface{}{
				"delivery_order_id": id,
				"reason":            err.Error(),
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID != nil {
		meta.UpdatedBy = *userID
	}

	response.SuccessResponse(c, deliveryOrder, meta)
}

// Deliver handles deliver delivery order request
func (h *DeliveryOrderHandler) Deliver(c *gin.Context) {
	id := c.Param("id")
	var req dto.DeliverDeliveryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	deliveryOrder, err := h.deliveryOrderUC.Deliver(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidDeliveryStatusTransition {
			coreErrors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{
				"message": "Delivery order must be in shipped status to deliver",
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID != nil {
		meta.UpdatedBy = *userID
	}

	response.SuccessResponse(c, deliveryOrder, meta)
}

// SelectBatches handles batch selection request (FIFO/FEFO)
func (h *DeliveryOrderHandler) SelectBatches(c *gin.Context) {
	var req dto.BatchSelectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	batches, err := h.deliveryOrderUC.SelectBatches(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrDeliveryProductNotFound {
			coreErrors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{
				"product_id": req.ProductID,
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, batches, nil)
}

// ListItems handles list delivery order items request with pagination
func (h *DeliveryOrderHandler) ListItems(c *gin.Context) {
	deliveryOrderID := c.Param("id")

	var req dto.ListDeliveryOrderItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidQueryParamResponse(c)
		return
	}

	items, pagination, err := h.deliveryOrderUC.ListItems(c.Request.Context(), deliveryOrderID, &req)
	if err != nil {
		if err == usecase.ErrDeliveryOrderNotFound {
			coreErrors.ErrorResponse(c, "DELIVERY_ORDER_NOT_FOUND", map[string]interface{}{
				"delivery_order_id": deliveryOrderID,
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
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

// AuditTrail handles list delivery order audit trail with pagination.
func (h *DeliveryOrderHandler) AuditTrail(c *gin.Context) {
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

	entries, total, err := h.deliveryOrderUC.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, entries, meta)
}
