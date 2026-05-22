package handler

import (
	"math"
	"strconv"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gin-gonic/gin"
)

// POSOrderHandler handles POS order operations
type POSOrderHandler struct {
	uc usecase.POSOrderUsecase
}

// NewPOSOrderHandler creates the handler
func NewPOSOrderHandler(uc usecase.POSOrderUsecase) *POSOrderHandler {
	return &POSOrderHandler{uc: uc}
}

// Create opens a new empty order for the selected outlet/table context
func (h *POSOrderHandler) Create(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	order, err := h.uc.Create(c.Request.Context(), &req, uc.userID)
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// GetByID returns a single order with items and payments
func (h *POSOrderHandler) GetByID(c *gin.Context) {
	order, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// List returns paginated orders
func (h *POSOrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}

	params := repositories.POSOrderListParams{
		SessionID: c.Query("session_id"),
		OutletID:  c.Query("outlet_id"),
		Status:    c.Query("status"),
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	}

	orders, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	hasNext := page < totalPages
	hasPrev := page > 1
	var nextPage, prevPage *int
	if hasNext {
		n := page + 1
		nextPage = &n
	}
	if hasPrev {
		p := page - 1
		prevPage = &p
	}

	response.SuccessResponse(c, orders, &response.Meta{
		Pagination: &response.PaginationMeta{
			Page: page, PerPage: perPage, Total: int(total),
			TotalPages: totalPages, HasNext: hasNext, HasPrev: hasPrev,
			NextPage: nextPage, PrevPage: prevPage,
		},
	})
}

// Confirm validates stock and locks order for payment
func (h *POSOrderHandler) Confirm(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	var req dto.ConfirmOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	order, err := h.uc.Confirm(c.Request.Context(), c.Param("id"), &req, uc.userID)
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// Void cancels the order
func (h *POSOrderHandler) Void(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	var req dto.VoidOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	order, err := h.uc.Void(c.Request.Context(), c.Param("id"), &req, uc.userID)
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// AddItem appends an item to the order
func (h *POSOrderHandler) AddItem(c *gin.Context) {
	var req dto.AddOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	order, err := h.uc.AddItem(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// UpdateItem modifies an existing order item
func (h *POSOrderHandler) UpdateItem(c *gin.Context) {
	var req dto.UpdateOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	order, err := h.uc.UpdateItem(c.Request.Context(), c.Param("id"), c.Param("itemID"), &req)
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// RemoveItem deletes an item from the order
func (h *POSOrderHandler) RemoveItem(c *gin.Context) {
	order, err := h.uc.RemoveItem(c.Request.Context(), c.Param("id"), c.Param("itemID"))
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// GetCatalog returns POS-available products with live stock for the outlet
func (h *POSOrderHandler) GetCatalog(c *gin.Context) {
	items, err := h.uc.GetCatalog(c.Request.Context(), c.Param("outletID"))
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, items, nil)
}

// AssignTable assigns a table to the order
func (h *POSOrderHandler) AssignTable(c *gin.Context) {
	var req dto.AssignTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	order, err := h.uc.AssignTable(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

func handlePOSOrderError(c *gin.Context, err error) {
	switch err {
	case usecase.ErrPOSOrderNotFound:
		coreErrors.NotFoundResponse(c, "pos_order", "")
	case usecase.ErrPOSOrderCannotModify:
		coreErrors.ErrorResponse(c, "POS_ORDER_CANNOT_MODIFY", nil, nil)
	case usecase.ErrPOSOrderItemNotFound:
		coreErrors.NotFoundResponse(c, "pos_order_item", "")
	case usecase.ErrPOSProductNotAvailable:
		coreErrors.ErrorResponse(c, "POS_PRODUCT_NOT_AVAILABLE", nil, nil)
	case usecase.ErrPOSInsufficientStock:
		coreErrors.ErrorResponse(c, "INSUFFICIENT_STOCK", nil, nil)
	case usecase.ErrPOSItemAlreadyServed:
		coreErrors.ErrorResponse(c, "POS_ITEM_ALREADY_SERVED", nil, nil)
	case usecase.ErrPOSOutletForbidden:
		coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
	default:
		coreErrors.InternalServerErrorResponse(c, "")
	}
}

// Serve marks an order as SERVED — food has been delivered to the table.
func (h *POSOrderHandler) Serve(c *gin.Context) {
	order, err := h.uc.MarkServed(c.Request.Context(), c.Param("id"))
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// Complete marks a paid order as COMPLETED — customer has left the table.
func (h *POSOrderHandler) Complete(c *gin.Context) {
	order, err := h.uc.MarkCompleted(c.Request.Context(), c.Param("id"))
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// ServeItem marks a single order item as served.
func (h *POSOrderHandler) ServeItem(c *gin.Context) {
	order, err := h.uc.MarkItemServed(c.Request.Context(), c.Param("id"), c.Param("itemID"))
	if err != nil {
		handlePOSOrderError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// ListOutlets returns all outlets available for POS, filtered by current scope.
func (h *POSOrderHandler) ListOutlets(c *gin.Context) {
	outlets, err := h.uc.ListOutlets(c.Request.Context())
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, outlets, nil)
}
