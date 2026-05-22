package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
	"github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	usecase usecase.InventoryUsecase
}

func NewInventoryHandler(usecase usecase.InventoryUsecase) *InventoryHandler {
	return &InventoryHandler{
		usecase: usecase,
	}
}

func (h *InventoryHandler) GetStockList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	warehouseID := c.Query("warehouse_id")
	productID := c.Query("product_id")
	lowStock := c.Query("low_stock") == "true"
	status := c.Query("status")
	hasExpiring := c.Query("has_expiring") == "true"
	hasExpired := c.Query("has_expired") == "true"

	// Parse optional is_ingredient filter: accept "true"/"false"; absent = no filter
	var isIngredient *bool
	if v := c.Query("is_ingredient"); v == "true" {
		t := true
		isIngredient = &t
	} else if v == "false" {
		f := false
		isIngredient = &f
	}

	req := &dto.GetInventoryListRequest{
		Page:         page,
		PerPage:      perPage,
		Search:       search,
		WarehouseID:  warehouseID,
		ProductID:    productID,
		LowStock:     lowStock,
		Status:       status,
		HasExpiring:  hasExpiring,
		HasExpired:   hasExpired,
		IsIngredient: isIngredient,
	}

	result, err := h.usecase.GetStockList(c.Request.Context(), req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

func (h *InventoryHandler) GetTreeWarehouses(c *gin.Context) {
	result, err := h.usecase.GetTreeWarehouses(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, result, nil)
}

func (h *InventoryHandler) GetTreeProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	warehouseID := c.Query("warehouse_id")

	if warehouseID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "warehouse_id is required", nil, nil)
		return
	}

	// Parse optional is_ingredient filter
	var isIngredient *bool
	if v := c.Query("is_ingredient"); v == "true" {
		t := true
		isIngredient = &t
	} else if v == "false" {
		f := false
		isIngredient = &f
	}

	req := &dto.GetInventoryTreeProductsRequest{
		Page:         page,
		PerPage:      perPage,
		Search:       search,
		WarehouseID:  warehouseID,
		IsIngredient: isIngredient,
	}

	result, err := h.usecase.GetTreeProducts(c.Request.Context(), req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil, nil)
		return
	}

	response.SuccessResponse(c, result, nil) // Meta is inside result
}

func (h *InventoryHandler) GetTreeBatches(c *gin.Context) {
	warehouseID := c.Query("warehouse_id")
	productID := c.Query("product_id")

	if warehouseID == "" || productID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "warehouse_id and product_id are required", nil, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	req := &dto.GetInventoryTreeBatchesRequest{
		WarehouseID: warehouseID,
		ProductID:   productID,
		Page:        page,
		PerPage:     perPage,
	}

	result, err := h.usecase.GetTreeBatches(c.Request.Context(), req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

func (h *InventoryHandler) GetInventoryMetrics(c *gin.Context) {
	result, err := h.usecase.GetInventoryMetrics(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, result, nil)
}

func (h *InventoryHandler) GetProductLedgers(c *gin.Context) {
	productID := strings.TrimSpace(c.Param("product_id"))
	if productID == "" {
		productID = strings.TrimSpace(c.Param("id"))
	}
	if _, err := uuid.Parse(productID); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "product_id must be a valid UUID", map[string]interface{}{"product_id": "invalid UUID"})
		return
	}

	page := 1
	if rawPage := strings.TrimSpace(c.Query("page")); rawPage != "" {
		parsedPage, err := strconv.Atoi(rawPage)
		if err != nil || parsedPage <= 0 {
			response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "page must be a positive integer", map[string]interface{}{"page": rawPage})
			return
		}
		page = parsedPage
	}

	limit := 20
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "limit must be a positive integer", map[string]interface{}{"limit": rawLimit})
			return
		}
		limit = parsedLimit
	}

	req := &dto.GetProductStockLedgersRequest{
		Page:            page,
		Limit:           limit,
		TransactionType: strings.TrimSpace(c.Query("transaction_type")),
		DateFrom:        strings.TrimSpace(c.Query("date_from")),
		DateTo:          strings.TrimSpace(c.Query("date_to")),
	}

	result, err := h.usecase.GetProductLedgers(c.Request.Context(), productID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInventoryInput) {
			response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
			return
		}

		response.StandardErrorResponse(c, http.StatusInternalServerError, response.ErrCodeInternalServerError, "failed to fetch stock ledgers", map[string]interface{}{"cause": err.Error()})
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetStockByProduct is an alias endpoint for inventory stock detail by product.
// It currently returns the same ledger payload used by /inventory/products/:product_id/ledgers.
func (h *InventoryHandler) GetStockByProduct(c *gin.Context) {
	h.GetProductLedgers(c)
}
