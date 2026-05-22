package handler

import (
	"errors"
	"net/http"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gin-gonic/gin"
)

// PublicPOSHandler handles unauthenticated customer self-order endpoints.
type PublicPOSHandler struct {
	uc usecase.PublicPOSUsecase
}

// NewPublicPOSHandler creates a PublicPOSHandler.
func NewPublicPOSHandler(uc usecase.PublicPOSUsecase) *PublicPOSHandler {
	return &PublicPOSHandler{uc: uc}
}

// GetTableInfo resolves a QR token and returns the menu + outlet info.
func (h *PublicPOSHandler) GetTableInfo(c *gin.Context) {
	token := c.Param("token")
	info, err := h.uc.GetTableInfo(c.Request.Context(), token)
	if err != nil {
		handlePublicPOSError(c, err)
		return
	}
	response.SuccessResponse(c, info, nil)
}

// CreateCustomerOrder places a new self-order for the customer.
func (h *PublicPOSHandler) CreateCustomerOrder(c *gin.Context) {
	token := c.Param("token")

	var req dto.CreateCustomerOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	order, err := h.uc.CreateCustomerOrder(c.Request.Context(), token, &req)
	if err != nil {
		handlePublicPOSError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": order})
}

// GetOrderStatus returns the current status of a customer order.
func (h *PublicPOSHandler) GetOrderStatus(c *gin.Context) {
	token := c.Param("token")
	orderID := c.Param("orderId")

	order, err := h.uc.GetOrderStatus(c.Request.Context(), token, orderID)
	if err != nil {
		handlePublicPOSError(c, err)
		return
	}
	response.SuccessResponse(c, order, nil)
}

// InitiateDigitalPayment creates a Xendit invoice for a customer order.
func (h *PublicPOSHandler) InitiateDigitalPayment(c *gin.Context) {
	token := c.Param("token")
	orderID := c.Param("orderId")

	var req dto.InitiateCustomerPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	result, err := h.uc.InitiateDigitalPayment(c.Request.Context(), token, orderID, &req)
	if err != nil {
		handlePublicPOSError(c, err)
		return
	}
	response.SuccessResponse(c, result, nil)
}

// MarkPayAtCashier acknowledges that the customer will pay at the counter.
func (h *PublicPOSHandler) MarkPayAtCashier(c *gin.Context) {
	token := c.Param("token")
	orderID := c.Param("orderId")

	order, err := h.uc.MarkPayAtCashier(c.Request.Context(), token, orderID)
	if err != nil {
		handlePublicPOSError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{
		"acknowledged": true,
		"order":        order,
	}, nil)
}

// CancelCustomerOrder allows a customer to cancel their own un-processed order.
func (h *PublicPOSHandler) CancelCustomerOrder(c *gin.Context) {
	token := c.Param("token")
	orderID := c.Param("orderId")

	order, err := h.uc.CancelCustomerOrder(c.Request.Context(), token, orderID)
	if err != nil {
		handlePublicPOSError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{
		"cancelled": true,
		"order":     order,
	}, nil)
}

// CancelDigitalPayment cancels a pending digital invoice and falls back to cashier payment mode.
func (h *PublicPOSHandler) CancelDigitalPayment(c *gin.Context) {
	token := c.Param("token")
	orderID := c.Param("orderId")
	paymentID := c.Param("paymentId")

	order, err := h.uc.CancelDigitalPayment(c.Request.Context(), token, orderID, paymentID)
	if err != nil {
		handlePublicPOSError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{
		"cancelled": true,
		"order":     order,
	}, nil)
}

// handlePublicPOSError maps public-facing errors to HTTP responses.
func handlePublicPOSError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrPublicTokenInvalid):
		coreErrors.NotFoundResponse(c, "qr_token", "")
	case errors.Is(err, usecase.ErrPOSOrderNotFound):
		coreErrors.NotFoundResponse(c, "pos_order", "")
	case errors.Is(err, usecase.ErrPOSPaymentNotFound):
		coreErrors.NotFoundResponse(c, "pos_payment", "")
	case errors.Is(err, usecase.ErrPOSOutletForbidden):
		coreErrors.NotFoundResponse(c, "pos_order", "")
	case errors.Is(err, usecase.ErrPOSOrderAlreadyPaid):
		coreErrors.ErrorResponse(c, "POS_ORDER_ALREADY_PAID", nil, nil)
	case errors.Is(err, usecase.ErrPOSOrderCannotModify):
		coreErrors.ErrorResponse(c, "POS_ORDER_CANNOT_MODIFY", nil, nil)
	case errors.Is(err, usecase.ErrPOSProductNotAvailable):
		coreErrors.ErrorResponse(c, "POS_PRODUCT_NOT_AVAILABLE", nil, nil)
	case errors.Is(err, usecase.ErrPOSInsufficientStock):
		coreErrors.ErrorResponse(c, "INSUFFICIENT_STOCK", nil, nil)
	default:
		coreErrors.InternalServerErrorResponse(c, "")
	}
}
