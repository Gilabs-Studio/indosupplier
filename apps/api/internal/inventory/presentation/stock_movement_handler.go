package presentation

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
	"github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	"github.com/gin-gonic/gin"
)

type StockMovementHandler struct {
	service          usecase.StockMovementService
	inventoryUsecase usecase.InventoryUsecase
}

func NewStockMovementHandler(service usecase.StockMovementService, inventoryUsecase usecase.InventoryUsecase) *StockMovementHandler {
	return &StockMovementHandler{
		service:          service,
		inventoryUsecase: inventoryUsecase,
	}
}

func (h *StockMovementHandler) GetMovements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	req := &dto.GetStockMovementsRequest{
		Page:        page,
		PerPage:     perPage,
		Search:      c.Query("search"),
		WarehouseID: c.Query("warehouse_id"),
		ProductID:   c.Query("product_id"),
		Type:        c.Query("type"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
	}

	movements, pagination, err := h.service.GetMovements(c, req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stock movements", err, nil)
		return
	}

	meta := &response.Meta{
		Pagination: pagination,
	}

	response.SuccessResponse(c, movements, meta)
}

func (h *StockMovementHandler) CreateMovement(c *gin.Context) {
	var req dto.CreateManualMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err, nil)
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication context not found", nil, nil)
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authenticated user context", nil, nil)
		return
	}
	req.CreatedBy = userID

	if err := h.inventoryUsecase.CreateManualStockMovement(c, &req); err != nil {
		if errors.Is(err, usecase.ErrInsufficientStock) {
			response.ErrorResponse(c, http.StatusUnprocessableEntity, "INSUFFICIENT_STOCK", "Requested quantity exceeds available stock", nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create stock movement", err, nil)
		return
	}

	response.SuccessResponse(c, nil, nil)
}
