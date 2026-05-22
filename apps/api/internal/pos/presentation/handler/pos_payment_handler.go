package handler

import (
	"errors"
	"strings"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gin-gonic/gin"
)

// POSPaymentHandler handles POS payment processing
type POSPaymentHandler struct {
	uc usecase.POSPaymentUsecase
}

// NewPOSPaymentHandler creates the handler
func NewPOSPaymentHandler(uc usecase.POSPaymentUsecase) *POSPaymentHandler {
	return &POSPaymentHandler{uc: uc}
}

func getOrderIDParam(c *gin.Context) string {
	if orderID := c.Param("orderID"); orderID != "" {
		return orderID
	}
	return c.Param("id")
}

// ProcessCash processes an immediate cash or card payment
func (h *POSPaymentHandler) ProcessCash(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	orderID := getOrderIDParam(c)
	var req dto.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	payment, err := h.uc.ProcessCash(c.Request.Context(), orderID, &req, uc.userID)
	if err != nil {
		handlePOSPaymentError(c, err)
		return
	}
	response.SuccessResponse(c, payment, nil)
}

// InitiateDigitalPayment creates a Xendit invoice for the order and returns the payment details
func (h *POSPaymentHandler) InitiateDigitalPayment(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	orderID := getOrderIDParam(c)
	var req dto.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	payment, err := h.uc.InitiateDigitalPayment(c.Request.Context(), orderID, &req, uc.userID, uc.companyID)
	if err != nil {
		handlePOSPaymentError(c, err)
		return
	}
	response.SuccessResponse(c, payment, nil)
}

// CancelPendingPayment cancels a pending digital payment and reopens cashier fallback flow.
func (h *POSPaymentHandler) CancelPendingPayment(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	orderID := getOrderIDParam(c)
	paymentID := c.Param("paymentID")
	if paymentID == "" {
		paymentID = c.Param("paymentId")
	}

	var req dto.CancelPendingPaymentRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			coreErrors.HandleValidationError(c, err)
			return
		}
	}

	if err := h.uc.CancelPendingPayment(c.Request.Context(), orderID, paymentID, uc.userID, req.Reason); err != nil {
		handlePOSPaymentError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{"cancelled": true}, nil)
}

// GetByOrder returns all payments for an order
func (h *POSPaymentHandler) GetByOrder(c *gin.Context) {
	payments, err := h.uc.GetByOrderID(c.Request.Context(), getOrderIDParam(c))
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, payments, nil)
}

// XenditWebhook handles server-to-server invoice callbacks from Xendit.
// This endpoint must NOT require authentication — Xendit calls it directly.
// The webhook token header is verified inside the handler.
func (h *POSPaymentHandler) XenditWebhook(c *gin.Context) {
	var payload dto.XenditWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	if err := h.uc.ConfirmXenditWebhook(c.Request.Context(), &payload); err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, gin.H{"status": "ok"}, nil)
}

func handlePOSPaymentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrPOSOrderNotFound):
		coreErrors.NotFoundResponse(c, "pos_order", "")
	case errors.Is(err, usecase.ErrPOSOrderAlreadyPaid):
		coreErrors.ErrorResponse(c, "POS_ORDER_ALREADY_PAID", nil, nil)
	case errors.Is(err, usecase.ErrPOSPaymentNotFound):
		coreErrors.NotFoundResponse(c, "pos_payment", "")
	case errors.Is(err, usecase.ErrPOSInvalidPayment):
		coreErrors.ErrorResponse(c, "POS_INSUFFICIENT_PAYMENT", nil, nil)
	case strings.HasPrefix(err.Error(), "XENDIT_NOT_CONNECTED:"):
		reason := "Payment gateway is not available"
		coreErrors.ErrorResponse(c, "XENDIT_NOT_CONNECTED", map[string]interface{}{"reason": reason}, nil)
	case strings.HasPrefix(err.Error(), "XENDIT_INVOICE_FAILED:"):
		reason := "Failed to create digital payment invoice"
		coreErrors.ErrorResponse(c, "XENDIT_INVOICE_FAILED", map[string]interface{}{"reason": reason}, nil)
	default:
		coreErrors.ErrorResponse(c, "INTERNAL_SERVER_ERROR", map[string]interface{}{"reason": err.Error()}, nil)
	}
}
